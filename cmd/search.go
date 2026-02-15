package cmd

import (
	"fmt"
	"strings"

	"github.com/jari/goodreads-cli/internal"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for books on Goodreads",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		books, err := client.Search(query)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(books) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		fmt.Printf("%-12s %-50s %s\n", "ID", "TITLE", "AUTHOR")
		fmt.Printf("%-12s %-50s %s\n", "---", "-----", "------")
		for _, b := range books {
			title := b.Title
			if len(title) > 48 {
				title = title[:45] + "..."
			}
			fmt.Printf("%-12s %-50s %s\n", b.ID, title, b.Author)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
