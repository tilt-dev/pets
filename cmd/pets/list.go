package pets

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/windmilleng/pets/internal/proc"
)

var ListCmd = &cobra.Command{
	Use: "list",
	Run: func(cms *cobra.Command, args []string) {
		procfs, err := proc.NewProcFS()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		procs, err := procfs.ProcsFromFS()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if len(procs) == 0 {
			fmt.Println("No pets running")
			return
		}

		fmt.Printf("Process ID\tName\n")
		for _, p := range procs {
			fmt.Printf("%d\t%s\n", p.Pid, p.DisplayName)
		}
	},
}