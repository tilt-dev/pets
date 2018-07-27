package pets

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windmilleng/pets/internal/proc"
)

var ListCmd = &cobra.Command{
	Use: "list",
	Run: func(cms *cobra.Command, args []string) {

		// TODO — Create a new procfs object
		procs, _ := proc.ProcFS.ProcsFromFS()

		for _, p := range procs {
			fmt.Printf("Process%s\n", p.DisplayName)
		}
	},
}
