package proc

import (
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/windmilleng/pets/internal/service"
)

// Process state that we expect to be written to disk and queried.
type PetsProc struct {
	// A name to show to humans
	DisplayName string `json:",omitempty"`

	// The process ID of a running process.
	Pid int `json:",omitempty"`

	// When the process started
	StartTime time.Time

	// The hostname that the process is listening on (e.g., 'localhost')
	Hostname string `json:",omitempty"`

	// The port that the process is listening on (e.g., '8080')
	Port int `json:",omitempty"`

	// The name+tier of the service that this process exposes
	ServiceName service.Name `json:",omitempty"`
	ServiceTier service.Tier `json:",omitempty"`
}

// Creates a new PetsProc that we know is listening on the given host and port.
//
// Calling this method automatically creates a copy because it's a struct method
// rather than a pointer method.
func (p PetsProc) WithExposedHost(hostname string, port int) PetsProc {
	p.Hostname = hostname
	p.Port = port
	return p
}

// Creates a new PetsProc that matches a service key.
//
// Calling this method automatically creates a copy because it's a struct method
// rather than a pointer method.
func (p PetsProc) WithServiceKey(key service.Key) PetsProc {
	p.ServiceName = key.Name
	p.ServiceTier = key.Tier
	return p
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
