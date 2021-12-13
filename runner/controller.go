package runner

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/nokia/ntt/protobuf"
)

// The Test Controller provides a set of operations to control and monitor the
// test execution. It tracks which test runners are available for running
// tests, hands out tests to runners, and returns results as well as other data
// files.
//
// The Controller is the equivalent to the ETSI Test Controller and Management
// Interface (TCI-TM)
type Controller struct {
	// gRPC requirement to have forward compatible implementations.
	pb.UnimplementedControlServer

	// Running tests
	tests   map[string]test
	testsMu sync.Mutex
}

// A test that is being run.
type test struct {
	id  string             // Unique ID
	req *pb.RunTestRequest // The request that started this test.
	r   Runner             // The backend that is running this test.
}

func (ctrl *Controller) RunTest(ctx context.Context, req *pb.RunTestRequest) (*pb.RunTestResponse, error) {
	ctrl.testsMu.Lock()
	defer ctrl.testsMu.Unlock()

	// TODO(5nord): Create Runner where to get the suite from?

	t := test{
		id:  fmt.Sprint(len(ctrl.tests) + 1),
		req: req,
	}

	ctrl.tests[t.id] = t

	return &pb.RunTestResponse{
		ID: t.id,
	}, nil
}

func (ctrl *Controller) Subscribe(*pb.SubscribeRequest, pb.Control_SubscribeServer) error {

	// TODO(5nord): How broadcast to all subscribers?
	return nil
}
