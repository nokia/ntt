package printer

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/nokia/ntt/control"
)

type TAPPrinter struct {
	n       int
	success int
	failed  int
}

func NewTAPPrinter() *TAPPrinter {
	return &TAPPrinter{}
}

func (p *TAPPrinter) Print(ev control.Event) {
	switch ev := ev.(type) {
	case control.LogEvent:
		fmt.Printf("# %s\n", strings.ReplaceAll(strings.TrimRightFunc(ev.Text, unicode.IsSpace), "\n", "\n# "))
	case control.StartEvent:
		p.n++
		fmt.Printf("# %s (%s): started\n", ev.Name, ev.ID)
	case control.TickerEvent:
	case control.StopEvent:
		verdict := "not ok"
		if verdict == "pass" || verdict == "done" {
			verdict = "ok"
			p.success++
		} else {
			p.failed++
		}
		fmt.Printf("%s %d - %s\n", verdict, p.n, ev.ID)
	case control.ErrorEvent:
		jobID := ""
		if job := control.UnwrapJob(ev); job != nil {
			jobID = " " + job.ID + ":"
		}
		fmt.Printf("#%s error: %s\n", jobID, ev.Error())
	default:
		panic(fmt.Sprintf("unknown event type %T", ev))
	}
}

func (p *TAPPrinter) Close() error {
	switch {
	case p.n == 0:
		fmt.Println("1..0 # SKIP no tests")
	case p.success == p.n && p.failed == 0:
		fmt.Printf("# passed all %d tests.\n", p.n)
		fmt.Printf("1..%d\n", p.n)
	default:
		fmt.Printf("# failed %d among %d tests.\n", p.n-p.success, p.n)
		fmt.Printf("1..%d\n", p.n)
	}
	return nil
}
