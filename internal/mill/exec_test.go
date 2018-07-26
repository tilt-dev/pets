package mill

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/windmilleng/pets/internal/proc"
	"github.com/windmilleng/wmclient/pkg/dirs"
)

func TestPrint(t *testing.T) {
	dir, _ := ioutil.TempDir("", t.Name())
	file := filepath.Join(dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`print("hello")`), os.FileMode(0777))

	f := newPetFixture(t)
	petsitter, stdout := f.petsitter, f.stdout
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

	f := newPetFixture(t)
	petsitter, stdout := f.petsitter, f.stdout
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

	f := newPetFixture(t)
	petsitter, stdout := f.petsitter, f.stdout
	petsitter.ExecFile(file)
	defer os.RemoveAll(dir)

	out := stdout.String()
	if out != "meow\n" {
		t.Errorf("Expected 'meow'. Actual: %s", out)
	}
}

func TestStart(t *testing.T) {
	dir, _ := ioutil.TempDir("", t.Name())
	file := filepath.Join(dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`print(start("sleep 10"))`), os.FileMode(0777))

	f := newPetFixture(t)
	petsitter, stdout := f.petsitter, f.stdout
	petsitter.ExecFile(file)
	defer os.RemoveAll(dir)

	out := stdout.String()
	if !strings.Contains(out, "pid") {
		t.Errorf("Expected 'meow'. Actual: %s", out)
	}
}

func TestLoadGoGet(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	petsitter := &Petsitter{Stdout: stdout, Stderr: stderr}
	dir, _ := ioutil.TempDir("", t.Name())
	file := filepath.Join(dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`
load("go-get://github.com/windmilleng/blorg-frontend", blorg_fe_dir="dir")
print(blorg_fe_dir)
`), os.FileMode(0777))
	err := petsitter.ExecFile(file)
	defer os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}

	out := stdout.String()
	if !strings.Contains(out, "github.com/windmilleng/blorg-frontend") {
		t.Errorf("Expected import 'blorg-frontend'. Actual: %s", out)
	}
}

type petFixture struct {
	petsitter *Petsitter
	stdout    *bytes.Buffer
	stderr    *bytes.Buffer
	dir       string
}

func newPetFixture(t *testing.T) *petFixture {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	dir, _ := ioutil.TempDir("", t.Name())
	wmDir := dirs.NewWindmillDirAt(dir)
	procfs, err := proc.NewProcFSWithDir(wmDir)
	if err != nil {
		t.Fatal(err)
	}
	runner := proc.NewRunner(procfs)
	return &petFixture{
		petsitter: &Petsitter{
			Stdout: stdout,
			Stderr: stderr,
			Runner: runner,
		},
		stdout: stdout,
		stderr: stderr,
		dir:    dir,
	}
}

func (f *petFixture) tearDown() {
	os.RemoveAll(f.dir)
}
