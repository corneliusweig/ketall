/*
Copyright 2019 Cornelius Weig

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package color

import (
	"fmt"
	"io"
	"strings"
)

type TableColor int

var (
	None        = TableColor(0)
	Red         = TableColor(31)
	Green       = TableColor(32)
	Yellow      = TableColor(33)
	Blue        = TableColor(34)
	Purple      = TableColor(35)
	Cyan        = TableColor(36)
	White       = TableColor(37)
	LightRed    = TableColor(91)
	LightGreen  = TableColor(92)
	LightYellow = TableColor(93)
	LightBlue   = TableColor(94)
	LightPurple = TableColor(95)
	LightCyan   = TableColor(96)
)

func (c TableColor) Fprint(out io.Writer, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(out, "\xff\033[1;%dm\xff%s\xff\033[0m\xff", c, fmt.Sprint(a...))
}

func (c TableColor) Fprintln(out io.Writer, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(out, "\xff\033[1;%dm\xff%s\xff\033[0m\xff\n", c, strings.TrimSuffix(fmt.Sprintln(a...), "\n"))
}

func (c TableColor) Fprintf(out io.Writer, format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(out, "\xff\033[1;%dm\xff%s\xff\033[0m\xff", c, fmt.Sprintf(format, a...))
}
