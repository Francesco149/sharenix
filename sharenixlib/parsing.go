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
	"encoding/json"
	"fmt"
	"github.com/ChrisTrenkamp/goxpath"
	"github.com/ChrisTrenkamp/goxpath/tree/xmltree"
	"github.com/Francesco149/jsonpath"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// TODO: named regex support

// ParseRegexList parses a list of regular expressions on the given input
// and returns a slice of slices of strings with the match groups of each regex
func ParseRegexList(input string, regexList []string) (
	res [][]string, err error) {

	res = make([][]string, len(regexList))

	for i, regex := range regexList {
		var re *regexp.Regexp
		re, err = regexp.Compile(regex)
		if err != nil {
			return
		}

		matches := re.FindAllStringSubmatch(input, -1)

		if len(matches) > 0 {
			res[i] = matches[0]
		} else {
			res[i] = nil
		}
	}

	return
}

// parseRegexSyntax parses a $n$ or $n,n$ substring and returns the regexp match
// that should replace it
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

func parseJsonSyntax(syntax string, jsonblob []byte) string {
	DebugPrintln("jsonblob:", string(jsonblob))
	DebugPrintln("syntax:", syntax)

	paths, err := jsonpath.ParsePaths("$." + syntax + "+")
	if err != nil {
		DebugPrintln(err)
		return "(invalid jsonpath)" // TODO: throw errors
	}

	DebugPrintln("paths:", paths)

	eval, err := jsonpath.EvalPathsInBytes(jsonblob, paths)
	if err != nil {
		DebugPrintln(err)
		return "(invalid json)"
	}

	result, ok := eval.Next()
	if !ok || eval.Error != nil {
		DebugPrintln("result is", result, "err is", eval.Error)
		return "(jsonpath not found)"
	}

	DebugPrintln("jsonpath result:", result.Pretty(true))

	var val interface{}
	err = json.Unmarshal(result.Value, &val)
	if err != nil {
		DebugPrintln(err)
		return "(failed to parse json value)"
	}

	return fmt.Sprintf("%v", val)
}

func parseXmlSyntax(syntax string, xml []byte) string {
	xpExec, err := goxpath.Parse(syntax)
	if err != nil {
		return "(invalid xpath)"
	}

	xTree, err := xmltree.ParseXML(bytes.NewBuffer(xml))
	if err != nil {
		return "(invalid xml)"
	}

	res, err := xpExec.Exec(xTree)
	if err != nil {
		return "(xpath not found)"
	}

	return res.String()
}

const (
	_ = iota
	ParseRegex
	ParseJson
	ParseXml
)

// ParseUrl replaces the following syntaxes in url and returns the
// modified string.
// - regex matches: $regex:n,n$, $regex:n$, $n,n$, $n$
// - json paths: $json:some.json.element$
// - xml xpaths: $xml:/root/some/xml/element$
func ParseUrl(response []byte, url string, regexResults [][]string) string {
	// 1:1 port of ShareX's code to achieve the closest similarity in behaviour

	if len(url) == 0 {
		return ""
	}

	urlRunes := []rune(url)
	resultRunes := make([]rune, 0)

	syntaxStart := false
	parseType := ParseRegex
	syntaxStartIndex := 0

	for i := 0; i < len(urlRunes); i++ {
		if urlRunes[i] == rune('$') {
			if !syntaxStart {
				syntaxStart = true

				syntaxCheck := strings.ToLower(string(urlRunes[i+1:]))

				if strings.HasPrefix(syntaxCheck, "regex:") {
					parseType = ParseRegex
					syntaxStartIndex = i + 7
				} else if strings.HasPrefix(syntaxCheck, "json:") {
					parseType = ParseJson
					syntaxStartIndex = i + 6
				} else if strings.HasPrefix(syntaxCheck, "xml:") {
					parseType = ParseXml
					syntaxStartIndex = i + 5
				} else {
					syntaxStartIndex = i + 1
				}
			} else {
				parseText :=
					strings.TrimSpace(string(urlRunes[syntaxStartIndex:i]))

				if len(parseText) != 0 {
					var parseRes string

					switch parseType {
					default:
					case ParseRegex:
						parseRes = parseRegexSyntax(parseText, regexResults)
					case ParseJson:
						parseRes = parseJsonSyntax(parseText, response)
					case ParseXml:
						parseRes = parseXmlSyntax(parseText, response)
					}

					resultRunes = append(resultRunes, []rune(parseRes)...)
				}

				syntaxStart = false
			}
		} else if !syntaxStart {
			resultRunes = append(resultRunes, urlRunes[i])
		}
	}

	return string(resultRunes)
}

// Parses a uri list returned by "x-special/gnome-copied-files"
// and returns a slice of strings with all of the file uris
// note: this assumes that each file uri starts with file:/// which I hope is
//       the standard guaranteed format for x-special/gnome-copied-files.
func ParseUriList(list string) (res []*url.URL) {
	re := regexp.MustCompile(`file:/{2,}(-\.)?([^\s/?\.#-]+\.?)+(/[^\s]*)?`)
	uris := re.FindAllString(list, -1)
	for _, uri := range uris {
		u, err := url.Parse(uri)
		if err != nil {
			continue
		}
		res = append(res, u)
	}
	return
}
