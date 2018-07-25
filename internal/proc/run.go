package proc

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type Runner struct {
	fs ProcFS
}

func NewRunner(fs ProcFS) Runner {
	return Runner{
		fs: fs,
	}
}

// Run a command, waiting until it exits.
//
// args: The command to run.
// cwd: The current working directory.
func (r Runner) Run(args []string, cwd string) error {
	return r.RunWithIO(args, cwd, os.Stdout, os.Stderr)
}

// Run a command, waiting until it exits, forwarding all stdout/stderr to the given streams.
func (r Runner) RunWithIO(args []string, cwd string, stdout, stderr io.Writer) error {
	pCmd, err := r.startWithIO(args, cwd, stdout, stderr)
	if err != nil {
		return fmt.Errorf("Run: %v", err)
	}
	err = pCmd.Cmd.Wait()
	r.fs.RemoveProc(pCmd.Proc)
	return err
}

// Starts a command, waiting until it exits, forwarding all stdout/stderr to the given streams.
func (r Runner) startWithIO(args []string, cwd string, stdout, stderr io.Writer) (PetsCommand, error) {
	if len(args) == 0 {
		return PetsCommand{}, fmt.Errorf("Empty args: %v", args)
	}

	cmd := exec.Command(args[0], args[1:]...)

	// Sets the process group ID so that if this process spawns sub-processes,
	// we can kill them later.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmd.Dir = cwd
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return r.startCmd(cmd, args[0])
}

// Start a command, and return information about its running state.
func (r Runner) startCmd(cmd *exec.Cmd, displayName string) (PetsCommand, error) {
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
	err = r.fs.AddProc(proc)
	if err != nil {
		return PetsCommand{}, err
	}
	return PetsCommand{
		Proc: proc,
		Cmd:  cmd,
	}, nil
}
