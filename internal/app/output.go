//Copyright 2022 Google LLC
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package app

import (
	"fmt"
	"io"
	"os"

	"github.com/mitchellh/colorstring"
)

type Output struct {
	w        io.Writer
	colorize *colorstring.Colorize
}

func NewStdOutOutput() *Output {
	return &Output{
		w:        os.Stdout,
		colorize: NewColorize(),
	}
}

func NewSilentOutput() *Output {
	return &Output{
		w: io.Discard,
	}
}

func (o *Output) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(o.w, format, a...)
}

func (o *Output) ColorPrintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(o.w, o.Color(format), a...)
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
