package printer

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/nokia/ntt/control"
)

var (
	ColorFatal   = color.New(color.FgRed, color.Bold)
	ColorFailure = color.New(color.FgRed, color.Bold)
	ColorWarning = color.New(color.FgYellow, color.Bold)
	ColorSuccess = color.New()
	ColorStart   = color.New()
	ColorRunning = color.New(color.Faint)
	ColorLog     = color.New(color.Faint)
	Colors       = func(v string) *color.Color {
		switch v {
		case "pass":
			return ColorSuccess
		case "inconc":
			return ColorWarning
		case "none":
			return ColorWarning
		case "done":
			return color.New()
		default:
			return ColorFailure
		}
	}

	ErrCommandFailed = fmt.Errorf("command failed")
)

type Printer interface {
	Print(ev control.Event)
}
