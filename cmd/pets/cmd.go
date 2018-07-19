package pets

import (
	"fmt"

	"github.com/spf13/cobra"
)

// func main() {
// 	Execute()
// }

func init() {
	RootCmd.AddCommand(DryRunCmd)
	// 	 rootCmd.PersistentFlags().String("dry-run")
}

var RootCmd = &cobra.Command{
	Use:   "pets [arguments]",
	Short: "PETS makes it easy to manage lots of servers running on your machine that you want to keep a close eye on for local development.",
	Long: `A PETS file is like a Makefile for running servers and connecting them 
			to other servers. With PETS, we can switch back and forth quickly 
			between servers running locally and servers running in the cloud.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello :)")
	},
}

// func Execute() {
// 	if err := rootCmd.Execute(); err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
// }
