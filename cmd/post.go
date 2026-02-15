package cmd

import (
	"fmt"

	"github.com/jari/goodreads-cli/internal"
	"github.com/spf13/cobra"
)

var message string

var postCmd = &cobra.Command{
	Use:   "post <topic-id>",
	Short: "Post to a discussion",
	Long:  "Post a comment to a Goodreads discussion topic",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		topicID := args[0]

		fmt.Println("Launching browser...")
		browser, err := internal.NewBrowser(true)
		if err != nil {
			return fmt.Errorf("launching browser: %w", err)
		}
		defer browser.Close()

		if !browser.IsLoggedIn() {
			return fmt.Errorf("not logged in — run 'goodreads login' first")
		}

		fmt.Printf("Posting to topic %s...\n", topicID)
		if err := internal.PostComment(browser, topicID, message); err != nil {
			return err
		}

		fmt.Println("Done!")
		return nil
	},
}

func init() {
	postCmd.Flags().StringVar(&message, "message", "", "message to post")
	_ = postCmd.MarkFlagRequired("message")
	rootCmd.AddCommand(postCmd)
}
