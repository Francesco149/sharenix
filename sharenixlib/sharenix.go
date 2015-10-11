/*
   Copyright 2014 Franc[e]sco (lolisamurai@tfwno.gf)
   This file is part of sharenix.
   sharenix is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   sharenix is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with sharenix. If not, see <http://www.gnu.org/licenses/>.
*/

// Package sharenixlib contains the core functionalities of sharenix
// it can be used to implement custom front-ends for sharenix.
// NOTE: to compile this, you need >=gtk/gdk-3.10 and >=go-1.3.1
// You will also need my modified fork of gotk3: github.com/Francesco149/gotk3
// (go get it then rename it to github.com/conformal/gotk3 so that it can be
// properly imported)
package sharenixlib

import (
	"bytes"
	"errors"
	"flag"
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/gtk"
	"image"
	"image/png"
	"net/http"
	"os"
	"os/exec"
)

const ShareNixDebug = true
const ShareNixVersion = "ShareNix 0.1.1a"

// UploadFile uploads a file
// cfg: the ShareNix config
// sitecfg: the target site config
// path: file path
// silent: disables all console output except errors
// notification: displays a gtk notification for the upload
func UploadFile(cfg *Config, sitecfg *SiteConfig, path string,
	silent, notification bool) (*http.Response, string, error) {

	sitecfg, err := cfg.HandleFileType(sitecfg, path, silent)
	if err != nil {
		return nil, "", err
	}

	Println(silent, "Uploading file to", sitecfg.Name)
	// TODO: notification
	return SendFilePostRequest(sitecfg.RequestURL, sitecfg.FileFormName,
		path, sitecfg.Arguments)
}

