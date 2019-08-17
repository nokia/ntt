/*
Package session provides machine-wide unique sessions.

A session is a small unique integer, which is tied to a process id (PID) and may
typically be used to create a unique local IP address.

Note, this package was implemented to provide a smooth migration from Nokia
internal test tool (k3-run) to ntt and maybe replaced in the future.
*/
package session

var SharedDir = "/var/run/ntt"

// Clean returns a slice of alive sessions
func Clean(sessions []session) []session {
	list := make([]session, 0)
	for _, s := range sessions {
		if s.Alive() {
			list = append(list, s)
		}
	}
	return list
}
