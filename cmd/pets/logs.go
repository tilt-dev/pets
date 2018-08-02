package pets

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/windmilleng/pets/internal/proc"
	"github.com/windmilleng/pets/internal/service"
)

var LogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get logs for running servers",
	Example: `pets logs
pets logs frontend`,
}

func runLogsCmd(cmd *cobra.Command, args []string) {
	if len(args) > 1 {
		LogsCmd.Usage()

		fmt.Printf("\nToo many arguments: %+v\n", args)
		os.Exit(1)
	}

	procfs, err := proc.NewProcFS()
	if err != nil {
		fatal(err)
	}

	procs, err := procfs.ProcsFromFS()
	if err != nil {
		fatal(err)
	}

	// Right now, this only gets logs from running services. What about logs
	// from dead services?
	name := service.Name("")
	if len(args) == 1 {
		name = service.Name(args[0])
	}

	printed := false
	for _, p := range procs {
		if p.ServiceName == "" {
			continue
		}

		// If the user specified a particular service, skip any services that don't match.
		if name != "" && p.ServiceName != name {
			continue
		}

		contents, err := procfs.ReadLogFile(p.ServiceKey())
		if err != nil {
			fatal(err)
		}

		if name == "" {
			// If the user is printing logs for multiple services, print a header.
			if printed {
				fmt.Println("")
			}
			fmt.Printf(`--------------------------
PETS logs: %s-%s
--------------------------
`, p.ServiceName, p.ServiceTier)
		}

		fmt.Print(contents)
		printed = true
	}

	if !printed {
		if name == "" {
			fmt.Println("No running services")
		} else {
			fmt.Printf("No running services matching: %s\n", name)
		}
	}
}

func initLogsCmd() {
	LogsCmd.Run = runLogsCmd
	RootCmd.AddCommand(LogsCmd)
}
