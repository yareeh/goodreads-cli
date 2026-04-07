package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yareeh/goodreads-cli/internal"
)

var searchJSONFlag bool

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

		if searchJSONFlag {
			data, err := json.MarshalIndent(books, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
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
	searchCmd.Flags().BoolVar(&searchJSONFlag, "json", false, "Output results as JSON")
}
