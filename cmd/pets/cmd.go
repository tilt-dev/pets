package pets

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var dryRun bool

// func main() {
// 	Execute()
// }

func init() {
	RootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "just print recommended commands, don't run them")
	RootCmd.AddCommand(ListCmd)
}

var RootCmd = &cobra.Command{
	Use:   "pets [arguments]",
	Short: "PETS makes it easy to manage lots of servers running on your machine that you want to keep a close eye on for local development.",
	Long: `A PETS file is like a Makefile for running servers and connecting them 
	to other servers. With PETS, we can switch back and forth quickly
	between servers running locally and servers running in the cloud.`,
	Run: pets,
}

func pets(cmd *cobra.Command, args []string) {
	if dryRun {
		fmt.Fprintln(os.Stderr, "pets dry-run")
	} else {
		fmt.Fprintln(os.Stderr, "You ran pets!")
	}
}

// func Execute() {
// 	if err := rootCmd.Execute(); err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
// }
