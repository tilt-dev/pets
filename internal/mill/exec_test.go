package mill

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/windmilleng/pets/internal/proc"
	"github.com/windmilleng/pets/internal/school"
	"github.com/windmilleng/wmclient/pkg/dirs"
)

func TestPrint(t *testing.T) {
	f := newPetFixture(t)
	defer f.tearDown()

	petsitter, stdout := f.petsitter, f.stdout
	file := filepath.Join(f.dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`print("hello")`), os.FileMode(0777))
	err := petsitter.ExecFile(file)
	if err != nil {
		t.Fatal(err)
	}

	out := stdout.String()
	if out != "hello\n" {
		t.Errorf("Expected 'hello'. Actual: %s", out)
	}
}

func TestPrintFail(t *testing.T) {
	f := newPetFixture(t)
	defer f.tearDown()

	petsitter, stdout := f.petsitter, f.stdout
	file := filepath.Join(f.dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`print(hello)`), os.FileMode(0777))

	err := petsitter.ExecFile(file)
	out := stdout.String()
	if !(out == "" && strings.Contains(err.Error(), "undefined: hello")) {
		t.Errorf("Expected 'hello'. Actual: %s. Err: %s", out, err)
	}
}

func TestRun(t *testing.T) {
	f := newPetFixture(t)
	defer f.tearDown()

	petsitter, stdout := f.petsitter, f.stdout
	file := filepath.Join(f.dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`run("echo meow")`), os.FileMode(0777))
	err := petsitter.ExecFile(file)
	if err != nil {
		t.Fatal(err)
	}

	out := stdout.String()
	if out != "meow\n" {
		t.Errorf("Expected 'meow'. Actual: %s", out)
	}
}

func TestStart(t *testing.T) {
	f := newPetFixture(t)
	defer f.tearDown()

	petsitter, stdout := f.petsitter, f.stdout
	file := filepath.Join(f.dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`print(start("sleep 10"))`), os.FileMode(0777))
	petsitter.ExecFile(file)

	out := stdout.String()
	if !strings.Contains(out, "pid") {
		t.Errorf("Expected 'meow'. Actual: %s", out)
	}
}

func TestLoadGoGet(t *testing.T) {
	f := newPetFixture(t)
	defer f.tearDown()

	petsitter, stdout := f.petsitter, f.stdout
	file := filepath.Join(f.dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`
load("go-get://github.com/windmilleng/blorg-frontend", blorg_fe_dir="dir")
print(blorg_fe_dir)
`), os.FileMode(0777))
	err := petsitter.ExecFile(file)
	if err != nil {
		t.Fatal(err)
	}

	out := stdout.String()
	if !strings.Contains(out, "github.com/windmilleng/blorg-frontend") {
		t.Errorf("Expected import 'blorg-frontend'. Actual: %s", out)
	}
}

func TestLoadRelative(t *testing.T) {
	f := newPetFixture(t)
	defer f.tearDown()

	petsitter, stdout := f.petsitter, f.stdout

	file := filepath.Join(f.dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`
load("inner", "random_number")
print(random_number())
`), os.FileMode(0777))

	innerFile := filepath.Join(f.dir, "inner", "Petsfile")
	os.MkdirAll(filepath.Dir(innerFile), os.FileMode(0777))
	ioutil.WriteFile(innerFile, []byte(`
def random_number():
  return 4
`), os.FileMode(0777))

	err := petsitter.ExecFile(file)
	if err != nil {
		t.Fatal(err)
	}

	out := stdout.String()
	if out != "4\n" {
		t.Errorf("Expected print '4'. Actual: %s", out)
	}
}

type petFixture struct {
	t         *testing.T
	petsitter *Petsitter
	stdout    *bytes.Buffer
	stderr    *bytes.Buffer
	dir       string
	procfs    proc.ProcFS
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
	school := school.NewPetSchool(procfs)
	return &petFixture{
		t:         t,
		petsitter: NewPetsitter(stdout, stderr, runner, procfs, school),
		stdout:    stdout,
		stderr:    stderr,
		dir:       dir,
		procfs:    procfs,
	}
}

func (f *petFixture) tearDown() {
	os.RemoveAll(f.dir)
}
