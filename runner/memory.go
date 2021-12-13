package runner

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/nokia/ntt/protobuf"
	"google.golang.org/grpc"
)

// NewInMemoryController creates a test controller running in a Go routine
func NewInMemoryController() (*grpc.ClientConn, error) {
	ln := &memListener{
		c: make(chan net.Conn),
	}

	go func() {
		s := grpc.NewServer()
		pb.RegisterControlServer(s, &Controller{tests: make(map[string]test)})
		log.Fatal(s.Serve(ln))
	}()

	conn, err := grpc.Dial("mem",
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			c1, c2 := net.Pipe()
			ln.c <- c1
			return c2, nil
		}))

	return conn, err
}

type memAddr string

func (a memAddr) Network() string { return "mem" }
func (a memAddr) String() string  { return string(a) }

type memListener struct {
	c chan net.Conn
}

func (ln *memListener) Accept() (net.Conn, error) {
	conn, ok := <-ln.c
	if !ok {
		return nil, fmt.Errorf("closed")
	}
	return conn, nil
}

func (ln *memListener) Addr() net.Addr {
	return memAddr(fmt.Sprintf("%p", ln))
}

func (ln *memListener) Close() error {
	close(ln.c)
	return nil
}
