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
	dir, _ := ioutil.TempDir("", t.Name())
	file := filepath.Join(dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`print("hello")`), os.FileMode(0777))
	ExecFile(file, stdout)
	defer os.RemoveAll(dir)

	out := stdout.String()
	if out != "hello\n" {
		t.Errorf("Expected 'hello'. Actual: %s", out)
	}
}

func TestPrintFail(t *testing.T) {
	stdout := &bytes.Buffer{}
	dir, _ := ioutil.TempDir("", t.Name())
	file := filepath.Join(dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`print(hello)`), os.FileMode(0777))
	defer os.RemoveAll(dir)

	err := ExecFile(file, stdout)
	out := stdout.String()
	if !(out == "" && strings.Contains(err.Error(), "undefined: hello")) {
		t.Errorf("Expected 'hello'. Actual: %s. Err: %s", out, err)
	}
}
