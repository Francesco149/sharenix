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
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// SendGetRequest sends a GET request with params
func SendGetRequest(baseurl string, params map[string]string) (
	res *http.Response, err error) {

	u, err := url.Parse(baseurl)
	if err != nil {
		return
	}

	q := u.Query()
	for name, value := range params {
		q.Set(name, value)
	}
	u.RawQuery = q.Encode()

	DebugPrintln(u)
	res, err = http.Get(u.String())
	return
}

// SendGetRequest sends a POST request with params
func SendPostRequest(baseurl string, params map[string]string) (
	*http.Response, error) {

	q := url.Values{}
	for name, value := range params {
		q.Set(name, value)
	}
	DebugPrintln(baseurl, q)
	return http.PostForm(baseurl, q)
}

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

// SendFilePostRequest prepares a multipart
// file upload POST request and sends it
func SendFilePostRequest(url, fileParamName, filePath string,
	extraParams map[string]string, extraHeaders map[string]string,
) (res *http.Response, filename string, err error) {

	filename = filepath.Base(filePath)

	realmime, err := SniffMimeType(filePath)
	if err != nil {
		return
	}

	// try opening the file
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	// prepare request body buffer
	buf := &bytes.Buffer{}

	// create a multipart file header with the given param name and file
	w := multipart.NewWriter(buf)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(fileParamName), escapeQuotes(filename)))
	h.Set("Content-Type", realmime)

	formfile, err := w.CreatePart(h)
	if err != nil {
		return
	}

	// write the file to the file header
	_, err = io.Copy(formfile, file)

	// append extra params
	for param, val := range extraParams {
		err = w.WriteField(param, val)
		if err != nil {
			return
		}
	}

	ctype := w.FormDataContentType()

	err = w.Close()
	if err != nil {
		return
	}

	// finally create the request
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return
	}

	// set type & boundary
	req.Header.Set("Content-Type", ctype)

	// extra headers
	for hname, hval := range extraHeaders {
		req.Header.Set(hname, hval)
	}

	// send request
	client := &http.Client{}
	DebugPrintln(req.URL, extraParams, extraHeaders)
	res, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	return
}
