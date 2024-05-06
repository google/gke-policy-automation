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

// Package outputs implements data output produced by the application
package outputs

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
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
}

func NewStdOutOutput() *Output {
	return &Output{
		w:         os.Stdout,
		tabWriter: initTabWriter(os.Stdout, defColWidth, tabWidth, tabPadding, tabChar),
	}
}

func NewSilentOutput() *Output {
	return &Output{
		w:         io.Discard,
		tabWriter: initTabWriter(io.Discard, 0, 0, 0, 0),
	}
}

func (o *Output) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(o.w, format, a...)
}

func (o *Output) TabPrintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(o.tabWriter, format, a...)
}

func (o *Output) InitTabs(minColWidth int, tabPadding int) {
	o.tabWriter.Flush()
	o.tabWriter = initTabWriter(o.w, minColWidth, tabWidth, tabPadding, tabChar)
}

func (o *Output) TabFlush() (err error) {
	return o.tabWriter.Flush()
}

func (o *Output) ErrorPrint(message string, cause error) (n int, err error) {
	errF := color.New(color.Bold, color.FgHiRed).Sprint
	errTitleF := color.New(color.Bold, color.FgHiWhite).Sprintf
	return fmt.Fprintf(o.w, "%s %s %v\n",
		errF("Error:"),
		errTitleF("%s:", message),
		cause,
	)
}

func initTabWriter(output io.Writer, minWidth, tabWidth, padding int, padChar byte) *tabwriter.Writer {
	return tabwriter.NewWriter(output, minWidth, tabWidth, padding, padChar, 0)
}
