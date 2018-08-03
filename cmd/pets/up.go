package pets

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/windmilleng/pets/internal/mill"
	"github.com/windmilleng/pets/internal/service"
)

var upTier string
var upOverrides []string

var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "Start servers specified in the Petsfile",
	Long: `Start servers specified in the Petsfile.

To start all servers with tier "local", run: 'pets up'

To start all servers on a different provider tier (for example, k8s), run: 'pets up --tier=k8s'

To start a single server and all its dependencies, run: 'pets up my-server'.
`,
	Example: `pets up
pets up frontend
pets up frontend --tier=k8s`,
}

func runUpCmd(cmd *cobra.Command, args []string) {
	if len(args) > 1 {
		UpCmd.Usage()

		fmt.Printf("\nToo many arguments: %+v\n", args)
		os.Exit(1)
	}

	overrideMap := make(map[service.Name]service.Tier)
	for _, override := range upOverrides {
		parts := strings.Split(override, "=")
		if len(parts) != 2 {
			fmt.Printf("--with flag should have format 'service=tier'. Actual value: %s\n", override)
			os.Exit(1)
		}

		overrideMap[service.Name(parts[0])] = service.Tier(parts[1])
	}

	analyticsService.Incr("cmd.up", nil)
	defer analyticsService.Flush(time.Second)

	file := mill.GetFilePath()
	petsitter, err := newPetsitter()
	if err != nil {
		fatal(err)
	}

	err = petsitter.ExecFile(file)
	if err != nil {
		fatal(err)
	}

	school := petsitter.School

	for name, tier := range overrideMap {
		err = school.AddOverride(name, tier)
		if err != nil {
			fatal(err)
		}
	}

	if len(args) == 1 {
		_, err = school.UpByKey(service.Key{
			Name: service.Name(args[0]),
			Tier: service.Tier(upTier),
		})
		if err != nil {
			fatal(err)
		}
	} else {
		_, err = school.UpByTier(service.Tier(upTier))
		if err != nil {
			fatal(err)
		}
	}
}

func initUpCmd() {
	RootCmd.AddCommand(UpCmd)
	UpCmd.Run = runUpCmd
	UpCmd.Flags().StringVar(&upTier, "tier", "local", "The tier of servers to start up. Defaults to 'local'")
	UpCmd.Flags().StringSliceVar(&upOverrides, "with", nil, "Override servers in the server graph. Example: --with=backend=k8s")
}
