package mill

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecFile(t *testing.T) {
	stdout := &bytes.Buffer{}
	dir, _ := ioutil.TempDir("", t.Name())
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "Millfile")
	ioutil.WriteFile(file, []byte(`print("hello")`), os.FileMode(0777))
	ExecFile(file, stdout)

	out := stdout.String()
	if out != "hello\n" {
		t.Errorf("Expected 'hello'. Actual: %s", out)
	}
}

func TestExecFileFail(t *testing.T) {
	stdout := &bytes.Buffer{}
	dir, _ := ioutil.TempDir("", t.Name())
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "Millfile")
	ioutil.WriteFile(file, []byte(`print(hello)`), os.FileMode(0777))

	err := ExecFile(file, stdout)
	out := stdout.String()
	if !(out == "" && strings.Contains(err.Error(), "undefined: hello")) {
		t.Errorf("Expected 'hello'. Actual: %s. Err: %s", out, err)
	}
}
