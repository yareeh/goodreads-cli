package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "goodreads",
	Short: "A CLI for interacting with Goodreads",
	Long:  "goodreads-cli lets you search books, manage shelves, track reading progress, and post to discussions — all from the command line.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
