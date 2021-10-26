package app

import (
	"fmt"
	"io"
	"os"

	"github.com/mitchellh/colorstring"
)

type Output struct {
	w        io.Writer
	colorize colorstring.Colorize
}

func init() {
	def = Output{
		w: os.Stdout,
		colorize: colorstring.Colorize{
			Colors: colorstring.DefaultColors,
			Reset:  true,
		},
	}
}

var def Output

func Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(def.w, format, a...)
}

func ErrorPrint(message string, cause error) (n int, err error) {
	return fmt.Fprint(def.w, def.colorize.Color(fmt.Sprintf("[bold][red]Error: [white]%s: [reset][white]%v\n", message, cause)))
}

func Color(v string) string {
	return def.colorize.Color(v)
}
