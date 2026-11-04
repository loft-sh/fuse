package fuse

import (
	"bytes"
	"errors"
	"os/exec"
	"syscall"
)

func unmount(dir string) error {
	// Like mount, only use the helper if we may not unmount ourselves.
	err := syscall.Unmount(dir, 0)
	if err != syscall.EPERM && err != syscall.EACCES {
		return err
	}

	cmd := exec.Command("fusermount", "-u", dir)
	output, err := cmd.CombinedOutput()
	if err != nil && err.Error() != errNoChildProcesses && err.Error() != "waitid: no child processes" {
		if len(output) > 0 {
			output = bytes.TrimRight(output, "\n")
			msg := err.Error() + ": " + string(output)
			err = errors.New(msg)
		}
		return err
	}
	return nil
}