func MakeArchiveDir() error {
	// create archive dir
	direxists, err := FileExists("./archive/")
	if err != nil {
		return err
	}
	if !direxists {
		err = os.Mkdir("./archive/", os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

// UploadFullScreen captures a full screen screenshot,
// saves it in the archive and uploads it
// cfg: the ShareNix config
// sitecfg: the target site config
// silent: disables all console output except errors
// notification: displays a gtk notification for the upload
func UploadFullScreen(cfg *Config, sitecfg *SiteConfig,
	silent, notification bool) (*http.Response, string, error) {

	Println(silent, "Taking screenshot...")
	// capture screen
	var img *image.RGBA
	img, err := CaptureScreen()
	if err != nil {
		return nil, "", err
	}

	err = MakeArchiveDir()
	if err != nil {
		return nil, "", err
	}

	// save to archive
	afilepath := GenerateArchivedFilename("png")
	tmpfile, err := os.Create(afilepath)
	if err != nil {
		return nil, "", err
	}
	err = png.Encode(tmpfile, img)
	tmpfile.Close()

	// TODO: notification

	// upload
	Println(silent, "Uploading to", sitecfg.Name)
	return SendFilePostRequest(sitecfg.RequestURL, sitecfg.FileFormName,
		afilepath, sitecfg.Arguments)
}

// UploadClipboard grabs an image or a file from the clipboard,
// saves it in the archive and uploads it
// cfg: the ShareNix config
// sitecfg: the target site config
// silent: disables all console output except errors
// notification: displays a gtk notification for the upload
func UploadClipboard(cfg *Config, sitecfg *SiteConfig,
	silent, notification bool) (res *http.Response, filename string, err error) {

	clipboard, err := GetClipboard()
	if err != nil {
		return
	}

	DebugPrintln("Looking for copied files...")
	selectiondata, err := clipboard.WaitForContents(
		gdk.GdkAtomIntern("x-special/gnome-copied-files", false))

	if err != nil {
		err = MakeArchiveDir()
		if err != nil {
			return
		}

		var pixbuf *gdk.Pixbuf
		// assume that the user has copied an image
		DebugPrintln("Looking for copied raw images...")
		pixbuf, err = clipboard.WaitForImage()
		if err != nil {
			return
		}

		DebugPrintln("Colorspace:", int(pixbuf.GetColorspace()), "Channels:", pixbuf.GetNChannels(),
			"Has Alpha:", pixbuf.GetHasAlpha(), "BPS:", pixbuf.GetBitsPerSample(), "Width:",
			pixbuf.GetWidth(), "Height:", pixbuf.GetHeight(), "Rowstride:", pixbuf.GetRowstride(),
			"Byte length:", pixbuf.GetByteLength())

		// encode png to archive and upload
		// TODO: check if the image format can change and how to deal with it
		pixels := pixbuf.GetPixels()
		pic := &image.RGBA{pixels, 4 * pixbuf.GetWidth(), image.Rect(0, 0, pixbuf.GetWidth(), pixbuf.GetHeight())}

		afilepath := GenerateArchivedFilename("png")
		var tmpfile *os.File
		tmpfile, err = os.Create(afilepath)
		if err != nil {
			return
		}
		err = png.Encode(tmpfile, pic)
		tmpfile.Close()

		return UploadFile(cfg, sitecfg, afilepath, silent, notification)
	}

	// upload first copied file with UploadFile
	// TODO: batch upload all copied files
	selectionstr := string(selectiondata.GetData())
	DebugPrintln(selectionstr)
	urilist := ParseUriList(selectionstr)
	if len(urilist) < 1 {
		err = errors.New("Could not grab copied file list")
		return
	}

	return UploadFile(cfg, sitecfg, urilist[0].Path, silent, notification)
}

/*
	ShareNix uploads a file with the given options
	cfg: ShareNix config
	mode:
		f/file: upload file
		fs/fullscreen: screenshot entire screen and upload
		s/section: select screen region and upload
		c/clipboard: upload clipboard contents
		r/record: record screen region and upload
		u/url: shorten url
	site: name of the target site
	silent: disables all console output except errors if enabled
	notification: displays a gtk notification for uploads if enabled
	open: automatically opens the uploaded file in the default browser
	copyurl: stores the url in the clipboard after uploading
*/
func ShareNix(cfg *Config, mode, site string, silent,
	notification, open, copyurl bool) (
	url, thumburl, deleteurl string, err error) {

	var sitecfg *SiteConfig
	var res *http.Response
	var filename string

	gtk.Init(nil)

	// initial upload mode check
	sitecfg, err = cfg.Parse(mode, site, silent)
	if err != nil {
		return
	}

	// call the correct upload handler
	switch mode {
	case "f", "file":
		if len(flag.Args()) != 1 {
			err = errors.New("No file provided")
			return
		}
		res, filename, err = UploadFile(cfg, sitecfg, flag.Args()[0], silent, notification)

	case "fs", "fullscreen":
		res, filename, err = UploadFullScreen(cfg, sitecfg, silent, notification)

	case "c", "clipboard":
		res, filename, err = UploadClipboard(cfg, sitecfg, silent, notification)

	case "s", "u", "section", "url":
		err = &NotImplementedError{}
	}

	if err != nil {
		return
	}

	switch sitecfg.ResponseType {
	case "RedirectionURL":
		DebugPrintln("Getting redirection url...")
		url = res.Request.URL.String()
	default:
		// parse response
		DebugPrintln("Parsing response...")
		rbody := &bytes.Buffer{}
		_, err = rbody.ReadFrom(res.Body)
		if err != nil {
			return
		}

		// parse all regular expressions
		var results [][]string
		results, err = ParseRegexList(rbody.String(), sitecfg.RegexList)
		if err != nil {
			return
		}

		// replace regular expression tags in urls
		url = ParseUrl(sitecfg.URL, results)
		thumburl = ParseUrl(sitecfg.ThumbnailURL, results)
		deleteurl = ParseUrl(sitecfg.DeletionURL, results)

		// empty url = take entire response as url
		if len(url) == 0 {
			url = rbody.String()
		}
	}

	AppendToHistory(url, thumburl, deleteurl, filename)

	if copyurl {
		DebugPrintln("Copying url to clipboard...")
		SetClipboardText(url)
	}

	if open {
		err = exec.Command("xdg-open", url).Run()
	}

	return
}
