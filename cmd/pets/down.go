package pets

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/windmilleng/pets/internal/proc"
)

var DownCmd = &cobra.Command{
	Use: "down",
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

		for _, p := range procs {
			// TODO: Decide what to do in edge cases, when killing pets doesn't work
			err := syscall.Kill(p.Pid, syscall.SIGINT)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	},
}
