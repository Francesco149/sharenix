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
	"github.com/BurntSushi/xgb"
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"
	"github.com/kardianos/osext"
	"github.com/mvdan/xurls"
	"html"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	ShareNixDebug   = true
	ShareNixVersion = "ShareNix 0.3.4a"
)

const (
	notifTime    = time.Second * 30
	infiniteTime = time.Duration(9000000000000000000)
)

// -----------------------------------------------------------------------------
// !! WARNING: Ghetto code ahead !!

var server *httptest.Server

// http://keighl.com/post/mocking-http-responses-in-golang/
// just cause I'm too lazy to add a special case to the output parsing
// for plugins
func fakeResponseStart(code int, body string) (*httptest.Server, *http.Client) {
	server = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
			w.Header().Set("Content-Type", "plain/text")
			fmt.Fprintln(w, body)
		}),
	)

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	client := &http.Client{Transport: transport}
	return server, client
}

func fakeResponseEnd() {
	if server != nil {
		server.Close()
		server = nil
	}
}

// -----------------------------------------------------------------------------

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

	for i := range sitecfg.Arguments {
		sitecfg.Arguments[i] = strings.Replace(
			sitecfg.Arguments[i], "$input$", path, -1)
	}

	Println(silent, "Uploading file to", sitecfg.Name)

	doThings := func() (res *http.Response, filename string, err error) {
		if sitecfg.RequestType == "PLUGIN" {
			var output string
			output, err = RunPlugin(sitecfg.RequestURL, sitecfg.Arguments)
			DebugPrintln("RunPlugin returned", len(output), "bytes:",
				output, "with error", err)
			if err != nil {
				return
			}

			server, client := fakeResponseStart(200, output)
			res, err = client.Get(server.URL + "/")
			if err != nil {
				return
			}

			filename = filepath.Base(path)
			return
		}
		return SendFilePostRequest(sitecfg.RequestURL,
			sitecfg.FileFormName, path, sitecfg.Arguments)
	}

	if notif {
		onload := func(w *gtk.Window) {
			res, filename, err = doThings()
			glib.IdleAdd(w.Destroy)
			DebugPrintln("Goroutine is exiting")
		}
		notiferr := Notifyf(infiniteTime,
			onload, "Uploading %s to %s...", path, sitecfg.Name)
		if notiferr != nil {
			err = notiferr
		}
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
		case "PLUGIN":
			output, err := RunPlugin(sitecfg.RequestURL, sitecfg.Arguments)
			if err != nil {
				return nil, err
			}
			server, client := fakeResponseStart(200, output)
			return client.Get(server.URL + "/")
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
	X, err := xgb.NewConn()
	if err != nil {
		return
	}
	defer X.Close()

	// capture screen
	img, err := CaptureScreen(X)
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

	// TODO: avoid repeating this loop in every upload function and move
	// it to its own func
	for i := range sitecfg.Arguments {
		sitecfg.Arguments[i] = strings.Replace(
			sitecfg.Arguments[i], "$input$", afilepath, -1)
	}

	// upload
	Println(silent, "Uploading to", sitecfg.Name)

	// TODO: make a more generic version of this switch to avoid repeating
	// this code over and over
	doThings := func() (*http.Response, string, error) {
		switch sitecfg.RequestType {
		case "GET":
			return nil, "", errors.New("GET file upload is not supported.")
		case "POST":
			return SendFilePostRequest(sitecfg.RequestURL, sitecfg.FileFormName,
				afilepath, sitecfg.Arguments)
		case "PLUGIN":
			output, err := RunPlugin(
				sitecfg.RequestURL, sitecfg.Arguments)
			if err != nil {
				return nil, "", err
			}
			server, client := fakeResponseStart(200, output)
			res, err := client.Get(server.URL + "/")
			return res, filepath.Base(afilepath), err
		default:
			return nil, "", errors.New("Unknown RequestType")
		}
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

	defaultConfig := sitecfg.Name == cfg.DefaultFileUploader

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
			if defaultConfig {
				sitecfg = cfg.GetServiceByName(cfg.DefaultUrlShortener)
			}
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
	notification: displays a gtk notification if enabled. note that dimissing
	              this notification will force quit the process and the function
	              will never return.
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

	if res == nil {
		err = fmt.Errorf("Request failed, but I don't know why!")
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

	fakeResponseEnd()

	url = strings.TrimSuffix(url, "\n")
	matchedUrls := xurls.Strict.FindAllString(url, -1)
	urli := xurls.Strict.FindIndex([]byte(url))
	if len(matchedUrls) == 1 && urli[0] == 0 &&
		len(url) == len(matchedUrls[0]) {
		// the result must only contain an url with no extra stuff to be
		// considered a valid response
		AppendToHistory(url, thumburl, deleteurl, filename)
	} else {
		err = fmt.Errorf("Request failed: %s", url)
	}

	if copyurl {
		DebugPrintln("Copying url to clipboard...")
		SetClipboardText(url)
	}

	if open && err == nil {
		err = exec.Command("xdg-open", url).Run()
	}

	// display results
	Println(silent, "URL:", url)
	if len(thumburl) > 0 {
		Println(silent, "Thumbnail URL:", thumburl)
	}
	if len(deleteurl) > 0 {
		Println(silent, "Deletion URL:", deleteurl)
	}

	if notification {
		if err != nil {
			Notifyf(notifTime, nil, html.EscapeString(err.Error()))
		} else {
			Notifyf(notifTime, nil, `<a href="%s">%s</a>`, url, url)
		}
	}

	return
}
