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

func (o *Output) ErrorPrint(message string, cause error) (n int, err error) {
	if o.colorize != nil {
		return fmt.Fprint(o.w, o.colorize.Color(fmt.Sprintf("[bold][red]Error: [white]%s: [reset][white]%v\n", message, cause)))
	}
	return fmt.Fprint(o.w, o.colorize.Color(fmt.Sprintf("Error: %s: %v\n", message, cause)))
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
