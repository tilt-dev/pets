package proc

import (
	"os/exec"
	"time"
)

// Process state that we expect to be written to disk and queried.
type PetsProc struct {
	// A name to show to humans
	DisplayName string

	// The process ID of a running process.
	Pid int

	// When the process started
	StartTime time.Time
}

type PetsCommand struct {
	Proc PetsProc
	Cmd  *exec.Cmd
}
