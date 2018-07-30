package proc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/windmilleng/wmclient/pkg/dirs"
)

const procPath = "pets/proc.json"

// Saves state about the currently running processes to the filesystem.
type ProcFS struct {
	wmDir *dirs.WindmillDir
	mu    *sync.Mutex
}

func NewProcFS() (ProcFS, error) {
	wmDir, err := dirs.UseWindmillDir()
	if err != nil {
		return ProcFS{}, fmt.Errorf("NewProcFS: %v", err)
	}
	return NewProcFSWithDir(wmDir)
}

func NewProcFSWithDir(wmDir *dirs.WindmillDir) (ProcFS, error) {
	fs := ProcFS{
		wmDir: wmDir,
		mu:    &sync.Mutex{},
	}
	err := fs.RemoveDeadProcs()
	if err != nil {
		return ProcFS{}, fmt.Errorf("NewProcFS: %v", err)
	}
	return fs, nil
}

// Add a proc to the JSON file
func (f ProcFS) AddProc(proc PetsProc) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	procs, err := f.procsFromFS()
	if err != nil {
		return err
	}

	procs = append(procs, proc)
	return f.procsToFS(procs)
}

// Remove a proc from the JSON file
func (f ProcFS) RemoveProc(proc PetsProc) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.filterProcs(func(p PetsProc) bool {
		return proc.Pid == p.Pid
	})
}

// Replace a proc in the JSON file matching the given proc's PID
func (f ProcFS) ModifyProc(proc PetsProc) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.mapProcs(func(p PetsProc) PetsProc {
		if proc.Pid == p.Pid {
			return proc
		}
		return p
	})
}

// Remove all dead proc from the JSON file
func (f ProcFS) RemoveDeadProcs() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.filterProcs(func(p PetsProc) bool {
		return !isAlive(p.Pid)
	})
}

// Remove a proc from the JSON file if it matches the filter.
func (f ProcFS) filterProcs(filter func(PetsProc) bool) error {
	return f.mapProcs(func(p PetsProc) PetsProc {
		if filter(p) {
			// Remove the process
			return PetsProc{}
		}
		return p
	})
}

// Map a proc from the JSON file to a new proc. If the new proc has Pid 0, remove it.
func (f ProcFS) mapProcs(mapFn func(PetsProc) PetsProc) error {
	procs, err := f.procsFromFS()
	if err != nil {
		return err
	}

	newProcs := []PetsProc{}
	for _, p := range procs {
		newP := mapFn(p)
		if newP.Pid != 0 {
			newProcs = append(newProcs, newP)
		}
	}

	return f.procsToFS(newProcs)
}

// Read all the procs from the JSON file
func (f ProcFS) ProcsFromFS() ([]PetsProc, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.procsFromFS()
}

// Read all the procs from the JSON file
func (f ProcFS) procsFromFS() ([]PetsProc, error) {
	contents, err := f.wmDir.ReadFile(procPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return procsFromReader(bytes.NewBufferString(contents))
}

// Write all the procs to the JSON file
func (f ProcFS) procsToFS(procs []PetsProc) error {
	out := &bytes.Buffer{}
	err := procsToWriter(out, procs)
	if err != nil {
		return err
	}

	return f.wmDir.WriteFile(procPath, out.String())
}

// Read a list of procs from any Reader
func procsFromReader(r io.Reader) ([]PetsProc, error) {
	result := []PetsProc{}
	decoder := json.NewDecoder(r)
	for decoder.More() {
		proc := PetsProc{}
		err := decoder.Decode(&proc)
		if err != nil {
			return nil, err
		}
		result = append(result, proc)
	}
	return result, nil
}

// Write a list of procs to any Writer
func procsToWriter(w io.Writer, procs []PetsProc) error {
	encoder := json.NewEncoder(w)
	for _, proc := range procs {
		err := encoder.Encode(proc)
		if err != nil {
			return err
		}
	}
	return nil
}
