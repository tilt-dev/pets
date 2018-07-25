package mill

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/windmilleng/pets/internal/proc"
)

func TestPrint(t *testing.T) {
	dir, _ := ioutil.TempDir("", t.Name())
	file := filepath.Join(dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`print("hello")`), os.FileMode(0777))

	petsitter, stdout := newTestPetsitter(t)
	petsitter.ExecFile(file)
	defer os.RemoveAll(dir)

	out := stdout.String()
	if out != "hello\n" {
		t.Errorf("Expected 'hello'. Actual: %s", out)
	}
}

func TestPrintFail(t *testing.T) {
	dir, _ := ioutil.TempDir("", t.Name())
	file := filepath.Join(dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`print(hello)`), os.FileMode(0777))
	defer os.RemoveAll(dir)

	petsitter, stdout := newTestPetsitter(t)
	err := petsitter.ExecFile(file)
	out := stdout.String()
	if !(out == "" && strings.Contains(err.Error(), "undefined: hello")) {
		t.Errorf("Expected 'hello'. Actual: %s. Err: %s", out, err)
	}
}

func TestRun(t *testing.T) {
	dir, _ := ioutil.TempDir("", t.Name())
	file := filepath.Join(dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`run("echo meow")`), os.FileMode(0777))

	petsitter, stdout := newTestPetsitter(t)
	petsitter.ExecFile(file)
	defer os.RemoveAll(dir)

	out := stdout.String()
	if out != "meow\n" {
		t.Errorf("Expected 'meow'. Actual: %s", out)
	}
}

func newTestPetsitter(t *testing.T) (*Petsitter, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	procfs, err := proc.NewProcFS()
	if err != nil {
		t.Fatal(err)
	}
	runner := proc.NewRunner(procfs)
	return &Petsitter{
		Stdout: stdout,
		Stderr: stderr,
		Runner: runner,
	}, stdout
}
