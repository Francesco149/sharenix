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
	"fmt"
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"
	"github.com/kardianos/osext"
	"github.com/mvdan/xurls"
	"image"
	"image/png"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

const (
	ShareNixDebug   = true
	ShareNixVersion = "ShareNix 0.2.1a"
)

const (
	notifTime    = time.Second * 30
	infiniteTime = time.Duration(9000000000000000000)
)

// UploadFile uploads a file
// cfg: the ShareNix config
// sitecfg: the target site config
// path: file path
// silent: disables all console output except errors
// notif: if true, a notification will display during and after the request
func UploadFile(cfg *Config, sitecfg *SiteConfig, path string,
	silent, notif bool) (res *http.Response, filename string, err error) {

	sitecfg, err = cfg.HandleFileType(sitecfg, path, silent)
	if err != nil {
		return
	}

	Println(silent, "Uploading file to", sitecfg.Name)

	doThings := func() (*http.Response, string, error) {
		return SendFilePostRequest(sitecfg.RequestURL,
			sitecfg.FileFormName, path, sitecfg.Arguments)
	}

	if notif {
		onload := func(w *gtk.Window) {
			res, filename, err = doThings()
			glib.IdleAdd(w.Destroy)
			DebugPrintln("Goroutine is exiting")
		}
		err = Notifyf(infiniteTime,
			onload, "Uploading %s to %s...", path, sitecfg.Name)
		return
	}
	return doThings()
}

// ShortenUrl shortens an url
// cfg: the ShareNix config
// sitecfg: the target site config
// url: url to be shortened
// silent: disables all console output except errors
// notif: if true, a notification will display during and after the request
func ShortenUrl(cfg *Config, sitecfg *SiteConfig, url string,
	silent, notif bool) (res *http.Response, err error) {

	for i := range sitecfg.Arguments {
		sitecfg.Arguments[i] = strings.Replace(
			sitecfg.Arguments[i], "$input$", url, -1)
	}

	Println(silent, "Shortening with", sitecfg.Name)

	doThings := func() (*http.Response, error) {
		switch sitecfg.RequestType {
		case "GET":
			return SendGetRequest(sitecfg.RequestURL, sitecfg.Arguments)
		case "POST":
			return SendPostRequest(sitecfg.RequestURL, sitecfg.Arguments)
		default:
			return nil, errors.New("Unknown RequestType")
		}
	}

	if notif {
		onload := func(w *gtk.Window) {
			res, err = doThings()
			glib.IdleAdd(w.Destroy)
			DebugPrintln("Goroutine is exiting")
		}
		err = Notifyf(infiniteTime,
			onload, "Shortening %s with %s...", url, sitecfg.Name)
		return
	}

	return doThings()
}

// GetArchiveDir returns the absolute path to the archive directory.
func GetArchiveDir() (archiveDir string, err error) {
	exeFolder, err := osext.ExecutableFolder()
	if err != nil {
		return
	}
	archiveDir = path.Join(exeFolder, "/archive/")
	return
}

