package cmd

import (
	"fmt"

	"github.com/jari/goodreads-cli/internal"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <book-id>",
	Short: "Start reading a new book",
	Long:  "Mark a book as currently reading on Goodreads",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bookID := args[0]

		fmt.Println("Launching browser...")
		browser, err := internal.NewBrowser(!noHeadless)
		if err != nil {
			return fmt.Errorf("launching browser: %w", err)
		}
		defer browser.Close()

		if !browser.IsLoggedIn() {
			return fmt.Errorf("not logged in — run 'goodreads login' first")
		}

		fmt.Printf("Marking book %s as currently reading...\n", bookID)
		if err := internal.MarkCurrentlyReading(browser, bookID); err != nil {
			return err
		}

		fmt.Println("Done!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
