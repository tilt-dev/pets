package main

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
	Use:   "hello",
	Short: "Hugo is a very fast static site generator",
	Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at http://hugo.spf13.com`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello world :)")
	},
}

// func Execute() {
// 	if err := rootCmd.Execute(); err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
// }
