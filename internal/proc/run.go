package proc

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Run a command, waiting until it exits.
//
// args: The command to run.
// cwd: The current working directory.
func Run(args []string, cwd string) error {
	return RunWithIO(args, cwd, os.Stdout, os.Stderr)
}

// Run a command, waiting until it exits, forwarding all stdout/stderr to the given streams.
func RunWithIO(args []string, cwd string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("Empty args: %v", args)
	}

	cmd := exec.Command(args[0], args[1:]...)

	// Sets the process group ID so that if this process spawns sub-processes,
	// we can kill them later.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmd.Dir = cwd
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	pCmd, err := Start(cmd, args[0])
	if err != nil {
		return fmt.Errorf("Run: %v", err)
	}
	return pCmd.Cmd.Wait()
}

// Start a command, and return information about its running state.
func Start(cmd *exec.Cmd, displayName string) (PetsCommand, error) {
	err := cmd.Start()
	if err != nil {
		return PetsCommand{}, err
	}

	process := cmd.Process
	proc := PetsProc{
		Pid:         process.Pid,
		DisplayName: displayName,
		StartTime:   time.Now(),
	}
	return PetsCommand{
		Proc: proc,
		Cmd:  cmd,
	}, nil
}
