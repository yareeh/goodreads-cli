package cmd

import (
	"fmt"

	"github.com/jari/goodreads-cli/internal"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Goodreads",
	Long:  "Log in to Goodreads using credentials from ~/.goodreads-cli.yaml via browser automation",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := internal.LoadConfig()
		if err != nil {
			return err
		}

		fmt.Println("Launching browser...")
		browser, err := internal.NewBrowser()
		if err != nil {
			return fmt.Errorf("launching browser: %w", err)
		}
		defer browser.Close()

		if browser.IsLoggedIn() {
			fmt.Println("Already logged in!")
			return nil
		}

		fmt.Println("Logging in to Goodreads...")
		if err := internal.Login(browser, cfg); err != nil {
			return err
		}

		fmt.Println("Login successful! Session saved.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
