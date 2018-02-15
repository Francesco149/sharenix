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

package sharenixlib

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
)

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// SniffMimeType sniffs the mime type of a binary file by reading the
// first 512 bytes
func SniffMimeType(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("SniffMimeType failed to open file: %s", err)
	}
	defer file.Close()

	// first 512 bytes are used to evaluate mime type
	first512 := make([]byte, 512)
	n, err := file.Read(first512)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(first512[:n]), nil
}

// SendRequest prepares HTTP request and sends it
//
// if fileParamName empty, no file field will be created and
// filePath is ignored
// if username is empty, no http auth header will be sent
//
// if method is GET or PUT, the parameters will be url-encoded, otherwise they
// will be fields of the multi-part form
//
// if method is PUT and filePath is set, the request body will be the contents
// of the file
func SendRequest(method, url, fileParamName, filePath string,
	extraParams map[string]string, extraHeaders map[string]string,
	username string, password string,
) (res *http.Response, filename string, err error) {

	if method == "GET" || method == "PUT" {
		var u *neturl.URL
		u, err = neturl.Parse(url)
		if err != nil {
			return
		}

		// url-encode extra params
		q := u.Query()
		for name, value := range extraParams {
			q.Set(name, value)
		}
		u.RawQuery = q.Encode()

		DebugPrintln(u)
		url = u.String()
	}

	// prepare request body buffer
	buf := &bytes.Buffer{}

	w := multipart.NewWriter(buf)

	if fileParamName != "" && method != "PUT" {
		filename = filepath.Base(filePath)

		var realmime string
		realmime, err = SniffMimeType(filePath)
		if err != nil {
			return
		}

		// try opening the file
		var file *os.File
		file, err = os.Open(filePath)
		if err != nil {
			return
		}
		defer file.Close()

		// create a multipart file header with the given param name and file
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				escapeQuotes(fileParamName), escapeQuotes(filename)))
		h.Set("Content-Type", realmime)

		var formfile io.Writer
		formfile, err = w.CreatePart(h)
		if err != nil {
			return
		}

		// write the file to the file header
		_, err = io.Copy(formfile, file)
	}

	// append extra params as form fields
	if method != "GET" && method != "PUT" {
		for param, val := range extraParams {
			err = w.WriteField(param, val)
			if err != nil {
				return
			}
		}
	}

	ctype := w.FormDataContentType()

	err = w.Close()
	if err != nil {
		return
	}

	// finally create the request
	var req *http.Request
	if method == "PUT" && filePath != "" {
		var freader *os.File
		freader, err = os.Open(filePath)
		defer freader.Close()
		if err != nil {
			return
		}
		req, err = http.NewRequest(method, url, freader)

		var realmime string
		realmime, err = SniffMimeType(filePath)
		if err != nil {
			return
		}

		req.Header.Set("Content-Type", realmime)
	} else {
		req, err = http.NewRequest(method, url, buf)

		// set type & boundary
		req.Header.Set("Content-Type", ctype)
	}

	if err != nil {
		return
	}

	// extra headers
	for hname, hval := range extraHeaders {
		req.Header.Set(hname, hval)
	}

	// auth
	if username != "" {
		req.SetBasicAuth(username, password)
	}

	// send request
	client := &http.Client{}
	requestDump, err := httputil.DumpRequest(req, false)
	if err != nil {
		DebugPrintln(err)
	} else {
		DebugPrintln(fmt.Sprintf("%q", requestDump))
	}
	res, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	return
}
