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
	"github.com/windmilleng/pets/internal/service"
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
		"run":      skylark.NewBuiltin("run", p.run),
		"start":    skylark.NewBuiltin("start", p.start),
		"service":  skylark.NewBuiltin("service", p.service),
		"register": skylark.NewBuiltin("register", p.register),
	}
}

func (p *Petsitter) run(t *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var cmdV skylark.Value

	if err := skylark.UnpackArgs(fn.Name(), args, kwargs,
		"cmdV", &cmdV,
	); err != nil {
		return nil, err
	}

	cmdArgs, err := argToCmd(fn, cmdV)
	if err != nil {
		return nil, err
	}

	cwd, err := p.wd(t)
	if err != nil {
		return nil, err
	}

	if err := p.Runner.RunWithIO(cmdArgs, cwd, p.Stdout, p.Stderr); err != nil {
		return nil, err
	}

	return skylark.None, nil
}

func (p *Petsitter) start(t *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var cmdV skylark.Value
	var process proc.PetsCommand

	if err := skylark.UnpackArgs(fn.Name(), args, kwargs,
		"cmdV", &cmdV,
	); err != nil {
		return nil, err
	}

	cmdArgs, err := argToCmd(fn, cmdV)
	if err != nil {
		return nil, err
	}

	cwd, err := p.wd(t)
	if err != nil {
		return nil, err
	}

	if process, err = p.Runner.StartWithIO(cmdArgs, cwd, p.Stdout, p.Stderr); err != nil {
		return nil, err
	}
	return petsProcToSkylarkValue(process.Proc), nil
}

func (p *Petsitter) register(t *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var name string
	var tier string
	var providerV *skylark.Function
	var depsV *skylark.List

	err := skylark.UnpackArgs(fn.Name(), args, kwargs,
		"name", &name,
		"tier", &tier,
		"provider", &providerV,
		"deps?", &depsV,
	)
	if err != nil {
		return nil, err
	}

	deps := []service.Name{}
	if depsV != nil {
		for i := 0; i < depsV.Len(); i++ {
			depV := depsV.Index(i)
			dep, ok := depV.(skylark.String)
			if !ok {
				return nil, fmt.Errorf("%s: deps must be a list of strings, got %T (%s)", fn.Name(), dep, dep)
			}
			deps = append(deps, service.Name(dep))
		}
	}

	if len(deps) < providerV.NumParams() {
		return nil, fmt.Errorf("%s: provider %q has %d parameters, but only %d deps listed",
			fn.Name(), providerV, providerV.NumParams(), len(deps))
	}

	provider := school.Provider(func(args []proc.PetsProc) (proc.PetsProc, error) {
		args = args[0:providerV.NumParams()]
		argsV := make([]skylark.Value, providerV.NumParams())
		for i, arg := range args {
			argsV[i] = petsProcToSkylarkValue(arg)
		}
		result, err := providerV.Call(p.newThread(), argsV, nil)
		if err != nil {
			return proc.PetsProc{}, err
		}

		return p.skylarkValueToPetsProc(result)
	})

	pos := t.TopFrame().Position()
	err = p.School.AddProvider(service.Key{
		Name: service.Name(name),
		Tier: service.Tier(tier),
	}, provider, deps, fmt.Sprintf("%s:%d", pos.Filename(), pos.Line))
	if err != nil {
		return nil, fmt.Errorf("%s: %v", fn.Name(), err)
	}

	return skylark.None, nil
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

// service(server, “localhost”, 8081)
func (p *Petsitter) service(t *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var server *skylark.Dict
	var host string
	var port int

	if err := skylark.UnpackArgs(fn.Name(), args, kwargs, "server", &server, "host", &host, "port", &port); err != nil {
		return nil, err
	}

	pr, err := p.skylarkValueToPetsProc(server)
	if err != nil {
		return nil, err
	}

	err = p.Procfs.ModifyProc(pr.WithExposedHost(host, port))
	if err != nil {
		return nil, err
	}

	return server, nil
}

// Get the working directory of the current skylark thread.
func (p *Petsitter) wd(t *skylark.Thread) (string, error) {
	frame := t.TopFrame()
	for frame.Position().Filename() == "<builtin>" {
		frame = frame.Parent()
	}
	file := frame.Position().Filename()
	if file == "" {
		return "", fmt.Errorf("Could not get the working directory for the current frame")
	}
	return filepath.Dir(file), nil
}

func (p *Petsitter) skylarkValueToPetsProc(v skylark.Value) (proc.PetsProc, error) {
	dict, ok := v.(*skylark.Dict)
	if !ok {
		return proc.PetsProc{}, fmt.Errorf("Not a valid pets process: %s", v)
	}

	skylarkPid, found, err := dict.Get(skylark.String("pid"))
	if !found {
		return proc.PetsProc{}, fmt.Errorf("Not a valid pets process: %s", v)
	}

	pkey, err := skylark.NumberToInt(skylarkPid)
	if err != nil {
		return proc.PetsProc{}, err
	}

	pid64, _ := pkey.Int64()
	pid := int(pid64)

	// from the pid, get the process
	procs, err := p.Procfs.ProcsFromFS()
	if err != nil {
		return proc.PetsProc{}, err
	}

	for _, p := range procs {
		// find when pid == proc
		if p.Pid == pid {
			return p, nil
		}
	}

	return proc.PetsProc{}, fmt.Errorf("Pets process missing: %s", v)
}

func petsProcToSkylarkValue(p proc.PetsProc) skylark.Value {
	pr := p.Pid
	d := &skylark.Dict{}
	pid := skylark.String("pid")
	proc := skylark.MakeInt(pr)
	d.Set(pid, proc)
	if p.Hostname != "" {
		d.Set(skylark.String("hostname"), skylark.String(p.Hostname))
	}
	if p.Port != 0 {
		d.Set(skylark.String("port"), skylark.MakeInt(p.Port))
	}
	return d
}

func argToCmd(b *skylark.Builtin, argV skylark.Value) ([]string, error) {
	switch argV := argV.(type) {
	case skylark.String:
		return []string{"bash", "-c", string(argV)}, nil
	default:
		return nil, fmt.Errorf("%v expects a string; got %T (%v)", b.Name(), argV, argV)
	}
}

func GetFilePath() string {
	cwd, _ := os.Getwd()

	return filepath.Join(cwd, Petsfile)
}
