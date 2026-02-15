package cmd

import (
	"fmt"

	"github.com/jari/goodreads-cli/internal"
	"github.com/spf13/cobra"
)

var (
	replyMessage  string
	replyBookID   string
	replyAuthorID string
)

var postReplyCmd = &cobra.Command{
	Use:   "post-reply <topic-id>",
	Short: "Reply to a discussion topic",
	Long:  "Post a comment to an existing Goodreads discussion topic",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		topicID := args[0]

		fmt.Println("Launching browser...")
		browser, err := internal.NewBrowser(!noHeadless)
		if err != nil {
			return fmt.Errorf("launching browser: %w", err)
		}
		defer browser.Close()

		if !browser.IsLoggedIn() {
			return fmt.Errorf("not logged in — run 'goodreads login' first")
		}

		fmt.Printf("Posting reply to topic %s...\n", topicID)
		if err := internal.PostReply(browser, topicID, replyMessage, replyBookID, replyAuthorID); err != nil {
			return err
		}

		fmt.Println("Done!")
		return nil
	},
}

var (
	topicURL       string
	topicSubject   string
	topicMessage   string
	topicBookID    string
	topicAuthorID  string
)

var postTopicCmd = &cobra.Command{
	Use:   "post-topic",
	Short: "Create a new discussion topic",
	Long: `Create a new discussion topic in a Goodreads group.

The --url flag should be the full new-topic URL from Goodreads, e.g.:
  https://www.goodreads.com/topic/new?context_id=220-goodreads-librarians-group&context_type=Group&topic[folder_id]=120471`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Launching browser...")
		browser, err := internal.NewBrowser(!noHeadless)
		if err != nil {
			return fmt.Errorf("launching browser: %w", err)
		}
		defer browser.Close()

		if !browser.IsLoggedIn() {
			return fmt.Errorf("not logged in — run 'goodreads login' first")
		}

		fmt.Println("Creating new topic...")
		if err := internal.PostNewTopic(browser, topicURL, topicSubject, topicMessage, topicBookID, topicAuthorID); err != nil {
			return err
		}

		fmt.Println("Done!")
		return nil
	},
}

func init() {
	postReplyCmd.Flags().StringVar(&replyMessage, "message", "", "message to post")
	postReplyCmd.Flags().StringVar(&replyBookID, "book", "", "book ID to reference")
	postReplyCmd.Flags().StringVar(&replyAuthorID, "author", "", "author ID to reference")
	_ = postReplyCmd.MarkFlagRequired("message")
	rootCmd.AddCommand(postReplyCmd)

	postTopicCmd.Flags().StringVar(&topicURL, "url", "", "full new-topic URL from Goodreads")
	postTopicCmd.Flags().StringVar(&topicSubject, "subject", "", "topic subject/title")
	postTopicCmd.Flags().StringVar(&topicMessage, "message", "", "topic body message")
	postTopicCmd.Flags().StringVar(&topicBookID, "book", "", "book ID to reference")
	postTopicCmd.Flags().StringVar(&topicAuthorID, "author", "", "author ID to reference")
	_ = postTopicCmd.MarkFlagRequired("url")
	_ = postTopicCmd.MarkFlagRequired("subject")
	_ = postTopicCmd.MarkFlagRequired("message")
	rootCmd.AddCommand(postTopicCmd)
}
