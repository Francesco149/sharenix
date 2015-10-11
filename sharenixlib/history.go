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
	"encoding/csv"
	//"fmt"
	"github.com/kardianos/osext"
	"os"
	"path"
)

// GetHistoryCSV returns the absolute path to the history csv.
func GetHistoryCSV() (csv string, err error) {
	exeFolder, err := osext.ExecutableFolder()
	if err != nil {
		return
	}
	csv = path.Join(exeFolder, "sharenix.csv")
	return
}

// GetUploadHistory returns all of the records in sharenix.csv
func GetUploadHistory() (res [][]string, err error) {
	csvPath, err := GetHistoryCSV()
	if err != nil {
		return
	}

	file, err := os.Open(csvPath)
	if err != nil {
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	res, err = reader.ReadAll()
	return
}

/*
func quote(str string) string {
	return fmt.Sprintf("\"%s\"", str)
}
*/

// AppendToHistory appends the given record to sharenix.csv
func AppendToHistory(url, thumbnailurl, deleteurl, filename string) (
	err error) {

	current, err := GetUploadHistory()
	if err != nil {
		current = make([][]string, 0)
	}

	// TODO: find a more efficient way to append to the file

	//current = append(current, []string{quote(url), quote(thumbnailurl),
	//	quote(deleteurl), quote(filename)})
	current = append(current, []string{url, thumbnailurl,
		deleteurl, filename})

	csvPath, err := GetHistoryCSV()
	if err != nil {
		return
	}

	file, err := os.Create(csvPath)
	if err != nil {
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = ';'
	err = writer.WriteAll(current)
	return
}
