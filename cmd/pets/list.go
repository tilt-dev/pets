package pets

import (
	"fmt"
	"os"
	"time"

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

		fmt.Printf("%-30s%-30s\n", "Name", "Age")
		for _, p := range procs {
			el := timeDur(p.TimeSince().Truncate(time.Minute))
			// d := p.TimeSince()
			fmt.Printf("%-30s%-30d\n", p.DisplayName, el)
		}
	},
}

// why does THIS work
func timeDur(d time.Duration) int {
	if seconds := int(d.Seconds()); seconds < -1 {
		return 0
	} else if seconds < 0 {
		return 0
	} else if seconds < 60 {
		return seconds
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return minutes
	} else if hours := int(d.Hours()); hours < 24 {
		return hours
	} else if hours < 24*365 {
		return hours / 24
	}
	return int(d.Hours() / 24 / 365)
}

//but THIS not
func timeDurString(d time.Duration) string {
	if seconds := int(d.Seconds()); seconds < -1 {
		return "<invalid>"
	} else if seconds < 0 {
		return "0s"
	} else if seconds < 60 {
		return string(seconds) + "s"
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return string(minutes) + "m"
	} else if hours := int(d.Hours()); hours < 24 {
		return string(hours) + "h"
	} else if hours < 24*365 {
		return string(hours/24) + "d"
	}
	return string(int(d.Hours() / 24 / 365))
}
