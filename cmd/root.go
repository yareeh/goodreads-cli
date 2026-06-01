package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yareeh/goodreads-cli/internal/version"
)

var noHeadless bool

var rootCmd = &cobra.Command{
	Use:     "goodreads",
	Short:   "A CLI for interacting with Goodreads",
	Long:    "goodreads-cli lets you search books, manage shelves, track reading progress, and post to discussions — all from the command line.",
	Version: version.Current(),
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noHeadless, "no-headless", false, "show the browser window for debugging")
}
