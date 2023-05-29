package pool

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/nokia/ntt/control"
)

var ErrNoFactory = errors.New("factory is not set")

// A Runner executes jobs in parallel.
type Runner struct {
	sync.Mutex
	maxWorkers int
	running    map[*control.Job]time.Time
	factory    control.RunnerFactory
}

func NewRunner(opts ...Option) (*Runner, error) {
	r := &Runner{
		maxWorkers: 1,
		running:    make(map[*control.Job]time.Time),
	}
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}
	if r.factory == nil {
		return nil, ErrNoFactory
	}
	return r, nil
}

func (r *Runner) Run(ctx context.Context) <-chan control.Event {
	wg := sync.WaitGroup{}
	results := make(chan control.Event, r.maxWorkers)
	for i := 0; i < r.maxWorkers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			worker, err := r.factory()
			if err != nil {
				results <- control.NewErrorEvent(err)
				return
			}

			for event := range worker.Run(ctx) {
				results <- event
			}
		}(i)
	}

	out := make(chan control.Event, r.maxWorkers)
	go func() {
		const secs = time.Duration(30.0)
		ticker := time.NewTicker(secs * time.Second)
		for {
			select {
			case res, ok := <-results:
				if !ok {
					close(out)
					return
				}
				ticker.Reset(secs * time.Second)
				switch ev := res.(type) {
				case control.StartEvent:
					r.running[ev.Job] = ev.Time()
				case control.StopEvent:
					ev.Begin = r.running[ev.Job]
				}
				out <- res
			case <-ticker.C:
				for job := range r.running {
					out <- control.NewTickerEvent(job)
				}
			}
		}
	}()

	// Wait for all workers to finish.
	go func() {
		wg.Wait()
		close(results)
	}()

	return out
}

type Option func(*Runner) error

func MaxWorkers(n int) Option {
	return func(r *Runner) error {
		r.maxWorkers = n
		return nil
	}
}

func WithFactory(f control.RunnerFactory) Option {
	return func(r *Runner) error {
		r.factory = f
		return nil
	}
}
