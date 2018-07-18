package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var DryRunCmd = &cobra.Command{
	Use: "dry-run",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello Mars :)")
	},
}
