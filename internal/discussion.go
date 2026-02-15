package internal

import (
	"fmt"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

// PostReply posts a comment to an existing Goodreads discussion topic.
func PostReply(b *Browser, topicID string, message string, bookID string, authorID string) error {
	url := fmt.Sprintf("https://www.goodreads.com/topic/show/%s", topicID)
	b.Page.MustNavigate(url)
	b.Page.MustWaitStable()

	// Add book/author mention if requested
	if err := addMention(b, bookID, authorID); err != nil {
		return err
	}

	// Find the comment textarea
	textarea, err := b.Page.Timeout(10 * time.Second).Element(`#comment_body_usertext`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find comment textarea: %w", err)
	}

	textarea.MustClick()
	textarea.MustInput(message)

	// Click the Post button
	postBtn, err := b.Page.Timeout(5 * time.Second).Element(`input[type="submit"][value="Post"]`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find Post button: %w", err)
	}
	postBtn.MustClick()

	b.Page.MustWaitStable()
	time.Sleep(2 * time.Second)

	return b.SaveCookies()
}

// PostNewTopic creates a new discussion topic in a Goodreads group.
// The topicURL should be the full new-topic URL including context_id, context_type, and folder_id.
func PostNewTopic(b *Browser, topicURL string, subject string, message string, bookID string, authorID string) error {
	b.Page.MustNavigate(topicURL)
	b.Page.MustWaitStable()

	// Add book/author mention if requested
	if err := addMention(b, bookID, authorID); err != nil {
		return err
	}

	// Fill in the subject/title field
	subjectField, err := b.Page.Timeout(10 * time.Second).Element(`input[name="topic[subject]"], input[name="topic[title]"], #topic_subject, #topic_title`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find topic subject field: %w", err)
	}
	subjectField.MustClick()
	subjectField.MustInput(subject)

	// Fill in the body textarea
	textarea, err := b.Page.Timeout(5 * time.Second).Element(`#comment_body_usertext`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find comment textarea: %w", err)
	}
	textarea.MustClick()
	textarea.MustInput(message)

	// Click the Post button
	postBtn, err := b.Page.Timeout(5 * time.Second).Element(`input[type="submit"][value="Post"]`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find Post button: %w", err)
	}
	postBtn.MustClick()

	b.Page.MustWaitStable()
	time.Sleep(2 * time.Second)

	return b.SaveCookies()
}

// addMention opens the "add book/author" lightbox and searches for the given book or author.
func addMention(b *Browser, bookID string, authorID string) error {
	if bookID == "" && authorID == "" {
		return nil
	}

	// Resolve the name from the ID by visiting the page
	if bookID != "" {
		name, err := resolveBookName(b, bookID)
		if err != nil {
			return fmt.Errorf("resolving book name: %w", err)
		}
		if err := b.Page.NavigateBack(); err != nil {
			return fmt.Errorf("navigating back: %w", err)
		}
		b.Page.MustWaitStable()

		if err := addBookMention(b, name, bookID); err != nil {
			return err
		}
	}

	if authorID != "" {
		name, err := resolveAuthorName(b, authorID)
		if err != nil {
			return fmt.Errorf("resolving author name: %w", err)
		}
		if err := b.Page.NavigateBack(); err != nil {
			return fmt.Errorf("navigating back: %w", err)
		}
		b.Page.MustWaitStable()

		if err := addAuthorMention(b, name, authorID); err != nil {
			return err
		}
	}

	return nil
}

// resolveBookName navigates to a book page and extracts the title.
func resolveBookName(b *Browser, bookID string) (string, error) {
	b.Page.MustNavigate(fmt.Sprintf("https://www.goodreads.com/book/show/%s", bookID))
	b.Page.MustWaitStable()

	titleEl, err := b.Page.Timeout(10 * time.Second).Element(`h1[data-testid="bookTitle"], h1.Text__title1`)
	if err != nil {
		saveDebugScreenshot(b)
		return "", fmt.Errorf("could not find book title: %w", err)
	}
	return titleEl.MustText(), nil
}

// resolveAuthorName navigates to an author page and extracts the name.
func resolveAuthorName(b *Browser, authorID string) (string, error) {
	b.Page.MustNavigate(fmt.Sprintf("https://www.goodreads.com/author/show/%s", authorID))
	b.Page.MustWaitStable()

	nameEl, err := b.Page.Timeout(10 * time.Second).Element(`h1.authorName span[itemprop="name"], h1.authorName, .authorName span`)
	if err != nil {
		saveDebugScreenshot(b)
		return "", fmt.Errorf("could not find author name: %w", err)
	}
	return nameEl.MustText(), nil
}

// openMentionBox clicks "add book/author" to open the lightbox.
func openMentionBox(b *Browser) error {
	addLink, err := b.Page.Timeout(5 * time.Second).ElementR(`a`, "add book/author")
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find 'add book/author' link: %w", err)
	}
	if err := addLink.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("clicking 'add book/author': %w", err)
	}
	time.Sleep(1 * time.Second)
	return nil
}

