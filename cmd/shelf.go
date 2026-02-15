package cmd

import (
	"fmt"

	"github.com/jari/goodreads-cli/internal"
	"github.com/spf13/cobra"
)

var shelfName string

var shelfCmd = &cobra.Command{
	Use:   "shelf <book-id>",
	Short: "Add a book to a shelf",
	Long:  "Add a book to a Goodreads shelf (currently-reading, want-to-read, read)",
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

		fmt.Printf("Adding book %s to shelf '%s'...\n", bookID, shelfName)
		if err := internal.AddToShelf(browser, bookID, shelfName); err != nil {
			return err
		}

		fmt.Println("Done!")
		return nil
	},
}

func init() {
	shelfCmd.Flags().StringVar(&shelfName, "shelf", "want-to-read", "shelf name (currently-reading, want-to-read, read)")
	rootCmd.AddCommand(shelfCmd)
}
