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
	"github.com/kardianos/osext"
	"os"
	"os/user"
	"path"
)

func GetHome() (res string) {
	if res = os.Getenv("HOME"); res != "" {
		return
	}

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	res = usr.HomeDir
	return
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

// GetHistoryCSV returns the absolute path to the history csv.
func GetHistoryCSV() (csv string, err error) {
	exeFolder, err := osext.ExecutableFolder()
	if err != nil {
		return
	}
	csv = path.Join(exeFolder, "sharenix.csv")
	return
}
