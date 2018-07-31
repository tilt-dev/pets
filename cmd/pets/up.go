package pets

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/windmilleng/pets/internal/mill"
	"github.com/windmilleng/pets/internal/service"
)

var upTier string

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

	file := mill.GetFilePath()
	petsitter, err := newPetsitter()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = petsitter.ExecFile(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	school := petsitter.School
	if len(args) == 1 {
		_, err = school.UpByKey(service.Key{
			Name: service.Name(args[0]),
			Tier: service.Tier(upTier),
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		_, err = school.UpByTier(service.Tier(upTier))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func initUpCmd() {
	RootCmd.AddCommand(UpCmd)
	UpCmd.Run = runUpCmd
	UpCmd.Flags().StringVar(&upTier, "tier", "local", "The tier of servers to start up. Defaults to 'local'")
}
