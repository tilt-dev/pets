package pets

import (
	"os"

	"github.com/windmilleng/pets/internal/mill"
	"github.com/windmilleng/pets/internal/proc"
	"github.com/windmilleng/pets/internal/school"
)

func newPetsitter() (*mill.Petsitter, error) {
	procfs, err := proc.NewProcFS()
	if err != nil {
		return nil, err
	}
	runner := proc.NewRunner(procfs)
	school := school.NewPetSchool(procfs)
	return mill.NewPetsitter(os.Stdout, os.Stderr, runner, procfs, school, dryRun), nil
}
