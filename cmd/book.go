package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yareeh/goodreads-cli/internal"
)

var bookJSONFlag bool

var bookCmd = &cobra.Command{
	Use:   "book <id>",
	Short: "Fetch full bibliographic details for a Goodreads book (ISBN, publisher, year, original title, …)",
	Long: `Fetch the full bibliographic record for a Goodreads book by its legacy ID.

Goodreads pages embed structured data (JSON-LD and a __NEXT_DATA__ Apollo state)
that carries ISBN-10/13, publisher, edition publication date, original title,
language, page count, and format. This command parses that data instead of
relying on LLM-driven scraping.

Example:
  goodreads book 18690730 --json
  goodreads book 18690730`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		book, err := client.FetchBookDetails(id)
		if err != nil {
			return fmt.Errorf("fetching book details: %w", err)
		}

		if bookJSONFlag {
			data, err := json.MarshalIndent(book, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("Title:          %s\n", book.Title)
		fmt.Printf("Author:         %s\n", book.Author)
		if book.OriginalTitle != "" && book.OriginalTitle != book.Title {
			fmt.Printf("Original title: %s\n", book.OriginalTitle)
		}
		if book.Year != "" {
			if book.Month != "" {
				fmt.Printf("Published:      %s %s\n", book.Month, book.Year)
			} else {
				fmt.Printf("Published:      %s\n", book.Year)
			}
		}
		if book.Publisher != "" {
			fmt.Printf("Publisher:      %s\n", book.Publisher)
		}
		if book.ISBN13 != "" {
			fmt.Printf("ISBN-13:        %s\n", book.ISBN13)
		}
		if book.ISBN != "" && book.ISBN != book.ISBN13 {
			fmt.Printf("ISBN-10:        %s\n", book.ISBN)
		}
		if book.Pages > 0 {
			fmt.Printf("Pages:          %d\n", book.Pages)
		}
		if book.Format != "" {
			fmt.Printf("Format:         %s\n", book.Format)
		}
		if book.Language != "" {
			fmt.Printf("Language:       %s\n", book.Language)
		}
		if book.URL != "" {
			fmt.Printf("URL:            %s\n", book.URL)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(bookCmd)
	bookCmd.Flags().BoolVar(&bookJSONFlag, "json", false, "Output the book as JSON")
}
