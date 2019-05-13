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

type Color int

var (
	None        = Color(0)
	Red         = Color(31)
	Green       = Color(32)
	Yellow      = Color(33)
	Blue        = Color(34)
	Purple      = Color(35)
	Cyan        = Color(36)
	White       = Color(37)
	LightRed    = Color(91)
	LightGreen  = Color(92)
	LightYellow = Color(93)
	LightBlue   = Color(94)
	LightPurple = Color(95)
	LightCyan   = Color(96)
)

func (c Color) Fprint(out io.Writer, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(out, "\xff\033[%dm\xff%s\xff\033[0m\xff", c, fmt.Sprint(a...))
}

func (c Color) Fprintln(out io.Writer, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(out, "\xff\033[%dm\xff%s\xff\033[0m\xff\n", c, strings.TrimSuffix(fmt.Sprintln(a...), "\n"))
}

func (c Color) Fprintf(out io.Writer, format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(out, "\xff\033[%dm\xff%s\xff\033[0m\xff", c, fmt.Sprintf(format, a...))
}
