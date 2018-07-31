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

		petsitter, err := newPetsitter()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = petsitter.ExecFile(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
