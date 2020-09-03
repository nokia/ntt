package log

import (
	"context"
	"io"
	"sync"
	"time"
)

type ConsoleLogger struct {
	Out io.Writer
	mu  sync.Mutex
}

type ConsoleTracer struct {
	name  string
	start time.Time
	ctx   context.Context
}

func (l *ConsoleLogger) Output(level Level, s string) error {
	if level > lvl {
		return nil
	}

	// TODO(5nord) Add support for terminal colors (have a look at logrus how it's done).

	l.mu.Lock()
	_, err := l.Out.Write([]byte(s))
	l.mu.Unlock()

	return err
}

func (t ConsoleTracer) Start(ctx context.Context, spanName string) (context.Context, Span) {
	if lvl < TraceLevel {
		return ctx, emptySpan{}
	}

	now := time.Now()
	Traceln(now, "trace enter", spanName)
	return ctx, ConsoleTracer{
		start: now,
		name:  spanName,
	}
}

func (t ConsoleTracer) End() {
	if lvl < TraceLevel {
		return
	}

	end := time.Now()
	Traceln(end, "trace leave", t.name, "took", end.Sub(t.start))
}

type emptySpan struct{}

func (e emptySpan) End() {}
