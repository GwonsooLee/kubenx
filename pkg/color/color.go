package color

import (
	"fmt"
	"github.com/fatih/color"
	"strings"
	"io"
)

type Color struct {
	color *color.Color
}

var (
	//Color Definition
	Red    = color.New(color.FgRed)
	Blue   = color.New(color.FgBlue)
	Green  = color.New(color.FgGreen)
	Yellow = color.New(color.FgYellow)
	Cyan   = color.New(color.FgCyan)
)

// Fprintln outputs the result to out, followed by a newline.
func (c Color) Fprintln(out io.Writer, a ...interface{}) {
	if c.color == nil {
		fmt.Fprintln(out, a...)
		return
	}

	fmt.Fprintln(out, c.color.Sprint(strings.TrimSuffix(fmt.Sprintln(a...), "\n")))
}
