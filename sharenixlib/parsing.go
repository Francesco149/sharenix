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
	"regexp"
	"strconv"
	"unicode"
)

// TODO: named regex support

// ParseRegexList parses a list of regular expressions on the given input and returns
// a slice of slices of strings with the match groups of each regex
func ParseRegexList(input string, regexList []string) (res [][]string, err error) {
	res = make([][]string, len(regexList))

	for i, regex := range regexList {
		var re *regexp.Regexp
		re, err = regexp.Compile(regex)
		if err != nil {
			return
		}

		res[i] = re.FindAllStringSubmatch(input, -1)[0]
	}

	return
}

// parseRegexSyntax parses a $n$ or $n,n$ substring and returns the regexp match that should replace it
func parseRegexSyntax(text string, regexResults [][]string) string {
	// 1:1 port of ShareX's code to achieve the closest similarity in behaviour

	if len(text) == 0 {
		return ""
	}

	regexIndexString := make([]rune, 0)
	var regexIndex int
	isGroupRegex := false
	i := 0

	textRunes := []rune(text)

	for ; i < len(textRunes); i++ {
		if unicode.IsDigit(textRunes[i]) {
			regexIndexString = append(regexIndexString, textRunes[i])
		} else {
			if textRunes[i] == rune(',') {
				isGroupRegex = true
			}

			break
		}
	}

	if len(regexIndexString) == 0 {
		return ""
	}

	regexIndex, err := strconv.Atoi(string(regexIndexString))
	if err != nil {
		return ""
	}

	if regexIndex < 1 || regexIndex > len(regexResults) {
		return ""
	}

	match := regexResults[regexIndex-1]

	if isGroupRegex && i+1 < len(textRunes) {
		group := textRunes[i+1:]
		groupNumber, err := strconv.Atoi(string(group))
		if err != nil {
			return ""
		}

		if groupNumber < 0 || groupNumber >= len(match) {
			return ""
		}

		return match[groupNumber]
	}

	return match[0]
}

// ParseUrl replaces $n$ and $n,n$ tags in the given url with the proper regex matches
func ParseUrl(url string, regexResults [][]string) string {
	// 1:1 port of ShareX's code to achieve the closest similarity in behaviour

	if len(url) == 0 {
		return ""
	}

	urlRunes := []rune(url)
	resultRunes := make([]rune, 0)
	regexStart := false
	regexStartIndex := 0

	for i := 0; i < len(urlRunes); i++ {
		if urlRunes[i] == rune('$') {
			if !regexStart {
				regexStart = true
				regexStartIndex = i
			} else {
				syntax := string(urlRunes[regexStartIndex+1 : i])
				regexResult := parseRegexSyntax(syntax, regexResults)

				if len(regexResult) != 0 {
					resultRunes = append(resultRunes, []rune(regexResult)...)
				}

				regexStart = false
				continue
			}
		}

		if !regexStart {
			resultRunes = append(resultRunes, urlRunes[i])
		}
	}

	return string(resultRunes)
}

// Parses a uri list returned by "x-special/gnome-copied-files"
// and returns a slice of strings with all of the file uris
func ParseUriList(list string) []string {
	re := regexp.MustCompile("((?:\\/[\\w\\.\\-]+)+)")
	return re.FindAllString(list, -1)
}
