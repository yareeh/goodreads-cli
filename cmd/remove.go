package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yareeh/goodreads-cli/internal"
)

var removeCmd = &cobra.Command{
	Use:   "remove <book-id>",
	Short: "Remove a book from your shelves",
	Long: "Take a book off all of your Goodreads shelves (the \"Remove from my shelf\" option).\n" +
		"A book that is on no shelf is left alone and the command succeeds.",
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

		fmt.Printf("Removing book %s from your shelves...\n", bookID)
		removed, err := internal.RemoveFromShelf(browser, bookID)
		if err != nil {
			return err
		}
		if removed {
			fmt.Println("Done!")
		} else {
			fmt.Println("Book is not on any shelf; nothing to remove.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
