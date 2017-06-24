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
	"os"
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

// DebugPrintf formats and prints the given text only
// if ShareNix is compiled with ShareNixDebug = true
func DebugPrintf(format string, a ...interface{}) (n int, err error) {
	if !ShareNixDebug {
		return
	}
	fmt.Printf("Debug: ")
	return fmt.Printf(format, a...)
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

func MkDirIfNotExists(dir string) error {
	direxists, err := FileExists(dir)
	if err != nil {
		return err
	}
	if !direxists {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}
