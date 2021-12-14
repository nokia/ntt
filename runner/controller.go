package runner

import (
	"context"
	"fmt"
	"sync"

	"github.com/nokia/ntt/project"
	pb "github.com/nokia/ntt/protobuf"
	"github.com/nokia/ntt/runner/k3s"
)

// The Test Controller provides a set of operations to control and monitor the
// test execution. It tracks which test runners are available for running
// tests, hands out tests to runners, and returns results as well as other data
// files.
//
// The Controller is the equivalent to the ETSI Test Controller and Management
// Interface (TCI-TM)
type Controller struct {
	// Log is the logger used by the controller.
	Log func(string)

	// gRPC requirement to have forward compatible implementations.
	pb.UnimplementedControlServer

	// Running tests
	tests   map[string]test
	testsMu sync.Mutex

	// Projects
	projects   map[string]project.Interface
	projectsMu sync.Mutex
}

// A test that is being run.
type test struct {
	id  string             // Unique ID
	req *pb.RunTestRequest // The request that started this test.
	r   Runner             // The backend that is running this test.
}

// A project is a collection of test files.
type proj struct {
	id      string
	name    string
	root    string
	sources []string
	imports []string
}

func (p *proj) Root() string {
	return p.root
}

func (p *proj) Sources() ([]string, error) {
	return p.sources, nil
}

func (p *proj) Imports() ([]string, error) {
	return p.imports, nil
}

// Writer interface for passing logs from backends to the user.
func (ctrl *Controller) Write(p []byte) (int, error) {
	s := string(p)
	if ctrl.Log != nil {
		ctrl.Log(s)
	}
	return len(s), nil
}

func (ctrl *Controller) RunTest(ctx context.Context, req *pb.RunTestRequest) (*pb.RunTestResponse, error) {

	p, err := ctrl.getProject(req.ProjectId)
	if err != nil {
		return nil, err
	}

	r, err := k3s.New(ctrl, p)
	if err != nil {
		return nil, err
	}

	//r.Run(ctrl, req.GetName())

	t := test{
		id:  fmt.Sprint(len(ctrl.tests) + 1),
		req: req,
		r:   r,
	}

	ctrl.testsMu.Lock()
	defer ctrl.testsMu.Unlock()

	ctrl.tests[t.id] = t

	return &pb.RunTestResponse{
		ID: t.id,
	}, nil
}

func (ctrl *Controller) Subscribe(*pb.SubscribeRequest, pb.Control_SubscribeServer) error {
	// TODO(5nord): How broadcast to all subscribers?
	return nil
}

func (ctrl *Controller) RegisterProject(ctx context.Context, req *pb.RegisterProjectRequest) (*pb.RegisterProjectResponse, error) {
	ctrl.projectsMu.Lock()
	defer ctrl.projectsMu.Unlock()

	p := &proj{
		id:      fmt.Sprint(len(ctrl.projects) + 1),
		name:    req.GetName(),
		root:    req.GetRootDir(),
		sources: req.GetSources(),
		imports: req.GetDependencies(),
	}
	ctrl.projects[p.id] = p
	return &pb.RegisterProjectResponse{
		ID: p.id,
	}, nil
}

func (ctrl *Controller) getProject(id string) (project.Interface, error) {
	ctrl.projectsMu.Lock()
	defer ctrl.projectsMu.Unlock()

	if p, ok := ctrl.projects[id]; ok {
		return p, nil
	}

	return nil, fmt.Errorf("project %s not found", id)
}
