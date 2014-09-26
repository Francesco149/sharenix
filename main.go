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

package main

// NOTE: to compile this, you need >=gtk/gdk-3.10 and >=go-1.3.1
// You will also need my modified fork of gotk3: github.com/Francesco149/gotk3
// (go get it then rename it to github.com/conformal/gotk3 so that it can be
// properly imported)

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/Francesco149/sharenix/sharenixlib"
	"io/ioutil"
)

func loadConfig() (cfg *sharenixlib.Config, err error) {
	cfg = &sharenixlib.Config{}

	// load config
	file, err := ioutil.ReadFile("./sharenix.json")
	if err != nil {
		return
	}
	err = json.Unmarshal(file, &cfg)
	return
}

func handleCLI() (err error) {
	cfg, err := loadConfig()
	if err != nil {
		return
	}

	// command line flags
	pmode := flag.String("m", "f", "\n\tUpload mode\n\tf/file: upload file\n\tfs/fullscreen: "+
		"screenshot entire screen and upload\n\ts/section: select screen region and upload\n\t"+
		"c/clipboard: upload clipboard contents\n\tr/record: record screen region and upload\n\t"+
		"u/url: shorten url\n")

	psite := flag.String("s", "default", "\n\tTarget site name (default = default site for the selected mode)\n")
	psilent := flag.Bool("q", false, "\n\tQuiet mode - disables all terminal output except errors\n")
	pnotification := flag.Bool("n", false, "\n\t(not yet implemented) Notification - "+
		"displays a GTK notification for the upload\n")
	popen := flag.Bool("o", false, "\n\tOpen url - automatically opens the uploaded file's url in the default browser\n")
	phistory := flag.Bool("history", false, "\n\tShow upload history (grep-able)\n")
	pversion := flag.Bool("v", false, "\n\tShows the program version\n")

	flag.Parse()
	if !flag.Parsed() {
		panic(errors.New("Unexpected flag error"))
	}

	if *pversion {
		fmt.Println(sharenixlib.ShareNixVersion)
		return
	}

	if *phistory {
		var csv [][]string
		csv, err = sharenixlib.GetUploadHistory()
		if err != nil {
			return
		}

		if len(csv) == 0 {
			fmt.Println("Empty!")
			return
		}

		for _, record := range csv {
			if len(record) < 4 {
				err = errors.New("Invalid csv")
				return
			}

			fmt.Println("*", record[3], "- URL:", record[0], "Thumbnail URL:", record[1],
				"Deletion URL:", record[2])
			fmt.Println()
		}

		return
	}

	// perform upload
	url, thumburl, deleteurl, err := sharenixlib.ShareNix(cfg, *pmode, *psite, *psilent, *pnotification, *popen)
	if err != nil {
		return
	}

	// display results
	sharenixlib.Println(*psilent, "URL:", url)
	if len(thumburl) > 0 {
		sharenixlib.Println(*psilent, "Thumbnail URL:", thumburl)
	}
	if len(deleteurl) > 0 {
		sharenixlib.Println(*psilent, "Deletion URL:", deleteurl)
	}

	return
}

func main() {
	err := handleCLI()
	if err != nil {
		fmt.Println(err)
	}
}
