package cmd

import (
	"fmt"

	"github.com/jari/goodreads-cli/internal"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove saved session",
	Long:  "Remove the saved Goodreads session, requiring a fresh login next time.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := internal.Logout(); err != nil {
			return err
		}
		fmt.Println("Logged out. Session removed.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
