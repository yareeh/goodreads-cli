package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <book>",
	Short: "Start reading a new book",
	Long:  "Mark a book as currently reading on Goodreads",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		book := args[0]
		// TODO: Implement after recording the add-to-currently-reading endpoint
		fmt.Printf("Marking '%s' as currently reading\n", book)
		fmt.Println("(Not yet implemented — use the recorder to capture request patterns)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
