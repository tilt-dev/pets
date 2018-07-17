package proc

import (
	"bytes"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cwd, _ := os.Getwd()
	err := RunWithIO([]string{"echo", "hello"}, cwd, stdout, stderr)
	if err != nil {
		t.Fatal(err)
	}

	out := stdout.String()
	if out != "hello\n" {
		t.Errorf("Expected 'hello'. Actual: %s", out)
	}
}
