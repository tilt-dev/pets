package pets

import (
	"fmt"
	"time"

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

		fmt.Printf("%-30s%-30s\n", "Name", "Age")
		for _, p := range procs {
			el := timeDur(p.TimeSince().Truncate(time.Minute))
			fmt.Printf("%-30s%-30s\n", p.DisplayName, el)
		}
	},
}

func timeDur(d time.Duration) string {
	if seconds := int(d.Seconds()); seconds < -1 {
		return fmt.Sprintf("<invalid>")
	} else if seconds < 0 {
		return fmt.Sprintf("0s")
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	} else if hours := int(d.Hours()); hours < 24 {
		return fmt.Sprintf("%dh", hours)
	} else if hours < 24*365 {
		return fmt.Sprintf("%dd", hours/24)
	}
	return fmt.Sprintf("%dy", int(d.Hours()/24/365))
}