// MakeArchiveDir creates the archive directory if it doesn't exist already.
func MakeArchiveDir() error {
	achiveDir, err := GetArchiveDir()
	if err != nil {
		return err
	}

	// create archive dir
	direxists, err := FileExists(achiveDir)
	if err != nil {
		return err
	}
	if !direxists {
		err = os.Mkdir(achiveDir, os.ModePerm)
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
// notif: if true, a notification will display during and after the request
func UploadFullScreen(cfg *Config, sitecfg *SiteConfig, silent, notif bool) (
	res *http.Response, file string, err error) {

	Println(silent, "Taking screenshot...")

	// capture screen
	img, err := CaptureScreen()
	if err != nil {
		return
	}

	// save to archive
	afilepath, err := GenerateArchivedFilename("png")
	if err != nil {
		return
	}

	tmpfile, err := os.Create(afilepath)
	if err != nil {
		return
	}

	err = png.Encode(tmpfile, img)
	tmpfile.Close()

	// upload
	Println(silent, "Uploading to", sitecfg.Name)

	doThings := func() (*http.Response, string, error) {
		return SendFilePostRequest(sitecfg.RequestURL, sitecfg.FileFormName,
			afilepath, sitecfg.Arguments)
	}

	if notif {
		onload := func(w *gtk.Window) {
			res, file, err = doThings()
			glib.IdleAdd(w.Destroy)
			DebugPrintln("Goroutine is exiting")
		}
		err = Notifyf(infiniteTime,
			onload, "Uploading screenshot to %s...", sitecfg.Name)
		return
	}
	return doThings()
}

// Creates and opens an archive file with the given extension.
func CreateArchiveFile(extension string) (
	tmpfile *os.File, path string, err error) {

	path, err = GenerateArchivedFilename(extension)
	if err != nil {
		return
	}
	tmpfile, err = os.Create(path)
	return
}

// UploadClipboard grabs an image or a file from the clipboard,
// saves it in the archive and uploads it
// cfg: the ShareNix config
// sitecfg: the target site config
// silent: disables all console output except errors
// notif: if true, a notification will display during and after the request
func UploadClipboard(cfg *Config, sitecfg *SiteConfig, silent, notif bool) (
	res *http.Response, filename string, err error) {

	clipboard, err := GetClipboard()
	if err != nil {
		return
	}

	// URI list (copied files)
	DebugPrintln("Looking for URI list...")
	selectiondata, err := clipboard.WaitForContents(
		gdk.GdkAtomIntern("x-special/gnome-copied-files", false))

	if err == nil {
		selectionstr := string(selectiondata.GetData())
		DebugPrintln(selectionstr)

		// upload first copied file with UploadFile
		// TODO: batch upload all copied files
		DebugPrintln("Trying to parse URI list...")
		urilist := ParseUriList(selectionstr)
		if len(urilist) > 0 {
			return UploadFile(
				cfg, sitecfg, urilist[0].Path, silent, notif)
		}
		DebugPrintln("URI list is empty")
	}

	// Plain text (shorten url or upload as text file)
	DebugPrintln("Looking for plain text...")
	selectionstr, err := clipboard.WaitForText()
	if err == nil && len(selectionstr) > 0 {
		DebugPrintln(selectionstr)

		DebugPrintln("Trying to parse as URL...")
		if xurls.Strict.MatchString(selectionstr) {
			sitecfg = cfg.GetServiceByName(cfg.DefaultUrlShortener)
			res, err = ShortenUrl(cfg, sitecfg,
				selectionstr, silent, notif)
			filename = selectionstr
			return
		}

		DebugPrintln("Trying to upload as plain text...")
		var afilepath string
		var tmpfile *os.File
		tmpfile, afilepath, err = CreateArchiveFile("txt")
		if err != nil {
			return
		}
		_, err = tmpfile.WriteString(selectionstr)
		tmpfile.Close()

		return UploadFile(cfg, sitecfg, afilepath, silent, notif)
	}

	// Raw image (copied from an image editor or from the browser)
	var pixbuf *gdk.Pixbuf
	DebugPrintln("Looking for copied raw images...")
	pixbuf, err = clipboard.WaitForImage()
	if err == nil {
		DebugPrintln("Colorspace:", int(pixbuf.GetColorspace()), "Channels:",
			pixbuf.GetNChannels(), "Has Alpha:", pixbuf.GetHasAlpha(), "BPS:",
			pixbuf.GetBitsPerSample(), "Width:", pixbuf.GetWidth(), "Height:",
			pixbuf.GetHeight(), "Rowstride:", pixbuf.GetRowstride(),
			"Byte length:", pixbuf.GetByteLength())

		// encode png to archive and upload
		pixels := pixbuf.GetPixels()
		pic := &image.RGBA{pixels, 4 * pixbuf.GetWidth(), image.Rect(0, 0,
			pixbuf.GetWidth(), pixbuf.GetHeight())}

		var afilepath string
		var tmpfile *os.File
		tmpfile, afilepath, err = CreateArchiveFile("png")
		if err != nil {
			return
		}

		err = png.Encode(tmpfile, pic)
		if err != nil {
			return
		}

		tmpfile.Close()

		return UploadFile(cfg, sitecfg, afilepath, silent, notif)
	}

	err = errors.New("Could not find any supported data in the clipboard")
	return
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
	notification: displays a gtk notification if enabled
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

	err = MakeArchiveDir()
	if err != nil {
		return
	}

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
		res, filename, err = UploadFile(cfg, sitecfg,
			flag.Args()[0], silent, notification)

	case "fs", "fullscreen":
		res, filename, err = UploadFullScreen(
			cfg, sitecfg, silent, notification)

	case "c", "clipboard":
		res, filename, err = UploadClipboard(cfg, sitecfg, silent, notification)

	case "u", "url":
		if len(flag.Args()) != 1 {
			err = errors.New("No url provided")
			return
		}
		res, err = ShortenUrl(cfg, sitecfg,
			flag.Args()[0], silent, notification)
		filename = flag.Args()[0]

	case "s", "section":
		err = &NotImplementedError{}
	}

	if err != nil {
		return
	}

	switch sitecfg.ResponseType {
	case "RedirectionURL":
		DebugPrintln("Getting redirection url...")
		url = res.Request.URL.String()
	case "Text":
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
	default:
		url = "Unrecognized ResponseType"
	}

	if xurls.Strict.MatchString(url) {
		AppendToHistory(url, thumburl, deleteurl, filename)
	} else {
		err = errors.New(fmt.Sprintf("Request failed: %s", url))
	}

	if copyurl {
		DebugPrintln("Copying url to clipboard...")
		SetClipboardText(url)
	}

	if open && err == nil {
		err = exec.Command("xdg-open", url).Run()
	}

	if notification {
		if err != nil {
			Notifyf(notifTime, nil, "%v", err)
		} else {
			Notifyf(notifTime, nil, `<a href="%s">%s</a>`, url, url)
		}
	}

	return
}
