package pets

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windmilleng/pets/internal/mill"
)

var DryRunCmd = &cobra.Command{
	Use: "dry-run",
	Run: func(cmd *cobra.Command, args []string) {
		file := mill.GetFilePath()
		stdout := &bytes.Buffer{}

		mill.ExecFile(file, stdout)

		out := stdout.String()
		fmt.Printf("Petsfile says: %s", out)
	},
}
