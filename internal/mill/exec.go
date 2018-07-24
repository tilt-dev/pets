package mill

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/skylark"
	"github.com/windmilleng/pets/internal/proc"
)

type Probe struct {
	Stdout io.Writer
	Stderr io.Writer
}

// ExecFile takes a Petsfile and parses it using the Skylark interpreter
func (p *Probe) ExecFile(file string) error {
	thread := &skylark.Thread{
		Print: func(_ *skylark.Thread, msg string) { fmt.Fprintln(p.Stdout, msg) },
	}

	_, err := skylark.ExecFile(thread, file, nil, p.builtins())
	return err
}

func (p *Probe) builtins() skylark.StringDict {
	return skylark.StringDict{
		"run": skylark.NewBuiltin("run", p.run),
	}
}

func (p *Probe) run(t *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var cmdV skylark.Value

	if err := skylark.UnpackArgs("cmdV", args, kwargs,
		"cmdV", &cmdV,
	); err != nil {
		return nil, err
	}

	cmdArgs, err := argToCmd(fn, cmdV)
	if err != nil {
		return nil, err
	}

	cwd, _ := os.Getwd()
	if err := proc.RunWithIO(cmdArgs, cwd, p.Stdout, p.Stderr); err != nil {
		return nil, err
	}

	return skylark.None, nil
}

func argToCmd(b *skylark.Builtin, argV skylark.Value) ([]string, error) {
	switch argV := argV.(type) {
	case skylark.String:
		return []string{"bash", "-c", string(argV)}, nil
	default:
		return nil, fmt.Errorf("%v expects a string or list of strings; got %T (%v)", b.Name(), argV, argV)
	}
}

func GetFilePath() string {
	const Petsfile = "Petsfile"
	cwd, _ := os.Getwd()

	return filepath.Join(cwd, Petsfile)
}
