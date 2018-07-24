package loader

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
)

// Loads a Go repo and returns an absolute directory path.
//
// The Go repo may or may not have a pets file.
//
// Callers can inject their own GOROOT/GOPATH. The default build.Context is build.Default.
// https://golang.org/pkg/go/build/#Default
//
// TODO(nick): Long-term, we will need to have some way to give PETS more
// control over which version of the repo loads.
//
// I personally think that the MVS (minimal version selection) approach is promising,
// where each repo has a .petslock file that describes the minimum version of its
// dependencies, and the user can specify later versions.
//
// Right now we do the dumbest possible thing of doing a 'go get', which will get
// the repo if it doesn't exist and use the repo on disk if it does exist.
func LoadGoRepo(importPath string, buildCtx build.Context) (string, error) {
	cmd := exec.Command("go", "get", importPath)

	env := append([]string{}, os.Environ()...)
	env = append(env,
		fmt.Sprintf("GOROOT=%s", buildCtx.GOROOT),
		fmt.Sprintf("GOPATH=%s", buildCtx.GOPATH))
	cmd.Env = env

	err := cmd.Run()
	if err != nil {
		exitErr, isExit := err.(*exec.ExitError)
		if isExit {
			return "", fmt.Errorf("go get %q failed with output:\n%s\n", importPath, string(exitErr.Stderr))
		}
		return "", fmt.Errorf("go get %q failed: %v", importPath, err)
	}

	// The build package doesn't expose an easy function for getting the absolute
	// path to the directory, so we have to do this ourselves.
	srcDirs := buildCtx.SrcDirs()
	foundDir := ""
	for _, srcDir := range srcDirs {
		importDir := filepath.Join(srcDir, importPath)
		info, err := os.Stat(importDir)
		if err != nil || !info.IsDir() {
			// Try the next candidate.
			continue
		}

		foundDir = importDir
		break
	}

	if foundDir == "" {
		return "", fmt.Errorf("LoadGoRepo(%q): package not found in %v", importPath, srcDirs)
	}
	return foundDir, nil
}
