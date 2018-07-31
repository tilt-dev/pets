package school

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/windmilleng/pets/internal/proc"
	"github.com/windmilleng/pets/internal/service"
	"github.com/windmilleng/wmclient/pkg/dirs"
)

// Create a bunch of fake test data
const local = service.Tier("local")

const blorgFrontend = service.Name("blorg-frontend")
const blorgBackend = service.Name("blorg-backend")
const blorglyBackend = service.Name("blorgly-backend")
const cockroach = service.Name("cockroach")

func localKey(name service.Name) service.Key {
	return service.NewKey(name, local)
}

// Test a server topology with one service, blorg-frontend
func TestOneServer(t *testing.T) {
	f := newSchoolFixture(t)
	defer f.tearDown()

	key := localKey(blorgFrontend)
	err := f.school.AddProvider(key, f.makeProvider(1), nil, "")
	if err != nil {
		t.Fatal(err)
	}

	proc, err := f.school.UpByKey(key)
	if err != nil {
		t.Fatal(err)
	}

	if proc.Pid != 1 {
		t.Errorf("Unexpected proc: %v", proc)
	}

	services, err := f.school.healthyServices()
	if err != nil {
		t.Fatal(err)
	}

	_, containsFE := services[key]
	if len(services) != 1 || !containsFE {
		t.Fatalf("Unexpected healthy services: %+v", services)
	}
}

// Test a server topology with three services, blorg-frontend, blorg-backend, and blorgly-backend
func TestThreeServers(t *testing.T) {
	f := newSchoolFixture(t)
	defer f.tearDown()

	key := localKey(blorgFrontend)
	err := f.school.AddProvider(key, f.makeProvider(1), []service.Name{blorgBackend, blorglyBackend}, "")
	if err != nil {
		t.Fatal(err)
	}

	err = f.school.AddProvider(localKey(blorgBackend), f.makeProvider(2), nil, "")
	if err != nil {
		t.Fatal(err)
	}

	err = f.school.AddProvider(localKey(blorglyBackend), f.makeProvider(3), nil, "")
	if err != nil {
		t.Fatal(err)
	}

	proc, err := f.school.UpByKey(key)
	if err != nil {
		t.Fatal(err)
	}

	if proc.Pid != 1 {
		t.Errorf("Unexpected proc: %v", proc)
	}

	if len(f.procs) != 3 {
		t.Errorf("Expected 3 procs. Actual: %v", f.procs)
	}
}

// Test a server diamond with both backends depending on cockroach
func TestServerDiamond(t *testing.T) {
	f := newSchoolFixture(t)
	defer f.tearDown()

	key := localKey(blorgFrontend)
	f.setupDiamond()

	proc, err := f.school.UpByKey(key)
	if err != nil {
		t.Fatal(err)
	}

	if proc.Pid != 1 {
		t.Errorf("Unexpected proc: %v", proc)
	}

	if len(f.procs) != 4 {
		t.Errorf("Expected 4 procs. Actual: %v", f.procs)
	}
}

func TestServerDiamondByTier(t *testing.T) {
	f := newSchoolFixture(t)
	defer f.tearDown()

	f.setupDiamond()

	procs, err := f.school.UpByTier("local")
	if err != nil {
		t.Fatal(err)
	}
	if len(procs) != 4 {
		t.Errorf("Expected 4 procs returned. Actual: %v", procs)
	}
	if len(f.procs) != 4 {
		t.Errorf("Expected 4 procs in fixture. Actual: %v", f.procs)
	}
}

func TestMissingDependency(t *testing.T) {
	f := newSchoolFixture(t)
	defer f.tearDown()

	key := localKey(blorgFrontend)
	err := f.school.AddProvider(key, f.makeProvider(1), []service.Name{blorgBackend}, "")
	if err != nil {
		t.Fatal(err)
	}

	err = f.school.AddProvider(localKey(blorgBackend), f.makeProvider(2), []service.Name{cockroach}, "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.school.UpByKey(key)
	if err == nil ||
		!strings.Contains(err.Error(), "Service \"blorg-frontend\" depends on service \"blorg-backend\"") ||
		!strings.Contains(err.Error(), "Service \"blorg-backend\" depends on service \"cockroach\"") ||
		!strings.Contains(err.Error(), "No provider found for service \"cockroach\", tier \"local\"") {
		t.Fatal(err)
	}
}

type schoolFixture struct {
	t      *testing.T
	dir    string
	procfs proc.ProcFS
	school *PetSchool
	procs  []proc.PetsProc
}

func newSchoolFixture(t *testing.T) *schoolFixture {
	dir, _ := ioutil.TempDir("", t.Name())
	wmDir := dirs.NewWindmillDirAt(dir)
	procfs, _ := proc.NewProcFSWithDir(wmDir)
	school := NewPetSchool(procfs)
	return &schoolFixture{
		t:      t,
		dir:    dir,
		procfs: procfs,
		school: school,
	}
}

func (f *schoolFixture) setupDiamond() {
	key := localKey(blorgFrontend)
	err := f.school.AddProvider(key, f.makeProvider(1), []service.Name{blorgBackend, blorglyBackend}, "")
	if err != nil {
		f.t.Fatal(err)
	}

	err = f.school.AddProvider(localKey(blorgBackend), f.makeProvider(2), []service.Name{cockroach}, "")
	if err != nil {
		f.t.Fatal(err)
	}

	err = f.school.AddProvider(localKey(blorglyBackend), f.makeProvider(3), []service.Name{cockroach}, "")
	if err != nil {
		f.t.Fatal(err)
	}

	err = f.school.AddProvider(localKey(cockroach), f.makeProvider(4), nil, "")
	if err != nil {
		f.t.Fatal(err)
	}
}

func (f *schoolFixture) makeProvider(pid int) Provider {
	return Provider(func(inputs []proc.PetsProc) (proc.PetsProc, error) {
		p := proc.PetsProc{
			Pid:      pid,
			Hostname: "localhost",
			Port:     1000 + pid,
		}
		err := f.procfs.AddProc(p)
		if err != nil {
			return proc.PetsProc{}, err
		}
		f.procs = append(f.procs, p)
		return p, nil
	})
}

func (f *schoolFixture) tearDown() {
	os.RemoveAll(f.dir)
}
