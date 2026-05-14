package service

import (
	"bookrank/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ExternalAPIService handles integration with external book APIs
type ExternalAPIService struct {
	httpClient *http.Client
}

// NewExternalAPIService creates a new ExternalAPIService
func NewExternalAPIService() *ExternalAPIService {
	return &ExternalAPIService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// OpenLibraryBook represents a book from OpenLibrary API
type OpenLibraryBook struct {
	Key         string   `json:"key"`
	Title       string   `json:"title"`
	Authors     []Author `json:"authors"`
	PublishDate []string `json:"publish_date"`
	Subjects    []string `json:"subjects"`
	Description interface{} `json:"description"`
	Covers      []int    `json:"covers"`
	ISBN10      []string `json:"isbn_10"`
	ISBN13      []string `json:"isbn_13"`
}

// Author represents an author from OpenLibrary API
type Author struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// OpenLibrarySearchResult represents search results from OpenLibrary
type OpenLibrarySearchResult struct {
	NumFound int `json:"numFound"`
	Start    int `json:"start"`
	Docs     []struct {
		Key           string   `json:"key"`
		Title         string   `json:"title"`
		AuthorName    []string `json:"author_name"`
		FirstPublishYear int   `json:"first_publish_year"`
		Subject       []string `json:"subject"`
		CoverI        int      `json:"cover_i"`
		ISBN          []string `json:"isbn"`
	} `json:"docs"`
}

// GoogleBooksItem represents a book from Google Books API
type GoogleBooksItem struct {
	ID         string `json:"id"`
	VolumeInfo struct {
		Title         string   `json:"title"`
		Authors       []string `json:"authors"`
		Publisher     string   `json:"publisher"`
		PublishedDate string   `json:"publishedDate"`
		Description   string   `json:"description"`
		Categories    []string `json:"categories"`
		ImageLinks    struct {
			SmallThumbnail string `json:"smallThumbnail"`
			Thumbnail      string `json:"thumbnail"`
		} `json:"imageLinks"`
		IndustryIdentifiers []struct {
			Type       string `json:"type"`
			Identifier string `json:"identifier"`
		} `json:"industryIdentifiers"`
	} `json:"volumeInfo"`
}

// GoogleBooksResponse represents the response from Google Books API
type GoogleBooksResponse struct {
	Kind       string            `json:"kind"`
	TotalItems int               `json:"totalItems"`
	Items      []GoogleBooksItem `json:"items"`
}

// BookMetadata represents enriched book metadata
type BookMetadata struct {
	Title           string                 `json:"title"`
	Authors         []string               `json:"authors"`
	Description     string                 `json:"description"`
	PublishedDate   string                 `json:"published_date"`
	Categories      []string               `json:"categories"`
	CoverURL        string                 `json:"cover_url"`
	ISBN10          string                 `json:"isbn_10"`
	ISBN13          string                 `json:"isbn_13"`
	Publisher       string                 `json:"publisher"`
	AdditionalData  map[string]interface{} `json:"additional_data"`
}

// SearchOpenLibrary searches for books using OpenLibrary API
func (s *ExternalAPIService) SearchOpenLibrary(title, author string) (*BookMetadata, error) {
	baseURL := "https://openlibrary.org/search.json"
	params := url.Values{}

	if title != "" {
		params.Add("title", title)
	}
	if author != "" {
		params.Add("author", author)
	}
	params.Add("limit", "1")

	searchURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := s.httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search OpenLibrary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenLibrary API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenLibrary response: %w", err)
	}

	var searchResult OpenLibrarySearchResult
	if err := json.Unmarshal(body, &searchResult); err != nil {
		return nil, fmt.Errorf("failed to parse OpenLibrary response: %w", err)
	}

	if len(searchResult.Docs) == 0 {
		return nil, errors.New("no books found in OpenLibrary")
	}

	doc := searchResult.Docs[0]
	metadata := &BookMetadata{
		Title:       doc.Title,
		Authors:     doc.AuthorName,
		Categories:  doc.Subject,
		AdditionalData: map[string]interface{}{
			"openlibrary_key":    doc.Key,
			"first_publish_year": doc.FirstPublishYear,
		},
	}

	if len(doc.ISBN) > 0 {
		for _, isbn := range doc.ISBN {
			if len(isbn) == 10 {
				metadata.ISBN10 = isbn
			} else if len(isbn) == 13 {
				metadata.ISBN13 = isbn
			}
		}
	}

	if doc.CoverI > 0 {
		metadata.CoverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", doc.CoverI)
	}

	return metadata, nil
}

// SearchGoogleBooks searches for books using Google Books API
func (s *ExternalAPIService) SearchGoogleBooks(title, author string) (*BookMetadata, error) {
	baseURL := "https://www.googleapis.com/books/v1/volumes"
	params := url.Values{}

	query := ""
	if title != "" {
		query += "intitle:" + title
	}
	if author != "" {
		if query != "" {
			query += "+"
		}
		query += "inauthor:" + author
	}

	if query == "" {
		return nil, errors.New("either title or author must be provided")
	}

	params.Add("q", query)
	params.Add("maxResults", "1")

	searchURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := s.httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search Google Books: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google Books API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Google Books response: %w", err)
	}

	var booksResponse GoogleBooksResponse
	if err := json.Unmarshal(body, &booksResponse); err != nil {
		return nil, fmt.Errorf("failed to parse Google Books response: %w", err)
	}

	if len(booksResponse.Items) == 0 {
		return nil, errors.New("no books found in Google Books")
	}

	item := booksResponse.Items[0]
	volumeInfo := item.VolumeInfo

	metadata := &BookMetadata{
		Title:         volumeInfo.Title,
		Authors:       volumeInfo.Authors,
		Description:   volumeInfo.Description,
		PublishedDate: volumeInfo.PublishedDate,
		Categories:    volumeInfo.Categories,
		Publisher:     volumeInfo.Publisher,
		AdditionalData: map[string]interface{}{
			"google_books_id": item.ID,
		},
	}

	// Extract ISBNs
	for _, identifier := range volumeInfo.IndustryIdentifiers {
		switch identifier.Type {
		case "ISBN_10":
			metadata.ISBN10 = identifier.Identifier
		case "ISBN_13":
			metadata.ISBN13 = identifier.Identifier
		}
	}

	// Set cover URL
	if volumeInfo.ImageLinks.Thumbnail != "" {
		metadata.CoverURL = volumeInfo.ImageLinks.Thumbnail
	} else if volumeInfo.ImageLinks.SmallThumbnail != "" {
		metadata.CoverURL = volumeInfo.ImageLinks.SmallThumbnail
	}

	return metadata, nil
}

