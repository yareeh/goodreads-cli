package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yareeh/goodreads-cli/internal"
)

var listShelfJSONFlag bool

var listShelfCmd = &cobra.Command{
	Use:   "list-shelf <shelf-name>",
	Short: "List the books on a shelf (currently-reading, want-to-read, read, …)",
	Long: `List the books on one of the logged-in user's shelves.

Goodreads renders each shelf at /review/list/<user_id>?shelf=<name>. This
command logs in (via the saved session cookies), discovers the user ID from
the signed-in home page, fetches the shelf, and prints the books on it.

Examples:
  goodreads list-shelf currently-reading
  goodreads list-shelf currently-reading --json
  goodreads list-shelf want-to-read
  goodreads list-shelf read`,
	Aliases: []string{"shelf-list", "shelved"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shelfName := args[0]

		// The /review/list/<user>?shelf=… endpoint has been walled
		// behind AWS WAF since July 2026 — the plain HTTP client sees
		// a 202 JS challenge. Route through rod, which executes the
		// challenge and returns the real page.
		fmt.Fprintln(cmd.ErrOrStderr(), "Launching browser (needed to clear AWS WAF challenge on shelf pages)…")
		browser, err := internal.NewBrowser(!noHeadless)
		if err != nil {
			return fmt.Errorf("launching browser: %w", err)
		}
		defer browser.Close()

		if !browser.IsLoggedIn() {
			return fmt.Errorf("not logged in — run 'goodreads login' first")
		}

		books, err := browser.ListShelf(shelfName)
		if err != nil {
			return fmt.Errorf("listing shelf %q: %w", shelfName, err)
		}

		if listShelfJSONFlag {
			data, err := json.MarshalIndent(books, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		if len(books) == 0 {
			fmt.Printf("No books on shelf %q.\n", shelfName)
			return nil
		}

		fmt.Printf("%-12s %-60s %s\n", "ID", "TITLE", "AUTHOR")
		fmt.Printf("%-12s %-60s %s\n", "---", "-----", "------")
		for _, b := range books {
			title := b.Title
			if len(title) > 58 {
				title = title[:55] + "..."
			}
			fmt.Printf("%-12s %-60s %s\n", b.ID, title, b.Author)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listShelfCmd)
	listShelfCmd.Flags().BoolVar(&listShelfJSONFlag, "json", false, "Output the shelf as JSON")
}
