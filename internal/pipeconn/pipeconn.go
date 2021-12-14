// Package pipeconn implements a net.Conn that is a bidrectional pipe.
package pipeconn

import (
	"context"
	"net"
	"sync"
)

type Listener struct {
	mu   sync.Mutex
	ch   chan net.Conn
	done chan struct{}
}

func Listen() *Listener {
	return &Listener{
		ch:   make(chan net.Conn),
		done: make(chan struct{}),
	}
}

func (l *Listener) Accept() (net.Conn, error) {
	select {
	case <-l.done:
		return nil, net.ErrClosed
	case c := <-l.ch:
		return c, nil
	}
}

func (l *Listener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	select {
	case <-l.done:
		break
	default:
		close(l.done)
	}
	return nil
}

func (l *Listener) Addr() net.Addr {
	return addr{}
}

func (l *Listener) Dial() (net.Conn, error) {
	return l.DialContext(context.Background())
}

func (l *Listener) DialContext(ctx context.Context) (net.Conn, error) {
	c1, c2 := net.Pipe()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-l.done:
		return nil, net.ErrClosed
	case l.ch <- c1:
		return c2, nil
	}
}

type addr struct{}

func (addr) Network() string { return "pipe" }
func (addr) String() string  { return "pipe" }
