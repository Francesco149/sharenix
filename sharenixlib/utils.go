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
	"fmt"
	"net/url"
	"os"
	"path"
	"time"
)

// Println prints the given thext only if silent is false
func Println(silent bool, a ...interface{}) (n int, err error) {
	if silent {
		return
	}
	return fmt.Println(a...)
}

// DebugPrintln prints the given text only
// if ShareNix is compiled with ShareNixDebug = true
func DebugPrintln(a ...interface{}) (n int, err error) {
	if !ShareNixDebug {
		return
	}
	fmt.Printf("Debug: ")
	return fmt.Println(a...)
}

// IsImage determines if a mime type is an image or not
func IsImage(mimeType string) bool {
	switch mimeType {
	case "image/bmp", "image/gif", "image/jpeg", "image/png":
		return true

	default:
		return false
	}

	return false
}

// IsUrl returns true if the file at the given path contains
// a plain text valid url
func IsUrl(filePath string) (bool, error) {
	const maxUrlSize = 2048

	data := make([]byte, maxUrlSize)
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return false, err
	}

	if stat.Size() > maxUrlSize {
		return false, nil
	}

	file.Read(data)

	_, err = url.ParseRequestURI(string(data))
	if err != nil {
		DebugPrintln("Not an url:", err)
		return false, nil
	}

	return true, nil
}

// FileExists returns true if the given directory or file exists
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GenerateArchivedFilename returns an unique file path inside
// archive/ that contains the current date, time and nanotime.
func GenerateArchivedFilename(extension string) (string, error) {
	t := time.Now()
	ye, mo, da := t.Date()
	hour, min, sec := t.Clock()

	archiveDir, err := GetArchiveDir()
	if err != nil {
		return "", err
	}

	return path.Join(archiveDir, fmt.Sprintf("%v-%v-%v_%v-%v-%v_%v.%s",
		ye, int(mo), da, hour, min, sec, t.UnixNano(), extension)), nil
}
