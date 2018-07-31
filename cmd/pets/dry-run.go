package pets

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/windmilleng/pets/internal/mill"
	"github.com/windmilleng/pets/internal/proc"
)

var DryRunCmd = &cobra.Command{
	Use: "dry-run",
	Run: func(cmd *cobra.Command, args []string) {
		file := mill.GetFilePath()

		procfs, err := proc.NewProcFS()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		runner := proc.NewRunner(procfs)
		petsitter := &mill.Petsitter{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Runner: runner,
			Procfs: procfs,
		}
		err = petsitter.ExecFile(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
