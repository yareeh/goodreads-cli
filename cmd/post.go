package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var message string

var postCmd = &cobra.Command{
	Use:   "post <group>",
	Short: "Post to a discussion",
	Long:  "Post a message to a Goodreads group discussion",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		group := args[0]
		// TODO: Implement after recording the post-to-discussion endpoint
		fmt.Printf("Posting to group '%s'\n", group)
		fmt.Println("(Not yet implemented — use the recorder to capture request patterns)")
		return nil
	},
}

func init() {
	postCmd.Flags().StringVar(&message, "message", "", "message to post")
	_ = postCmd.MarkFlagRequired("message")
	rootCmd.AddCommand(postCmd)
}
