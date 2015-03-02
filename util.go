/*
* Copyright 2015 Fabrício Godoy
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package gocli

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

const (
	PARAMETERS_PATTERN = `"((?:[^"]|\")*)"|(\S+)`
)

var rArgs *regexp.Regexp

func init() {
	rArgs, _ = regexp.Compile(PARAMETERS_PATTERN)
}

func readString() (string, error) {
	stdin := bufio.NewReader(os.Stdin)
	in, err := stdin.ReadString('\n')
	if err != nil {
		return "", err
	}

	in = strings.Replace(in, "\n", "", -1)
	in = strings.Replace(in, "\r", "", -1)

	return in, nil
}

func parseArgs(args string) []string {
	matches := rArgs.FindAllStringSubmatch(args, -1)
	if matches == nil {
		return []string{args}
	}

	retArgs := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m[1]) > 0 {
			retArgs = append(retArgs, m[1])
		} else {
			retArgs = append(retArgs, m[2])
		}
	}

	return retArgs
}
