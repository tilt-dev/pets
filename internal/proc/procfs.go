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

	return ProcFS{
		wmDir: wmDir,
		mu:    &sync.Mutex{},
	}, nil
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

	procs, err := f.procsFromFS()
	if err != nil {
		return err
	}

	newProcs := []PetsProc{}
	for _, p := range procs {
		if p.Pid != proc.Pid {
			newProcs = append(newProcs, proc)
		}
	}

	return f.procsToFS(newProcs)
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
