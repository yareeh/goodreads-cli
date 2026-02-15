package cmd

import (
	"fmt"

	"github.com/jari/goodreads-cli/internal"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Goodreads",
	Long:  "Log in to Goodreads using credentials from ~/.goodreads-cli.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := internal.LoadConfig()
		if err != nil {
			return err
		}

		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		fmt.Println("Logging in to Goodreads...")
		if err := internal.Login(client, cfg); err != nil {
			return err
		}

		fmt.Println("Login successful! Session saved.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
