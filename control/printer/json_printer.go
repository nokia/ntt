package printer

import (
	"fmt"
	"strings"

	"github.com/nokia/ntt/control"
	"github.com/nokia/ntt/internal/yaml"
)

type JSONPrinter struct {
}

func NewJSONPrinter() *JSONPrinter {
	return &JSONPrinter{}
}

func quote(s string) string {
	quoted, _ := yaml.MarshalJSON(s)
	return strings.TrimSpace(string(quoted))
}

func (p *JSONPrinter) Print(ev control.Event) {
	// control.Event provides too much information. Until we've figured out
	// what we need we display reduced information
	switch ev := ev.(type) {
	case control.LogEvent:
		fmt.Printf(`{time: %d, event: "log", job_id: "%s", text: %s }`, ev.Time().Unix(), ev.ID, quote(ev.Text))
	case control.StartEvent:
		fmt.Printf(`{time: %d, event: "start", job_id: "%s", name: %s }`, ev.Time().Unix(), ev.ID, quote(ev.Name))
	case control.TickerEvent:
	case control.StopEvent:
		fmt.Printf(`{time: %d, event: "stop", job_id: "%s", name: %s, verdict: %s}`, ev.Time().Unix(), ev.ID, quote(ev.Name), quote(ev.Verdict))
	case control.ErrorEvent:
		jobID := ""
		if job := control.UnwrapJob(ev); job != nil {
			jobID = job.ID + ", "
		}
		fmt.Printf(`{time: %d, event: "error", %stext: %s }`, ev.Time().Unix(), jobID, quote(ev.Err.Error()))
	default:
		panic(fmt.Sprintf("unknown event type %T", ev))
	}
	fmt.Println()
}
