package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for books on Goodreads",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		// TODO: Implement actual search after recording Goodreads search endpoints
		fmt.Printf("Searching for: %s\n", query)
		fmt.Println("(Not yet implemented — use the recorder to capture search request patterns)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
