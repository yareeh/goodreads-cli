package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var shelfName string

var shelfCmd = &cobra.Command{
	Use:   "shelf <book>",
	Short: "Add a book to a shelf",
	Long:  "Add a book to a Goodreads shelf (currently-reading, want-to-read, read)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		book := args[0]
		// TODO: Implement after recording shelf endpoints
		fmt.Printf("Adding '%s' to shelf '%s'\n", book, shelfName)
		fmt.Println("(Not yet implemented — use the recorder to capture shelf request patterns)")
		return nil
	},
}

func init() {
	shelfCmd.Flags().StringVar(&shelfName, "shelf", "want-to-read", "shelf name (currently-reading, want-to-read, read)")
	rootCmd.AddCommand(shelfCmd)
}