// FetchBookMetadata attempts to fetch metadata from multiple sources
func (s *ExternalAPIService) FetchBookMetadata(title, author string) (*BookMetadata, models.MetadataSource, error) {
	// Try OpenLibrary first
	if metadata, err := s.SearchOpenLibrary(title, author); err == nil {
		return metadata, models.SourceOpenLibrary, nil
	}

	// Fallback to Google Books
	if metadata, err := s.SearchGoogleBooks(title, author); err == nil {
		return metadata, models.SourceGoogleBooks, nil
	}

	return nil, "", errors.New("no metadata found from any source")
}

// ValidateBookData validates book data and enriches it with external metadata
func (s *ExternalAPIService) ValidateBookData(book *models.Book) error {
	// Basic validation
	if strings.TrimSpace(book.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(book.Author) == "" {
		return errors.New("author is required")
	}
	// Genre is optional - will be fetched automatically if not provided

	// Validate publication date if provided
	if book.PublicationDate != nil && book.PublicationDate.After(time.Now()) {
		return errors.New("publication date cannot be in the future")
	}

	// Validate description length
	if len(book.Description) > 5000 {
		return errors.New("description cannot exceed 5000 characters")
	}

	// Sanitize input
	book.Title = strings.TrimSpace(book.Title)
	book.Author = strings.TrimSpace(book.Author)
	book.Genre = strings.TrimSpace(book.Genre)
	book.Description = strings.TrimSpace(book.Description)

	return nil
}