// addBookMention opens the mention box, searches for a book by name, and clicks the Add button
// for the result matching the given book ID.
func addBookMention(b *Browser, bookName string, bookID string) error {
	if err := openMentionBox(b); err != nil {
		return err
	}

	// Book tab is selected by default. Type the book name and search.
	searchInput, err := b.Page.Timeout(5 * time.Second).Element(`#search_query`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find book search input: %w", err)
	}
	if err := searchInput.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("clicking search input: %w", err)
	}
	if err := searchInput.Input(bookName); err != nil {
		return fmt.Errorf("typing book name: %w", err)
	}

	// Click Search — form uses data-remote="true" (Rails AJAX)
	searchBtn, err := b.Page.Timeout(5 * time.Second).Element(`#add_mention_box_form input[type="submit"]`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find search button: %w", err)
	}
	if err := searchBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("clicking search button: %w", err)
	}
	time.Sleep(3 * time.Second)

	// Find the "Add" button whose onclick contains the book ID
	// The onclick looks like: gr.add_reference('[book:Title|228233676]')
	addBtnSelector := fmt.Sprintf(`#add_mention_book_results a.gr-button[onclick*="|%s]"]`, bookID)
	addBtn, err := b.Page.Timeout(10 * time.Second).Element(addBtnSelector)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find Add button for book %s: %w", bookID, err)
	}
	_, err = addBtn.Eval(`() => this.click()`, nil)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("clicking Add button for book: %w", err)
	}
	time.Sleep(1 * time.Second)

	return nil
}

// addAuthorMention opens the mention box, switches to Author tab, searches, and clicks the Add button
// for the result matching the given author ID.
func addAuthorMention(b *Browser, authorName string, authorID string) error {
	if err := openMentionBox(b); err != nil {
		return err
	}

	// Switch to Author tab
	authorTab, err := b.Page.Timeout(5 * time.Second).Element(`#authorLink`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find Author tab: %w", err)
	}
	if err := authorTab.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("clicking Author tab: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Type the author name and search
	authorInput, err := b.Page.Timeout(5 * time.Second).Element(`#quote_author_name`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find author search input: %w", err)
	}
	if err := authorInput.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("clicking author input: %w", err)
	}
	if err := authorInput.Input(authorName); err != nil {
		return fmt.Errorf("typing author name: %w", err)
	}

	searchBtn, err := b.Page.Timeout(5 * time.Second).Element(`#author_mention_form input[type="submit"]`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find author search button: %w", err)
	}
	if err := searchBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("clicking author search: %w", err)
	}
	time.Sleep(3 * time.Second)

	// Find the "Add" button whose onclick contains the author ID
	// The onclick looks like: gr.add_reference('[author:Name|513351]')
	addBtnSelector := fmt.Sprintf(`#add_mention_author_results a.gr-button[onclick*="|%s]"]`, authorID)
	addBtn, err := b.Page.Timeout(10 * time.Second).Element(addBtnSelector)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find Add button for author %s: %w", authorID, err)
	}
	_, err = addBtn.Eval(`() => this.click()`, nil)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("clicking Add button for author: %w", err)
	}
	time.Sleep(1 * time.Second)

	return nil
}
