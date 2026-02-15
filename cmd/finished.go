package cmd

import (
	"fmt"

	"github.com/jari/goodreads-cli/internal"
	"github.com/spf13/cobra"
)

var finishedCmd = &cobra.Command{
	Use:   "finished <book-id>",
	Short: "Mark a book as finished",
	Long:  "Mark a book as read/finished on Goodreads",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bookID := args[0]

		fmt.Println("Launching browser...")
		browser, err := internal.NewBrowser(true)
		if err != nil {
			return fmt.Errorf("launching browser: %w", err)
		}
		defer browser.Close()

		if !browser.IsLoggedIn() {
			return fmt.Errorf("not logged in — run 'goodreads login' first")
		}

		fmt.Printf("Marking book %s as read...\n", bookID)
		if err := internal.MarkRead(browser, bookID); err != nil {
			return err
		}

		fmt.Println("Done!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(finishedCmd)
}
