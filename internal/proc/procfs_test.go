package proc

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/windmilleng/pets/internal/service"
	"github.com/windmilleng/wmclient/pkg/dirs"
)

func TestProcFS(t *testing.T) {
	f := newProcFixture(t)
	defer f.tearDown()

	procfs := f.newProcFS()
	proc := PetsProc{Pid: 12345}
	err := procfs.AddProc(proc)
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"Pid":12345,"StartTime":"0001-01-01T00:00:00Z","TimeSince":0}
`
	f.assertProcFile(expected)

	err = procfs.RemoveProc(proc)
	if err != nil {
		t.Fatal(err)
	}
	f.assertProcFile("")
}

func TestProcFSDoubleAdd(t *testing.T) {
	f := newProcFixture(t)
	defer f.tearDown()

	procfs := f.newProcFS()
	proc := PetsProc{Pid: 12345}
	err := procfs.AddProc(proc)
	if err != nil {
		t.Fatal(err)
	}

	err = procfs.AddProc(proc)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected 'already exists' error. Actual: %v", err)
	}
}

func TestProcFSHost(t *testing.T) {
	f := newProcFixture(t)
	defer f.tearDown()

	procfs := f.newProcFS()
	proc := PetsProc{Pid: 12345}
	err := procfs.AddProc(proc)
	if err != nil {
		t.Fatal(err)
	}

	procfs.ModifyProc(proc.WithExposedHost("localhost", 8080))

	expected := `{"Pid":12345,"StartTime":"0001-01-01T00:00:00Z","TimeSince":0,"Hostname":"localhost","Port":8080}
`
	f.assertProcFile(expected)
}

func TestProcFSKey(t *testing.T) {
	f := newProcFixture(t)
	defer f.tearDown()

	procfs := f.newProcFS()
	proc := PetsProc{Pid: 12345}
	err := procfs.AddProc(proc)
	if err != nil {
		t.Fatal(err)
	}

	procfs.ModifyProc(proc.WithServiceKey(service.NewKey("frontend", "local")))

	expected := `{"DisplayName":"frontend-local","Pid":12345,"StartTime":"0001-01-01T00:00:00Z","TimeSince":9223372036854775807,"ServiceName":"frontend","ServiceTier":"local"}
`
	f.assertProcFile(expected)
}

func TestProcFSRemoveDead(t *testing.T) {
	f := newProcFixture(t)
	defer f.tearDown()

	procfs := f.newProcFS()
	cmd1 := exec.Command("echo", "1")
	cmd1.Start()

	cmd2 := exec.Command("sleep", "1000")
	cmd2.Start()
	defer cmd2.Process.Kill()

	procfs.AddProc(PetsProc{Pid: cmd1.Process.Pid})
	procfs.AddProc(PetsProc{Pid: cmd2.Process.Pid})

	cmd1.Wait()

	expected := fmt.Sprintf(`{"Pid":%d,"StartTime":"0001-01-01T00:00:00Z","TimeSince":0}
{"Pid":%d,"StartTime":"0001-01-01T00:00:00Z","TimeSince":0}
`, cmd1.Process.Pid, cmd2.Process.Pid)
	f.assertProcFile(expected)

	err := procfs.RemoveDeadProcs()
	if err != nil {
		t.Fatal(err)
	}

	expected = fmt.Sprintf(`{"Pid":%d,"StartTime":"0001-01-01T00:00:00Z","TimeSince":0}
`, cmd2.Process.Pid)
	f.assertProcFile(expected)
}

type procFixture struct {
	t   *testing.T
	dir string
}

func newProcFixture(t *testing.T) *procFixture {
	dir, _ := ioutil.TempDir("", t.Name())
	return &procFixture{
		t:   t,
		dir: dir,
	}
}

func (f *procFixture) tearDown() {
	os.RemoveAll(f.dir)
}

func (f *procFixture) newProcFS() ProcFS {
	wmDir := dirs.NewWindmillDirAt(f.dir)
	procfs, err := NewProcFSWithDir(wmDir)
	if err != nil {
		f.t.Fatal(err)
	}
	return procfs
}

func (f *procFixture) procFile() string {
	return filepath.Join(f.dir, procPath)
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
