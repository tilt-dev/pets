package proc

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/windmilleng/wmclient/pkg/dirs"
)

func TestProcFS(t *testing.T) {
	dir, _ := ioutil.TempDir("", t.Name())
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.RemoveAll(dir)
	defer os.Setenv("HOME", oldHome)

	procfs, err := NewProcFS()
	if err != nil {
		t.Fatal(err)
	}

	wmdir, _ := dirs.GetWindmillDir()
	file := filepath.Join(wmdir, procPath)
	proc := PetsProc{Pid: 12345}
	err = procfs.AddProc(proc)
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"Pid":12345,"StartTime":"0001-01-01T00:00:00Z"}
`
	assertProcFile(t, file, expected)

	err = procfs.RemoveProc(proc)
	if err != nil {
		t.Fatal(err)
	}
	assertProcFile(t, file, "")
}

func assertProcFile(t *testing.T, file string, expected string) {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	actual := string(contents)
	if expected != actual {
		t.Errorf("Expected contents: %s. Actual: %s", expected, actual)
	}
}
