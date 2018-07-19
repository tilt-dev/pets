package proc

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/windmilleng/wmclient/pkg/dirs"
)

func TestProcFS(t *testing.T) {
	f := newProcFixture(t)
	defer f.tearDown()

	procfs, err := NewProcFS()
	if err != nil {
		t.Fatal(err)
	}

	proc := PetsProc{Pid: 12345}
	err = procfs.AddProc(proc)
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"Pid":12345,"StartTime":"0001-01-01T00:00:00Z"}
`
	f.assertProcFile(expected)

	err = procfs.RemoveProc(proc)
	if err != nil {
		t.Fatal(err)
	}
	f.assertProcFile("")
}

func TestProcFSRemoveDead(t *testing.T) {
	f := newProcFixture(t)
	defer f.tearDown()

	procfs, err := NewProcFS()
	if err != nil {
		t.Fatal(err)
	}

	cmd1 := exec.Command("echo", "1")
	cmd1.Start()

	cmd2 := exec.Command("sleep", "1000")
	cmd2.Start()
	defer cmd2.Process.Kill()

	procfs.AddProc(PetsProc{Pid: cmd1.Process.Pid})
	procfs.AddProc(PetsProc{Pid: cmd2.Process.Pid})

	cmd1.Wait()

	expected := fmt.Sprintf(`{"Pid":%d,"StartTime":"0001-01-01T00:00:00Z"}
{"Pid":%d,"StartTime":"0001-01-01T00:00:00Z"}
`, cmd1.Process.Pid, cmd2.Process.Pid)
	f.assertProcFile(expected)

	err = procfs.RemoveDeadProcs()
	if err != nil {
		t.Fatal(err)
	}

	expected = fmt.Sprintf(`{"Pid":%d,"StartTime":"0001-01-01T00:00:00Z"}
`, cmd2.Process.Pid)
	f.assertProcFile(expected)
}

type procFixture struct {
	t       *testing.T
	oldHome string
	dir     string
}

func newProcFixture(t *testing.T) *procFixture {
	dir, _ := ioutil.TempDir("", t.Name())
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	return &procFixture{
		t:       t,
		oldHome: oldHome,
		dir:     dir,
	}
}

func (f *procFixture) tearDown() {
	os.RemoveAll(f.dir)
	os.Setenv("HOME", f.oldHome)
}

func (f *procFixture) procFile() string {
	wmdir, _ := dirs.GetWindmillDir()
	return filepath.Join(wmdir, procPath)
}

func (f *procFixture) assertProcFile(expected string) {
	file := f.procFile()
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		f.t.Fatal(err)
	}

	actual := string(contents)
	if expected != actual {
		f.t.Errorf("Expected contents: %s. Actual: %s", expected, actual)
	}
}
