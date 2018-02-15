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
// NOTE: to compile this, you need gtk 2.0 and >=go-1.3.1
package sharenixlib

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"html"
	"image/png"
	"io"
	"mvdan.cc/xurls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

const (
	ShareNixDebug   = true
	ShareNixVersion = "ShareNix 0.9.5a"
)

const (
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

// ReplaceKeywords replaces various keywords in a string
// $Y$: local year padded to 4 digits
// $M$: local month padded to 2 digits
// $D$: local day padded to 2 digits
// $h$: local hours padded to 2 digits
// $m$: local minutes padded to 2 digits
// $s$: local seconds padded to 2 digits
// $n$: local nanoseconds
func ReplaceKeywords(str string) string {
	t := time.Now()
	replacements := map[string]func() string {
		"$Y$": func() string { return fmt.Sprintf("%04d", t.Year()) },
		"$M$": func() string { return fmt.Sprintf("%02d", t.Month()) },
		"$D$": func() string { return fmt.Sprintf("%02d", t.Day()) },
		"$h$": func() string { return fmt.Sprintf("%02d", t.Hour()) },
		"$m$": func() string { return fmt.Sprintf("%02d", t.Minute()) },
		"$s$": func() string { return fmt.Sprintf("%02d", t.Second()) },
		"$n$": func() string { return fmt.Sprintf("%d", t.Nanosecond()) },
	}

	for key, formatter := range replacements {
		str = strings.Replace(str, key, formatter(), -1)
	}

	return str
}

// UploadFile uploads a file
// cfg: the ShareNix config
// sitecfg: the target site config
// path: file path
// silent: disables all console output except errors
// notif: if true, a notification will display during and after the request
func UploadFile(cfg *Config, sitecfg *SiteConfig, path string,
	silent, notif bool) (
	res *http.Response, filename string, newsitecfg *SiteConfig, err error) {

	// this hack fixes "invalid argument" when there's leftover zero bytes
	// in the paths
	// TODO: use better gtk bindings that don't leave nil bytes
	path = string(bytes.TrimRight([]byte(path), "\000"))

	newsitecfg = sitecfg
	sitecfg, err = cfg.HandleFileType(sitecfg, path, silent)
	if err != nil {
		return
	}

	basepath := filepath.Base(path)

	for i := range sitecfg.Arguments {
		sitecfg.Arguments[i] = strings.Replace(sitecfg.Arguments[i],
			"$input$", basepath, -1)
		sitecfg.Arguments[i] = ReplaceKeywords(sitecfg.Arguments[i])
	}

	sitecfg.RequestURL = strings.Replace(sitecfg.RequestURL, "$input$",
		basepath, -1)
	sitecfg.RequestURL = ReplaceKeywords(sitecfg.RequestURL)

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
		return SendRequest(sitecfg.RequestType, sitecfg.RequestURL,
			sitecfg.FileFormName, path, sitecfg.Arguments,
			sitecfg.Headers, sitecfg.Username, sitecfg.Password)
	}

	newsitecfg = sitecfg

	if notif {
		onload := func(w *gtk.Window) {
			res, filename, err = doThings()
			glib.IdleAdd(w.Destroy)
			DebugPrintln("Goroutine is exiting")
		}
		notiferr := Notifyf(cfg.XineramaHead, infiniteTime,
			onload, "Uploading %s to %s...", path, sitecfg.Name)
		if notiferr != nil {
			err = notiferr
		}
		return
	}

	res, filename, err = doThings()
	return
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
		sitecfg.Arguments[i] = ReplaceKeywords(sitecfg.Arguments[i])
	}

	sitecfg.RequestURL = strings.Replace(sitecfg.RequestURL, "$input$",
		url, -1)
	sitecfg.RequestURL = ReplaceKeywords(sitecfg.RequestURL)

	Println(silent, "Shortening with", sitecfg.Name)

	doThings := func() (*http.Response, error) {
		switch sitecfg.RequestType {
		case "PLUGIN":
			output, err := RunPlugin(sitecfg.RequestURL, sitecfg.Arguments)
			if err != nil {
				return nil, err
			}
			server, client := fakeResponseStart(200, output)
			return client.Get(server.URL + "/")
		default:
			res, _, err = SendRequest(sitecfg.RequestType,
				sitecfg.RequestURL, sitecfg.FileFormName, "",
				sitecfg.Arguments, sitecfg.Headers, sitecfg.Username,
				sitecfg.Password)
			return res, err
		}
	}

	if notif {
		onload := func(w *gtk.Window) {
			res, err = doThings()
			glib.IdleAdd(w.Destroy)
			DebugPrintln("Goroutine is exiting")
		}
		err = Notifyf(cfg.XineramaHead, infiniteTime,
			onload, "Shortening %s with %s...", url, sitecfg.Name)
		return
	}

	return doThings()
}

