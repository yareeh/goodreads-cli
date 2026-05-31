package internal

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ParseBookDetailsFromHTML extracts the bibliographic fields of a Goodreads
// book page into a populated Book. legacyID is the numeric ID from the URL
// (e.g. "18690730" for /book/show/18690730-tuokio-tuulessa).
//
// It reads the embedded __NEXT_DATA__ Apollo state — that's where the
// per-edition details live (ISBN, publisher, edition year). The JSON-LD
// block is a useful fallback for title/author when Apollo isn't there.
//
// The function does not perform any network I/O.
func ParseBookDetailsFromHTML(html string, legacyID string) (Book, error) {
	b := Book{ID: legacyID}

	apollo, err := extractApolloState(html)
	if err != nil {
		// Fall back to JSON-LD when Apollo isn't present (e.g. very old
		// page layouts or partial server-side renders).
		if ld, ldErr := extractJSONLD(html); ldErr == nil {
			applyJSONLD(&b, ld)
			return b, nil
		}
		return b, fmt.Errorf("could not extract structured data: %w", err)
	}

	book := findBookNode(apollo, legacyID)
	if book == nil {
		return b, fmt.Errorf("no Book node found for legacyId %s", legacyID)
	}

	if s, ok := book["title"].(string); ok {
		b.Title = s
	}
	if s, ok := book["webUrl"].(string); ok {
		b.URL = s
	}
	if s, ok := book["description"].(string); ok {
		b.Description = s
	}
	if s, ok := book["imageUrl"].(string); ok {
		b.ImageURL = s
	}

	// Author: chase the primaryContributorEdge.node.__ref into the
	// Contributor node and read its name.
	if edge, ok := book["primaryContributorEdge"].(map[string]any); ok {
		if node, ok := edge["node"].(map[string]any); ok {
			if ref, ok := node["__ref"].(string); ok {
				if c, ok := apollo[ref].(map[string]any); ok {
					if name, ok := c["name"].(string); ok {
						b.Author = name
					}
				}
			}
		}
	}

	// Edition details (ISBN, publisher, year, pages, language, format).
	if details, ok := book["details"].(map[string]any); ok {
		if s, ok := details["isbn"].(string); ok {
			b.ISBN = s
		}
		if s, ok := details["isbn13"].(string); ok {
			b.ISBN13 = s
		}
		if s, ok := details["publisher"].(string); ok {
			b.Publisher = s
		}
		if s, ok := details["format"].(string); ok {
			b.Format = s
		}
		if n, ok := details["numPages"].(float64); ok {
			b.Pages = int(n)
		}
		if lang, ok := details["language"].(map[string]any); ok {
			if s, ok := lang["name"].(string); ok {
				b.Language = s
			}
		}
		if ms, ok := details["publicationTime"].(float64); ok {
			t := time.UnixMilli(int64(ms)).UTC()
			b.Year = fmt.Sprintf("%d", t.Year())
			b.Month = t.Month().String()
		}
	}

	// Original title comes from the Work that this edition belongs to —
	// the Book node carries a `work.__ref` edge.
	if work, ok := book["work"].(map[string]any); ok {
		if ref, ok := work["__ref"].(string); ok {
			if w, ok := apollo[ref].(map[string]any); ok {
				if wd, ok := w["details"].(map[string]any); ok {
					if s, ok := wd["originalTitle"].(string); ok {
						b.OriginalTitle = s
					}
				}
			}
		}
	}

	return b, nil
}

// extractApolloState pulls the __NEXT_DATA__ payload's apolloState map
// out of the page. Returns the flat reference map keyed by entity
// references like "Book:kca://..." or "Work:kca://...".
func extractApolloState(html string) (map[string]any, error) {
	re := regexp.MustCompile(`(?s)__NEXT_DATA__"\s+type="application/json"\s*>(.*?)</script>`)
	m := re.FindStringSubmatch(html)
	if len(m) < 2 {
		return nil, fmt.Errorf("__NEXT_DATA__ block not found")
	}
	var payload struct {
		Props struct {
			PageProps struct {
				ApolloState map[string]any `json:"apolloState"`
			} `json:"pageProps"`
		} `json:"props"`
	}
	if err := json.Unmarshal([]byte(m[1]), &payload); err != nil {
		return nil, fmt.Errorf("decoding __NEXT_DATA__: %w", err)
	}
	if payload.Props.PageProps.ApolloState == nil {
		return nil, fmt.Errorf("apolloState missing from __NEXT_DATA__")
	}
	return payload.Props.PageProps.ApolloState, nil
}

// findBookNode walks the Apollo state map looking for the Book entry whose
// legacyId matches the requested ID. Multiple Book nodes appear on a single
// page (the displayed edition + the canonical work's primary book); we want
// the one the user actually opened.
func findBookNode(apollo map[string]any, legacyID string) map[string]any {
	for k, v := range apollo {
		if !strings.HasPrefix(k, "Book:") {
			continue
		}
		node, ok := v.(map[string]any)
		if !ok {
			continue
		}
		var idStr string
		switch n := node["legacyId"].(type) {
		case float64:
			idStr = fmt.Sprintf("%d", int64(n))
		case string:
			idStr = n
		}
		if idStr == legacyID {
			return node
		}
	}
	return nil
}

// extractJSONLD returns the first application/ld+json block as a generic map.
// Used as a fallback when Apollo state isn't present.
func extractJSONLD(html string) (map[string]any, error) {
	re := regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`)
	m := re.FindStringSubmatch(html)
	if len(m) < 2 {
		return nil, fmt.Errorf("JSON-LD block not found")
	}
	var ld map[string]any
	if err := json.Unmarshal([]byte(m[1]), &ld); err != nil {
		return nil, fmt.Errorf("decoding JSON-LD: %w", err)
	}
	return ld, nil
}

func applyJSONLD(b *Book, ld map[string]any) {
	if s, ok := ld["name"].(string); ok {
		b.Title = s
	}
	if s, ok := ld["isbn"].(string); ok {
		// JSON-LD's "isbn" is usually ISBN-13 on Goodreads pages.
		if len(s) == 13 {
			b.ISBN13 = s
		} else {
			b.ISBN = s
		}
	}
	if s, ok := ld["inLanguage"].(string); ok {
		b.Language = s
	}
	if s, ok := ld["bookFormat"].(string); ok {
		b.Format = s
	}
	if n, ok := ld["numberOfPages"].(float64); ok {
		b.Pages = int(n)
	}
	// JSON-LD author is usually an array of {@type: Person, name}.
	if authors, ok := ld["author"].([]any); ok && len(authors) > 0 {
		if a, ok := authors[0].(map[string]any); ok {
			if name, ok := a["name"].(string); ok {
				b.Author = name
			}
		}
	}
}
