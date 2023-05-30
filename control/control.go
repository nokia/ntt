package control

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrNoFactory = errors.New("factory is not set")

// A Runner runs one or multiple jobs and emits Events
type Runner interface {
	Run(context.Context) <-chan Event
}

// RunnerFactory creates a new Runner.
type RunnerFactory func() (Runner, error)

// A Controller executes jobs in parallel.
type Controller struct {
	sync.Mutex
	maxWorkers int
	running    map[*Job]time.Time
	factory    RunnerFactory
}

// New creates a new Controller.
func New(opts ...Option) (*Controller, error) {
	c := &Controller{
		maxWorkers: 1,
		running:    make(map[*Job]time.Time),
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	if c.factory == nil {
		return nil, ErrNoFactory
	}
	return c, nil
}

func (c *Controller) Run(ctx context.Context) <-chan Event {
	wg := sync.WaitGroup{}
	results := make(chan Event, c.maxWorkers)
	for i := 0; i < c.maxWorkers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			worker, err := c.factory()
			if err != nil {
				results <- NewErrorEvent(err)
				return
			}

			for event := range worker.Run(ctx) {
				results <- event
			}
		}(i)
	}

	out := make(chan Event, c.maxWorkers)
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
				case StartEvent:
					c.running[ev.Job] = ev.Time()
				case StopEvent:
					ev.Begin = c.running[ev.Job]
				}
				out <- res
			case <-ticker.C:
				for job := range c.running {
					out <- NewTickerEvent(job)
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

type Option func(*Controller) error

func MaxWorkers(n int) Option {
	return func(c *Controller) error {
		c.maxWorkers = n
		return nil
	}
}

func WithFactory(f RunnerFactory) Option {
	return func(c *Controller) error {
		c.factory = f
		return nil
	}
}
