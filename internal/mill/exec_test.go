package mill

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrint(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	dir, _ := ioutil.TempDir("", t.Name())
	file := filepath.Join(dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`print("hello")`), os.FileMode(0777))

	petsitter := &Petsitter{Stdout: stdout, Stderr: stderr}
	petsitter.ExecFile(file)
	defer os.RemoveAll(dir)

	out := stdout.String()
	if out != "hello\n" {
		t.Errorf("Expected 'hello'. Actual: %s", out)
	}
}

func TestPrintFail(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	dir, _ := ioutil.TempDir("", t.Name())
	file := filepath.Join(dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`print(hello)`), os.FileMode(0777))
	defer os.RemoveAll(dir)

	petsitter := &Petsitter{Stdout: stdout, Stderr: stderr}
	err := petsitter.ExecFile(file)
	out := stdout.String()
	if !(out == "" && strings.Contains(err.Error(), "undefined: hello")) {
		t.Errorf("Expected 'hello'. Actual: %s. Err: %s", out, err)
	}
}

func TestRun(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	dir, _ := ioutil.TempDir("", t.Name())
	file := filepath.Join(dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`run("echo meow")`), os.FileMode(0777))

	petsitter := &Petsitter{Stdout: stdout, Stderr: stderr}
	petsitter.ExecFile(file)
	defer os.RemoveAll(dir)

	out := stdout.String()
	if out != "meow\n" {
		t.Errorf("Expected 'meow'. Actual: %s", out)
	}
}
