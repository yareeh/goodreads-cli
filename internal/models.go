package internal

type Book struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Rating      string `json:"rating"`
	URL         string `json:"url"`
	ImageURL    string `json:"image_url"`
	Description string `json:"description"`

	// Bibliographic details populated by ParseBookDetailsFromHTML when a
	// full book page is fetched. Empty when the Book comes from autocomplete
	// search results.
	ISBN          string `json:"isbn,omitempty"`           // ISBN-10
	ISBN13        string `json:"isbn13,omitempty"`         // ISBN-13
	Publisher     string `json:"publisher,omitempty"`      // edition publisher
	OriginalTitle string `json:"original_title,omitempty"` // title in the original language
	Year          string `json:"year,omitempty"`           // edition publication year (YYYY)
	Month         string `json:"month,omitempty"`          // edition publication month (full English name)
	Pages         int    `json:"pages,omitempty"`
	Language      string `json:"language,omitempty"`
	Format        string `json:"format,omitempty"` // "Hardcover", "Paperback", "ebook", ...
}

type Shelf struct {
	Name      string `json:"name"`
	BookCount int    `json:"book_count"`
}

type ReadingProgress struct {
	BookID      string `json:"book_id"`
	CurrentPage int    `json:"current_page"`
	TotalPages  int    `json:"total_pages"`
	Percent     int    `json:"percent"`
}
