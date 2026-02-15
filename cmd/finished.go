package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var finishedCmd = &cobra.Command{
	Use:   "finished <book>",
	Short: "Mark a book as finished",
	Long:  "Mark a book as read/finished on Goodreads",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		book := args[0]
		// TODO: Implement after recording the mark-as-read endpoint
		fmt.Printf("Marking '%s' as finished\n", book)
		fmt.Println("(Not yet implemented — use the recorder to capture request patterns)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(finishedCmd)
}
