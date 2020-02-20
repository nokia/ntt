package ntt

import (
	"github.com/nokia/ntt/internal/session"
)

type Suite struct {
	id int // A unique session id
}

// Id returns the unique session id (aka K3_SESSION_ID). This ID is the smallest
// integer available on this machine.
func (s *Suite) Id() (int, error) {
	if s.id == 0 {
		id, err := session.Get()
		if err != nil {
			return 0, err
		}
		s.id = id
	}
	return s.id, nil
}

func init() {
	// TODO(5nord) We still have to figure how this sharedDir could be handled
	// more elegantly, maybe even with support for Windows.
	//
	// Change SharedDir to /tmp/k3 to be compatible with legacy k3 scripts.
	session.SharedDir = "/tmp/k3"
}
