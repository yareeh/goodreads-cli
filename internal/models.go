package internal

type Book struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Rating      string `json:"rating"`
	URL         string `json:"url"`
	ImageURL    string `json:"image_url"`
	Description string `json:"description"`
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
