package internal

import (
	"fmt"
	"time"
)

// PostComment posts a comment to a Goodreads discussion topic.
func PostComment(b *Browser, topicID string, message string) error {
	url := fmt.Sprintf("https://www.goodreads.com/topic/show/%s", topicID)
	b.Page.MustNavigate(url)
	b.Page.MustWaitStable()

	// Find the comment textarea
	textarea, err := b.Page.Timeout(10 * time.Second).Element(`#comment_body_usertext`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find comment textarea — are you on a valid topic page?: %w", err)
	}

	// Fill in the message
	textarea.MustClick()
	textarea.MustInput(message)

	// Click the Post button
	postBtn, err := b.Page.Timeout(5 * time.Second).Element(`input[type="submit"][value="Post"]`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find Post button: %w", err)
	}
	postBtn.MustClick()

	// Wait for the post to complete
	b.Page.MustWaitStable()
	time.Sleep(2 * time.Second)

	return b.SaveCookies()
}