// UploadFullScreen captures a full screen screenshot,
// saves it in the archive and uploads it
// cfg: the ShareNix config
// sitecfg: the target site config
// silent: disables all console output except errors
// notif: if true, a notification will display during and after the request
func UploadFullScreen(cfg *Config, sitecfg *SiteConfig, silent, notif bool) (
	res *http.Response, file string, newsitecfg *SiteConfig, err error) {

	newsitecfg = sitecfg

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
	basepath := filepath.Base(afilepath)

	for i := range sitecfg.Arguments {
		sitecfg.Arguments[i] = strings.Replace(sitecfg.Arguments[i],
			"$input$", basepath, -1)
		sitecfg.Arguments[i] = ReplaceKeywords(sitecfg.Arguments[i])
	}

	sitecfg.RequestURL = strings.Replace(sitecfg.RequestURL, "$input$",
		basepath, -1)
	sitecfg.RequestURL = ReplaceKeywords(sitecfg.RequestURL)

	// upload
	Println(silent, "Uploading to", sitecfg.Name)

	doThings := func() (*http.Response, string, error) {
		switch sitecfg.RequestType {
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
			return SendRequest(sitecfg.RequestType, sitecfg.RequestURL,
				sitecfg.FileFormName, afilepath, sitecfg.Arguments,
				sitecfg.Headers, sitecfg.Username, sitecfg.Password)
		}
	}

	if notif {
		onload := func(w *gtk.Window) {
			res, file, err = doThings()
			glib.IdleAdd(w.Destroy)
			DebugPrintln("Goroutine is exiting")
		}
		err = Notifyf(cfg.XineramaHead, infiniteTime,
			onload, "Uploading screenshot to %s...", sitecfg.Name)
		return
	}

	newsitecfg = sitecfg
	res, file, err = doThings()
	return
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

// TODO: merge these two funcs together
func ArchiveFile(path string) (err error) {
	path = string(bytes.TrimRight([]byte(path), "\000"))

	var tmpfile *os.File
	tmpfile, _, err = CreateArchiveFile(filepath.Ext(path)[1:])
	if err != nil {
		return
	}
	defer tmpfile.Close()

	src, err := os.Open(path)
	if err != nil {
		fmt.Println("???")
		return
	}
	defer src.Close()

	if _, err = io.Copy(tmpfile, src); err != nil {
		return
	}

	if err = tmpfile.Sync(); err != nil {
		return
	}

	return
}

// UploadClipboard grabs an image or a file from the clipboard,
// saves it in the archive and uploads it
// cfg: the ShareNix config
// sitecfg: the target site config
// silent: disables all console output except errors
// notif: if true, a notification will display during and after the request
func UploadClipboard(cfg *Config, sitecfg *SiteConfig, silent, notif bool) (
	res *http.Response, filename string, newsitecfg *SiteConfig, err error) {

	defaultConfig := sitecfg.Name == cfg.DefaultFileUploader
	newsitecfg = sitecfg

	clipboard := GetClipboard()

	// URI list (copied files)
	DebugPrintln("Looking for URI list...")
	selectiondata := clipboard.WaitForContents(
		gdk.AtomIntern("x-special/gnome-copied-files", false))

	// NOTE: this is supposed to be freed, but we're just not gonna care for now
	//       because we only call this once anyways

	ptr := selectiondata.GetData()

	if uintptr(ptr) != 0 {
		var bytes []byte
		hdr := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
		hdr.Data = uintptr(ptr)
		hdr.Len = selectiondata.GetLength()
		hdr.Cap = hdr.Len

		selectionstr := string(bytes)
		DebugPrintln(selectionstr)

		// upload first copied file with UploadFile
		// TODO: batch upload all copied files
		DebugPrintln("Trying to parse URI list...")
		urilist := ParseUriList(selectionstr)
		if len(urilist) > 0 {
			if err = ArchiveFile(urilist[0].Path); err != nil {
				return
			}
			// TODO: merge all archive calls into one in UploadFile

			return UploadFile(
				cfg, sitecfg, urilist[0].Path, silent, notif)
		}
		DebugPrintln("URI list is empty")
	} else {
		DebugPrintln("gtk_selection_data_get_data returned NULL")
	}

	// Plain text (shorten url or upload as text file)
	DebugPrintln("Looking for plain text...")
	selectionstr := clipboard.WaitForText()
	if len(selectionstr) > 0 {
		DebugPrintln(selectionstr)

		DebugPrintln("Trying to parse as URL...")
		if xurls.Strict().MatchString(selectionstr) {
			match := xurls.Strict().FindString(selectionstr)
			if strings.HasPrefix(selectionstr, match) {
				if defaultConfig {
					sitecfg = cfg.GetServiceByName(cfg.DefaultUrlShortener)
				}
				res, err = ShortenUrl(cfg, sitecfg,
					selectionstr, silent, notif)
				filename = selectionstr
				newsitecfg = sitecfg
				return
			}
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
	} else {
		DebugPrintln("gtk_clipboard_wait_for_text returned an empty string")
	}

	// Raw image (copied from an image editor or from the browser)
	DebugPrintln("Looking for copied raw images...")
	pixbuf := clipboard.WaitForImage()
	if uintptr(unsafe.Pointer(pixbuf.GPixbuf)) != 0 {
		DebugPrintln("Colorspace:", int(pixbuf.GetColorspace()),
			"Channels:", pixbuf.GetNChannels(),
			"Has Alpha:", pixbuf.GetHasAlpha(),
			"BPS:", pixbuf.GetBitsPerSample(),
			"Width:", pixbuf.GetWidth(),
			"Height:", pixbuf.GetHeight(),
			"Rowstride:", pixbuf.GetRowstride())

		// touch archive file
		var afilepath string
		var tmpfile *os.File
		tmpfile, afilepath, err = CreateArchiveFile("png")
		if err != nil {
			return
		}
		tmpfile.Close()

		// let gtk save it as a proper png so we don't have to
		// figure out what type of image we are dealing with
		pixbuf.Save(afilepath, "png")
		// TODO: for some reason this always returns an err which
		// prints as nil so we can't error check here :(

		return UploadFile(cfg, sitecfg, afilepath, silent, notif)
	} else {
		DebugPrintln("gtk_clipboard_wait_for_image returned NULL")
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

	// initial upload mode check
	sitecfg, err = cfg.Parse(mode, site, silent)
	if err != nil {
		return
	}

	// TODO: move all sitecfg switches here, the current method
	//       is a huge mess

	requiresgtk := false

	switch mode {
	case "c", "clipboard", "u", "url":
		requiresgtk = true
	}

	requiresgtk = requiresgtk || notification
	requiresgtk = requiresgtk || copyurl

	if requiresgtk {
		gtk.Init(nil)
	}

	// call the correct upload handler
	switch mode {
	case "f", "file":
		if len(flag.Args()) != 1 {
			err = errors.New("No file provided")
			return
		}
		if err = ArchiveFile(flag.Args()[0]); err != nil {
			return
		}
		res, filename, sitecfg, err = UploadFile(cfg, sitecfg,
			flag.Args()[0], silent, notification)

	case "fs", "fullscreen":
		res, filename, sitecfg, err = UploadFullScreen(
			cfg, sitecfg, silent, notification)

	case "c", "clipboard":
		res, filename, sitecfg, err =
			UploadClipboard(cfg, sitecfg, silent, notification)

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

		DebugPrintln(rbody)

		// parse all regular expressions
		var results [][]string
		results, err = ParseRegexList(rbody.String(), sitecfg.RegexList)
		if err != nil {
			return
		}

		// replace regular expressions and other tags in urls
		url = ParseUrl(rbody.Bytes(), sitecfg.URL, results)
		thumburl = ParseUrl(rbody.Bytes(), sitecfg.ThumbnailURL, results)
		deleteurl = ParseUrl(rbody.Bytes(), sitecfg.DeletionURL, results)

		// empty url = take entire response as url
		if len(url) == 0 {
			url = rbody.String()
		}
	default:
		url = "Unrecognized ResponseType"
	}

	fakeResponseEnd()

	url = strings.TrimSuffix(url, "\n")
	matchedUrls := xurls.Strict().FindAllString(url, -1)
	urli := xurls.Strict().FindIndex([]byte(url))
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
		if err != nil {
			DebugPrintln(err)
			err = nil
		}
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
			Notifyf(cfg.XineramaHead,
				time.Second*time.Duration(cfg.NotificationTime), nil,
				html.EscapeString(err.Error()))
		} else {
			Notifyf(cfg.XineramaHead,
				time.Second*time.Duration(cfg.NotificationTime), nil,
				`<a href="%s">%s</a>`, url, url)
		}
	}

	return
}
