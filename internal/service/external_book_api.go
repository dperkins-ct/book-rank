package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ExternalBookAPIService handles fetching book metadata from external APIs
type ExternalBookAPIService struct {
	httpClient *http.Client
}

// NewExternalBookAPIService creates a new ExternalBookAPIService
func NewExternalBookAPIService() *ExternalBookAPIService {
	return &ExternalBookAPIService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// BookMetadataResult represents the result from external book API
type BookMetadataResult struct {
	Title           string    `json:"title"`
	Author          string    `json:"author"`
	Authors         []string  `json:"authors,omitempty"`
	Genre           string    `json:"genre,omitempty"`
	Genres          []string  `json:"genres,omitempty"`
	PublicationDate *time.Time `json:"publication_date,omitempty"`
	Description     string    `json:"description,omitempty"`
	ISBN            string    `json:"isbn,omitempty"`
	CoverURL        string    `json:"cover_url,omitempty"`
	PageCount       int       `json:"page_count,omitempty"`
	Publisher       string    `json:"publisher,omitempty"`
	Language        string    `json:"language,omitempty"`
	Source          string    `json:"source,omitempty"`
}

// OpenLibraryResponse represents the Open Library API response
type OpenLibraryResponse struct {
	Docs []struct {
		Title           string    `json:"title"`
		AuthorName      []string  `json:"author_name"`
		Subject         []string  `json:"subject"`
		FirstPublishYear int     `json:"first_publish_year"`
		ISBN            []string  `json:"isbn"`
		CoverI          int       `json:"cover_i"`
		PageCountMedian int       `json:"page_count_median"`
		Publisher       []string  `json:"publisher"`
		Language        []string  `json:"language"`
	} `json:"docs"`
	NumFound int `json:"numFound"`
}

// FetchBookMetadata fetches book metadata from external APIs
func (s *ExternalBookAPIService) FetchBookMetadata(title, author string) (*BookMetadataResult, error) {
	// Try Open Library API first (it's free and has good coverage)
	result, err := s.fetchFromOpenLibrary(title, author)
	if err == nil && result != nil {
		return result, nil
	}

	// Could add more APIs here like Google Books API, etc.
	// For now, return what we found or error
	return result, err
}

// fetchFromOpenLibrary fetches book data from Open Library API
func (s *ExternalBookAPIService) fetchFromOpenLibrary(title, author string) (*BookMetadataResult, error) {
	// Build search query
	query := fmt.Sprintf("title:%s", url.QueryEscape(title))
	if author != "" {
		query += fmt.Sprintf(" AND author:%s", url.QueryEscape(author))
	}

	// Build request URL
	apiURL := fmt.Sprintf("https://openlibrary.org/search.json?q=%s&limit=1", url.QueryEscape(query))

	// Make request
	resp, err := s.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Open Library: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Open Library API returned status %d", resp.StatusCode)
	}

	// Parse response
	var apiResp OpenLibraryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Open Library response: %w", err)
	}

	// Check if we found any results
	if apiResp.NumFound == 0 || len(apiResp.Docs) == 0 {
		return nil, fmt.Errorf("no books found in Open Library")
	}

	// Convert to our format
	doc := apiResp.Docs[0]
	result := &BookMetadataResult{
		Title:  doc.Title,
		Source: "Open Library",
	}

	// Set author (use first author if multiple)
	if len(doc.AuthorName) > 0 {
		result.Author = doc.AuthorName[0]
		result.Authors = doc.AuthorName
	}

	// Set genre/subject (use first subject as genre)
	if len(doc.Subject) > 0 {
		result.Genre = s.cleanGenre(doc.Subject[0])
		result.Genres = doc.Subject[:min(5, len(doc.Subject))] // Limit to first 5
	}

	// Set publication date
	if doc.FirstPublishYear > 0 {
		pubDate := time.Date(doc.FirstPublishYear, 1, 1, 0, 0, 0, 0, time.UTC)
		result.PublicationDate = &pubDate
	}

	// Set ISBN (use first ISBN)
	if len(doc.ISBN) > 0 {
		result.ISBN = doc.ISBN[0]
	}

	// Set cover URL if available
	if doc.CoverI > 0 {
		result.CoverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", doc.CoverI)
	}

	// Set page count
	if doc.PageCountMedian > 0 {
		result.PageCount = doc.PageCountMedian
	}

	// Set publisher (use first publisher)
	if len(doc.Publisher) > 0 {
		result.Publisher = doc.Publisher[0]
	}

	// Set language (use first language)
	if len(doc.Language) > 0 {
		result.Language = doc.Language[0]
	}

	return result, nil
}

// cleanGenre cleans up genre strings from external APIs
func (s *ExternalBookAPIService) cleanGenre(genre string) string {
	// Remove common prefixes/suffixes and normalize
	cleaned := strings.TrimSpace(genre)
	cleaned = strings.ReplaceAll(cleaned, "_", " ")

	// Capitalize first letter of each word
	words := strings.Fields(cleaned)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

// SearchBooks searches for books using external APIs
func (s *ExternalBookAPIService) SearchBooks(query string, limit int) ([]*BookMetadataResult, error) {
	if limit <= 0 {
		limit = 10
	}

	// Build search URL
	apiURL := fmt.Sprintf("https://openlibrary.org/search.json?q=%s&limit=%d",
		url.QueryEscape(query), limit)

	// Make request
	resp, err := s.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search Open Library: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Open Library API returned status %d", resp.StatusCode)
	}

	// Parse response
	var apiResp OpenLibraryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Open Library response: %w", err)
	}

	// Convert to our format
	var results []*BookMetadataResult
	for _, doc := range apiResp.Docs {
		result := &BookMetadataResult{
			Title:  doc.Title,
			Source: "Open Library",
		}

		// Set author
		if len(doc.AuthorName) > 0 {
			result.Author = doc.AuthorName[0]
			result.Authors = doc.AuthorName
		}

		// Set genre
		if len(doc.Subject) > 0 {
			result.Genre = s.cleanGenre(doc.Subject[0])
			result.Genres = doc.Subject[:min(3, len(doc.Subject))]
		}

		// Set publication date
		if doc.FirstPublishYear > 0 {
			pubDate := time.Date(doc.FirstPublishYear, 1, 1, 0, 0, 0, 0, time.UTC)
			result.PublicationDate = &pubDate
		}

		// Set cover URL
		if doc.CoverI > 0 {
			result.CoverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-S.jpg", doc.CoverI)
		}

		results = append(results, result)
	}

	return results, nil
}