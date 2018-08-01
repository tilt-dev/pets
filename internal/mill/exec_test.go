package mill

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/windmilleng/pets/internal/proc"
	"github.com/windmilleng/pets/internal/school"
	"github.com/windmilleng/pets/internal/service"
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

func TestLoadRelativeWorkingDirectory(t *testing.T) {
	f := newPetFixture(t)
	defer f.tearDown()

	petsitter, stdout := f.petsitter, f.stdout

	file := filepath.Join(f.dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`
load("inner", "inner_pwd")
run("pwd")
inner_pwd()
`), os.FileMode(0777))

	innerFile := filepath.Join(f.dir, "inner", "Petsfile")
	os.MkdirAll(filepath.Dir(innerFile), os.FileMode(0777))
	ioutil.WriteFile(innerFile, []byte(`
def inner_pwd():
  run("pwd")
`), os.FileMode(0777))

	err := petsitter.ExecFile(file)
	if err != nil {
		t.Fatal(err)
	}

	out := stdout.String()
	expected := fmt.Sprintf("%s\n%s/inner\n", f.dir, f.dir)
	if out != expected {
		t.Errorf("Expected:\n%s\n\nActual:\n%s", expected, out)
	}
}

func TestRegister(t *testing.T) {
	f := newPetFixture(t)
	petsitter := f.petsitter

	file := filepath.Join(f.dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`
def start_local():
  result = service(start("sleep 100"), "localhost", 8080)
  print(result["host"])
  return result

register("blorg-frontend", "local", start_local)
`), os.FileMode(0777))

	err := petsitter.ExecFile(file)
	if err != nil {
		t.Fatal(err)
	}

	school := f.petsitter.School
	key := service.NewKey("blorg-frontend", "local")
	_, err = school.UpByKey(key)
	if err != nil {
		t.Fatal(err)
	}

	f.assertHasServiceKey(key)

	out := f.stdout.String()
	if !strings.Contains(out, "localhost:8080") {
		t.Errorf("Expected 'localhost:8080'. Actual: %s", out)
	}
}

func TestRegisterTwice(t *testing.T) {
	f := newPetFixture(t)
	petsitter := f.petsitter

	file := filepath.Join(f.dir, "Petsfile")
	ioutil.WriteFile(file, []byte(`
def start_local():
  return service(start("sleep 100"), "localhost", 8080)

register("blorg-frontend", "local", start_local)
register("blorg-frontend", "local", start_local)
`), os.FileMode(0777))

	err := petsitter.ExecFile(file)
	if err == nil ||
		!strings.Contains(err.Error(), "Duplicate provider") ||
		!strings.Contains(err.Error(), fmt.Sprintf("First:  %s/Petsfile:5", f.dir)) {
		t.Errorf("Expected duplicate provider error. Actual: %v", err)
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

func (f *petFixture) assertHasServiceKey(key service.Key) {
	procs, err := f.procfs.ProcsFromFS()
	if err != nil {
		f.t.Fatal(err)
	}

	for _, proc := range procs {
		if proc.ServiceName == key.Name && proc.ServiceTier == key.Tier {
			return
		}
	}

	f.t.Errorf("Service key not found in running service list: %+v", key)
}

func (f *petFixture) tearDown() {
	os.RemoveAll(f.dir)
}
