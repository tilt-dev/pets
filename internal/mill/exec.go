package mill

import (
	"fmt"
	"go/build"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/google/skylark"
	"github.com/windmilleng/pets/internal/loader"
	"github.com/windmilleng/pets/internal/proc"
	"github.com/windmilleng/pets/internal/school"
)

const Petsfile = "Petsfile"

type Petsitter struct {
	Stdout io.Writer
	Stderr io.Writer
	Runner proc.Runner
	Procfs proc.ProcFS
	School *school.PetSchool
}

func NewPetsitter(stdout, stderr io.Writer, runner proc.Runner, procfs proc.ProcFS, school *school.PetSchool) *Petsitter {
	return &Petsitter{
		Stdout: stdout,
		Stderr: stderr,
		Runner: runner,
		Procfs: procfs,
		School: school,
	}
}

// ExecFile takes a Petsfile and parses it using the Skylark interpreter
func (p *Petsitter) ExecFile(file string) error {
	thread := p.newThread()
	_, err := skylark.ExecFile(thread, file, nil, p.builtins())
	return err
}

func (p *Petsitter) newThread() *skylark.Thread {
	thread := &skylark.Thread{
		Print: func(_ *skylark.Thread, msg string) {
			fmt.Fprintln(p.Stdout, msg)
		},
		Load: p.load,
	}
	return thread
}

func (p *Petsitter) builtins() skylark.StringDict {
	return skylark.StringDict{
		"run":     skylark.NewBuiltin("run", p.run),
		"start":   skylark.NewBuiltin("start", p.start),
		"service": skylark.NewBuiltin("service", p.service),
	}
}

func (p *Petsitter) run(t *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
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
	if err := p.Runner.RunWithIO(cmdArgs, cwd, p.Stdout, p.Stderr); err != nil {
		return nil, err
	}

	return skylark.None, nil
}

func (p *Petsitter) start(t *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var cmdV skylark.Value
	var process proc.PetsCommand

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
	if process, err = p.Runner.StartWithIO(cmdArgs, cwd, p.Stdout, p.Stderr); err != nil {
		return nil, err
	}
	pr := process.Proc.Pid

	d := &skylark.Dict{}
	pid := skylark.String("pid")
	proc := skylark.MakeInt(pr)
	d.Set(pid, proc)
	return d, nil
}

func (p *Petsitter) load(t *skylark.Thread, module string) (skylark.StringDict, error) {
	url, err := url.Parse(module)
	if err != nil {
		return nil, err
	}

	switch url.Scheme {
	case "go-get":
		importPath := path.Join(url.Host, url.Path)
		if fmt.Sprintf("go-get://%s", importPath) != module {
			return nil, fmt.Errorf("go-get URLs may not contain query or fragment info")
		}

		// TODO(nick): Use the dir returned by LoadGoRepo to run PetsFile recursively
		dir, err := loader.LoadGoRepo(importPath, build.Default)
		if err != nil {
			return nil, fmt.Errorf("load: %v", err)
		}

		return p.execPetsFileAt(t, dir, true)
	case "":
		dir := filepath.Join(filepath.Dir(t.TopFrame().Position().Filename()), module)
		return p.execPetsFileAt(t, dir, false)
	default:
		return nil, fmt.Errorf("Unknown load() strategy: %s. Available load schemes: go-get", url.Scheme)
	}
}

func (p *Petsitter) execPetsFileAt(t *skylark.Thread, module string, isMissingOk bool) (skylark.StringDict, error) {
	result := map[string]skylark.Value{}
	result["dir"] = skylark.String(module)

	info, err := os.Stat(module)
	if err != nil {
		if os.IsNotExist(err) && isMissingOk {
			return skylark.StringDict(result), nil
		}
		return nil, err
	}

	// If the user tried to load a directory, check if that
	// directory has a Petsfile
	if info.Mode().IsDir() {
		module = path.Join(module, Petsfile)

		info, err = os.Stat(module)
		if err != nil {
			if os.IsNotExist(err) && isMissingOk {
				return skylark.StringDict(result), nil
			}
			return nil, err
		}
	}

	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("File %q should be a plaintext Petsfile", module)
	}

	// The most exciting part of the function is finally here! We have an executable
	// Petsfile, so run it and grab the globals.
	globals, err := skylark.ExecFile(t, module, nil, p.builtins())
	if err != nil {
		return nil, err
	}

	for key, val := range globals {
		result[key] = val
	}

	return skylark.StringDict(result), nil
}

// Service(server, “localhost”, 8081)
func (p *Petsitter) service(t *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var server skylark.Dict
	var host string
	var port int
	var pr proc.PetsProc
	var pkey skylark.Int
	var pid int

	if err := skylark.UnpackArgs("service", args, kwargs, "server", &server, "host", &host, "port", &port); err != nil {
		return nil, err
	}

	sPid, found, err := server.Get(skylark.String("pid"))
	if !found {
		return nil, err
	}
	pkey, err = skylark.NumberToInt(sPid)
	pid64, _ := pkey.Int64()
	pid = int(pid64)
	if err != nil {
		return nil, err
	}

	procs, err := p.Procfs.ProcsFromFS()
	if err != nil {
		return nil, err
	}
	for _, p := range procs {
		if p.Pid == pid {
			pr = p
		}
	}

	p.Procfs.ModifyProc(pr.WithExposedHost(host, port))

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
	cwd, _ := os.Getwd()

	return filepath.Join(cwd, Petsfile)
}
