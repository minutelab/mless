// +build !linux

package runtime

import "syscall"

func setPdeathsig(p *syscall.SysProcAttr, sig syscall.Signal) *syscall.SysProcAttr {
	return nil
}
