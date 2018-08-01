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
			p.TimeSince = time.Since(p.StartTime)
			d := p.TimeSince.Truncate(time.Minute)
			if seconds := int(d.Seconds()); seconds < -1 {
				fmt.Printf("%-30s%s\n", p.DisplayName, "<invalid>")
				return
			} else if seconds < 0 {
				fmt.Printf("%-30s%s\n", p.DisplayName, "0s")
				return
			} else if seconds < 60 {
				fmt.Printf("%-30s%ds\n", p.DisplayName, seconds)
				return
			} else if minutes := int(d.Minutes()); minutes < 60 {
				fmt.Printf("%-30s%dm\n", p.DisplayName, minutes)
				return
			} else if hours := int(d.Hours()); hours < 24 {
				fmt.Printf("%-30s%dh\n", p.DisplayName, hours)
				return
			} else if hours < 24*365 {
				fmt.Printf("%-30s%dh\n", p.DisplayName, hours/24)
			}
			fmt.Printf("%-30s%dy\n", p.DisplayName, int(d.Hours()/24/365))
			return
		}
	},
}
