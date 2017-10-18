package runtime

import "syscall"

func setPdeathsig(p *syscall.SysProcAttr, sig syscall.Signal) *syscall.SysProcAttr {
	if p == nil {
		p = &syscall.SysProcAttr{}
	}
	p.Pdeathsig = sig
	return p
}
