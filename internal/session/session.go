// Package session
package session

var SharedDir = "/var/run/ntt"

func Clean(sessions []session) []session {
	list := make([]session, 0)
	for _, s := range sessions {
		if s.Alive() {
			list = append(list, s)
		}
	}
	return list
}
