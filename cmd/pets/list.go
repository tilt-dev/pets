package pets

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windmilleng/pets/internal/proc"
)

var ListCmd = &cobra.Command{
	Use: "list",
	Run: func(cms *cobra.Command, args []string) {
		procfs, err := proc.NewProcFS()
		if err != nil {
			fatal(err)
		}

		procs, err := procfs.ProcsFromFS()
		if err != nil {
			fatal(err)
		}

		if len(procs) == 0 {
			fmt.Println("No pets running")
			return
		}

		fmt.Printf("%-20s%-20s\n", "Process ID", "Name")
		for _, p := range procs {
			fmt.Printf("%-20d%-20s\n", p.Pid, p.DisplayName)
		}
	},
}
