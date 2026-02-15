package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var page int

var updateCmd = &cobra.Command{
	Use:   "update <book>",
	Short: "Update reading progress",
	Long:  "Update reading progress for a book on Goodreads",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		book := args[0]
		// TODO: Implement after recording the update-progress endpoint
		fmt.Printf("Updating '%s' progress to page %d\n", book, page)
		fmt.Println("(Not yet implemented — use the recorder to capture request patterns)")
		return nil
	},
}

func init() {
	updateCmd.Flags().IntVar(&page, "page", 0, "current page number")
	_ = updateCmd.MarkFlagRequired("page")
	rootCmd.AddCommand(updateCmd)
}
