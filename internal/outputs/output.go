// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package outputs

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/mitchellh/colorstring"
)

const (
	tabWidth    = 0
	tabPadding  = 2
	defColWidth = 0
	tabChar     = ' '
)

type Output struct {
	w         io.Writer
	tabWriter *tabwriter.Writer
	colorize  *colorstring.Colorize
}

func NewStdOutOutput() *Output {
	return &Output{
		w:         os.Stdout,
		tabWriter: initTabWriter(defColWidth, tabWidth, tabPadding, tabChar),
		colorize:  NewColorize(),
	}
}

func NewSilentOutput() *Output {
	return &Output{
		w:         io.Discard,
		tabWriter: initSilentTabWriter(),
	}
}

func (o *Output) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(o.w, format, a...)
}

func (o *Output) TabPrintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(o.tabWriter, format, a...)
}

func (o *Output) ColorPrintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(o.w, o.Color(format), a...)
}

func (o *Output) InitTabs(minColWidth int) {
	o.tabWriter.Flush()
	o.tabWriter = initTabWriter(minColWidth, tabWidth, tabPadding, tabChar)
}

func (o *Output) TabFlush() (err error) {
	return o.tabWriter.Flush()
}

func (o *Output) ErrorPrint(message string, cause error) (n int, err error) {
	if o.colorize != nil {
		return fmt.Fprint(o.w, o.colorize.Color(fmt.Sprintf("[bold][red]Error: [light_gray]%s: [reset][light_gray]%v\n", message, cause)))
	}
	return fmt.Fprint(o.w, o.colorize.Color(fmt.Sprintf("Error: %s: %s\n", message, cause)))
}

func (o *Output) Color(v string) string {
	if o.colorize != nil {
		return o.colorize.Color(v)
	}
	return v
}

func NewColorize() *colorstring.Colorize {
	return &colorstring.Colorize{
		Colors: colorstring.DefaultColors,
		Reset:  true,
	}
}

func initTabWriter(minWidth, tabWidth, padding int, padChar byte) *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, minWidth, tabWidth, padding, padChar, 0)
}

func initSilentTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(io.Discard, 0, 0, 0, 0, 0)
}
