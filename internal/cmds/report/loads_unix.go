// +build !windows

//
package report

import "syscall"

func loads() [3]float64 {
	var si syscall.Sysinfo_t
	syscall.Sysinfo(&si)
	var a = [3]float64{
		float64(si.Loads[0]) / 65536.0,
		float64(si.Loads[1]) / 65536.0,
		float64(si.Loads[2]) / 65536.0,
	}
	return a
}
