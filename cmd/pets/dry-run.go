package pets

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/windmilleng/pets/internal/mill"
)

var DryRunCmd = &cobra.Command{
	Use: "dry-run",
	Run: func(cmd *cobra.Command, args []string) {
		file := mill.GetFilePath()

		petsitter := &mill.Petsitter{Stdout: os.Stdout, Stderr: os.Stderr}
		err := petsitter.ExecFile(file)
		if err != nil {
			fmt.Println(err)
		}
	},
}
