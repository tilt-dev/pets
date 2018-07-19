package proc

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Process state that we expect to be written to disk and queried.
type PetsProc struct {
	// A name to show to humans
	DisplayName string `json:",omitempty"`

	// The process ID of a running process.
	Pid int `json:",omitempty"`

	// When the process started
	StartTime time.Time
}

type PetsCommand struct {
	Proc PetsProc
	Cmd  *exec.Cmd
}

// Checks if a process is still alive.
func isAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// 0 is a POSIX trick to ask the OS if it knows about the process
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
