package proc

import (
	"os/exec"
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
