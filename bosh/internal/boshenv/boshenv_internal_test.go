package boshenv

import "os/exec"

func FakeExec(execCmd func(string, ...string) *exec.Cmd) Option {
	return func(c *BOSHCLI) error {
		c.execCmd = execCmd
		return nil
	}
}
