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
	"os/exec"
	"path"
)

// RunPlugin starts pluginName in the plugin directory passing command-line
// params in the following format:
// 	pluginName -param1Name=param1Value ... -paramXName=paramXValue param_tail
// For example, calling
// 	RunPlugin("foo", map[string]string{
// 		"hello": "world",
// 		"someflag": "true",
//		"_tail": "bar",
// 	})
// will execute
// 	foo -hello=world -someflag=true bar
// Returns the last line outputted to stdout by the plugin and an error if any.
// Any trailing newlines at the end of the output are stripped.
func RunPlugin(pluginName string,
	extraParams map[string]string) (output string, err error) {

	formattedArgs := []string{extraParams["_tail"]}
	delete(extraParams, "_tail")
	for paramName, paramValue := range extraParams {
		formattedArgs = append(
			[]string{fmt.Sprintf("-%s=%s", paramName, paramValue)},
			formattedArgs...)
	}

	pluginsDir, err := GetPluginsDir()
	if err != nil {
		return
	}

	outdata, err := exec.Command(path.Join(pluginsDir, pluginName),
		formattedArgs...).CombinedOutput()
	DebugPrintln("exec.CombinedOutput returned:\n",
		string(outdata), "with error", err)
	if len(outdata) == 0 && err == nil {
		err = fmt.Errorf("Plugin did not return any output.")
		return
	}
	outdata = bytes.TrimSuffix(outdata, []byte{0x0A})
	ilastline := bytes.LastIndex(outdata, []byte{0x0A})
	if ilastline != -1 {
		outdata = outdata[ilastline+1:]
	} else {
		DebugPrintln("Plugin output was one line long")
	}
	output = string(outdata)
	return
}
