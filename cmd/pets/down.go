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

			// Pets starts all processes with a process group. -p.Pid is a posix trick
			// to kill all processes in the group. This is helpful for things like 'go run'
			// that spawn subprocesses, so that the subprocesses get killed too.
			pgid := -p.Pid
			err := syscall.Kill(pgid, syscall.SIGINT)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	},
}
