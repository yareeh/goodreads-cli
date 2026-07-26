package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yareeh/goodreads-cli/internal"
)

var shelfName string

var shelfCmd = &cobra.Command{
	Use:   "shelf <book-id>",
	Short: "Add a book to a shelf",
	Long: `Add a book to a Goodreads shelf.

The three exclusive reading-status shelves (want-to-read, currently-reading,
and read) are selected through Goodreads' shelf picker. Other names are
treated as non-exclusive custom shelves. Every successful write is verified
by re-listing the target shelf before this command prints Done!.`,
	Args: cobra.ExactArgs(1),
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

		fmt.Println("Done! Verified by re-listing the target shelf.")
		return nil
	},
}

func init() {
	shelfCmd.Flags().StringVar(&shelfName, "shelf", "want-to-read", "exclusive or custom shelf name")
	rootCmd.AddCommand(shelfCmd)
}
