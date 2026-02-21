package internal

import (
	"encoding/json"
	"testing"
)

func TestBookJSON(t *testing.T) {
	book := Book{
		ID:          "55145261",
		Title:       "Project Hail Mary",
		Author:      "Andy Weir",
		Rating:      "4.5",
		URL:         "https://www.goodreads.com/book/show/55145261",
		ImageURL:    "https://example.com/img.jpg",
		Description: "A lone astronaut must save the earth.",
	}

	data, err := json.Marshal(book)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded Book
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.ID != book.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, book.ID)
	}
	if decoded.Title != book.Title {
		t.Errorf("Title = %q, want %q", decoded.Title, book.Title)
	}
	if decoded.Author != book.Author {
		t.Errorf("Author = %q, want %q", decoded.Author, book.Author)
	}
}

func TestShelfJSON(t *testing.T) {
	shelf := Shelf{Name: "want-to-read", BookCount: 42}

	data, err := json.Marshal(shelf)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded Shelf
	json.Unmarshal(data, &decoded)
	if decoded.Name != "want-to-read" {
		t.Errorf("Name = %q, want %q", decoded.Name, "want-to-read")
	}
	if decoded.BookCount != 42 {
		t.Errorf("BookCount = %d, want 42", decoded.BookCount)
	}
}

func TestReadingProgressJSON(t *testing.T) {
	rp := ReadingProgress{
		BookID:      "12345",
		CurrentPage: 150,
		TotalPages:  300,
		Percent:     50,
	}

	data, _ := json.Marshal(rp)
	var decoded ReadingProgress
	json.Unmarshal(data, &decoded)

	if decoded.Percent != 50 {
		t.Errorf("Percent = %d, want 50", decoded.Percent)
	}
	if decoded.BookID != "12345" {
		t.Errorf("BookID = %q, want %q", decoded.BookID, "12345")
	}
}
