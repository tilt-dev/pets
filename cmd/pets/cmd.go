package pets

import (
	"fmt"
	"os"

	"github.com/google/skylark"
	"github.com/spf13/cobra"
	"github.com/windmilleng/wmclient/pkg/analytics"
)

const petsAppName = "pets"

var analyticsService analytics.Analytics

var dryRun bool

func init() {
	RootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "just print recommended commands, don't run them")
	RootCmd.AddCommand(ListCmd)
	RootCmd.AddCommand(DownCmd)
	initLogsCmd()
	initUpCmd()
}

func Execute() error {
	var analyticsCmd *cobra.Command
	var err error

	analyticsService, analyticsCmd, err = analytics.Init(petsAppName)
	if err != nil {
		return err
	}

	status, err := analytics.OptStatus()
	if err != nil {
		return err
	}

	if status == analytics.OptDefault {
		fmt.Fprintf(os.Stderr, "Send anonymized usage data to Windmill [y/n]? ")

		var response string
		fmt.Scanln(&response)
		if response == "" || response[0] == 'y' || response[0] == 'Y' {
			analytics.SetOpt(analytics.OptIn)
			fmt.Fprintln(os.Stderr, "Thanks! Setting 'pets analytics opt in'")
		} else {
			analytics.SetOpt(analytics.OptOut)
			fmt.Fprintln(os.Stderr, "Thanks! Setting 'pets analytics opt out'")
		}

		fmt.Fprintln(os.Stderr, "You set can update your privacy preferences later with 'pets analytics'")
	}

	RootCmd.AddCommand(analyticsCmd)

	// NOTE(nick): uncomment this code to generate markdown usage
	//doc.GenMarkdownTree(RootCmd, "./docs")

	return RootCmd.Execute()
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

func fatal(err error) {
	evalErr, isEvalErr := err.(*skylark.EvalError)
	if isEvalErr {
		fmt.Println(evalErr.Backtrace())
	} else {
		fmt.Println(err)
	}
	os.Exit(1)
}
