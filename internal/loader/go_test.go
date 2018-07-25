package loader

import (
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadGoRepo(t *testing.T) {
	f := newGoFixture(t)
	defer f.tearDown()

	dir, err := LoadGoRepo("github.com/windmilleng/blorg-frontend", f.buildCtx())
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(dir, f.dir) {
		t.Fatalf("Expected Go repo downloaded inside test tempdir. Actual: %s", dir)
	}
}

func TestLoadGoRepoFails(t *testing.T) {
	f := newGoFixture(t)
	defer f.tearDown()

	_, err := LoadGoRepo("github.com/windmilleng/blorg-nonsense", f.buildCtx())
	if err == nil || !strings.Contains(err.Error(), "failed with output") {
		t.Errorf("Expected error with exit status. Actual: %v", err)
	}
}

type goFixture struct {
	t   *testing.T
	dir string
}

func newGoFixture(t *testing.T) *goFixture {
	dir, _ := ioutil.TempDir("", t.Name())
	return &goFixture{
		t:   t,
		dir: dir,
	}
}

func (f *goFixture) buildCtx() build.Context {
	buildCtx := build.Default
	buildCtx.GOPATH = f.dir
	return buildCtx
}

func (f *goFixture) tearDown() {
	os.RemoveAll(f.dir)
}
