package mill

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/skylark"
)

// ExecFile takes a Petsfile and parses it using the Skylark interpreter
func ExecFile(file string, stdout io.Writer) error {
	thread := &skylark.Thread{
		Print: func(_ *skylark.Thread, msg string) { fmt.Fprintln(stdout, msg) },
	}
	_, err := skylark.ExecFile(thread, file, nil, nil)
	return err
}

func GetFilePath() string {
	const Petsfile = "Petsfile"
	cwd, _ := os.Getwd()

	return filepath.Join(cwd, Petsfile)
}
