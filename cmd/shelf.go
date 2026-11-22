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

The reading-status shelves (want-to-read, currently-reading, read) are
chosen in the book page's shelf dialog. Any other name must be one of your
own shelves, such as ebook or audiobook (matched case-insensitively); a name
you don't have is an error, never a new shelf. The command fails unless the
book is on the shelf afterwards.`,
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

		fmt.Println("Done!")
		return nil
	},
}

func init() {
	shelfCmd.Flags().StringVar(&shelfName, "shelf", "want-to-read", "want-to-read, currently-reading, read, or one of your own shelves")
	rootCmd.AddCommand(shelfCmd)
}
