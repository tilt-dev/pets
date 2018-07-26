package proc

import (
	"bytes"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	f := newProcFixture(t)
	defer f.tearDown()

	procfs := f.newProcFS()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cwd, _ := os.Getwd()

	r := NewRunner(procfs)
	err := r.RunWithIO([]string{"echo", "hello"}, cwd, stdout, stderr)
	if err != nil {
		t.Fatal(err)
	}

	out := stdout.String()
	if out != "hello\n" {
		t.Errorf("Expected 'hello'. Actual: %s", out)
	}
}

func TestStartAddsToProcFS(t *testing.T) {
	f := newProcFixture(t)
	defer f.tearDown()

	procfs := f.newProcFS()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cwd, _ := os.Getwd()

	r := NewRunner(procfs)
	petsCmd, err := r.StartWithIO([]string{"sleep", "10"}, cwd, stdout, stderr)
	if err != nil {
		t.Fatal(err)
	}
	defer petsCmd.Cmd.Process.Kill()

	procs, err := procfs.procsFromFS()
	if err != nil {
		t.Fatal(err)
	}

	if len(procs) != 1 || procs[0].Pid != petsCmd.Proc.Pid {
		t.Fatalf("Unexpected procs on disk: %v", procs)
	}
}
