package pets

import (
	"fmt"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/windmilleng/pets/internal/proc"
)

var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "Kill all processes started by pets",
	Run: func(cms *cobra.Command, args []string) {
		analyticsService.Incr("cmd.down", nil)
		defer analyticsService.Flush(time.Second)

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

		for _, p := range procs {
			if p.ServiceName != "" {
				fmt.Printf("Stopping %s\n", p.ServiceKey())
			} else {
				fmt.Printf("Stopping pid %d\n", p.Pid)
			}

			// TODO: Decide what to do in edge cases, when killing pets doesn't work
			// for now, ignore any errors.

			// Pets starts all processes with a process group. -p.Pid is a posix trick
			// to kill all processes in the group. This is helpful for things like 'go run'
			// that spawn subprocesses, so that the subprocesses get killed too.
			pgid := -p.Pid
			syscall.Kill(pgid, syscall.SIGINT)
		}

		procfs.RemoveAllProcs()
	},
}
