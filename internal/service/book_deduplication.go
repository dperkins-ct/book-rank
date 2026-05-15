package service

import (
	"bookrank/internal/models"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// BookDeduplicationService handles book deduplication logic
type BookDeduplicationService struct{}

// NewBookDeduplicationService creates a new BookDeduplicationService
func NewBookDeduplicationService() *BookDeduplicationService {
	return &BookDeduplicationService{}
}

// NormalizeString normalizes a string for comparison by:
// - Converting to lowercase
// - Removing diacritics/accents
// - Removing punctuation and extra spaces
func (s *BookDeduplicationService) NormalizeString(input string) string {
	// Convert to lowercase
	normalized := strings.ToLower(input)

	// Remove diacritics using unicode normalization
	transformer := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, _ = transform.String(transformer, normalized)

	// Remove punctuation and normalize spaces
	var builder strings.Builder
	for _, r := range normalized {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			builder.WriteRune(r)
		}
	}

	// Normalize spaces (multiple spaces become single space, trim)
	normalized = strings.Join(strings.Fields(builder.String()), " ")

	return strings.TrimSpace(normalized)
}

// CreateBookKey creates a unique key for book matching
func (s *BookDeduplicationService) CreateBookKey(title, author string) string {
	normalizedTitle := s.NormalizeString(title)
	normalizedAuthor := s.NormalizeString(author)
	return normalizedTitle + "|" + normalizedAuthor
}

// BooksMatch determines if two books are likely the same
func (s *BookDeduplicationService) BooksMatch(book1, book2 *models.Book) bool {
	if book1 == nil || book2 == nil {
		return false
	}

	key1 := s.CreateBookKey(book1.Title, book1.Author)
	key2 := s.CreateBookKey(book2.Title, book2.Author)

	return key1 == key2
}

// CalculateMatchScore returns a score (0-100) indicating how likely two books are the same
func (s *BookDeduplicationService) CalculateMatchScore(book1, book2 *models.Book) int {
	if book1 == nil || book2 == nil {
		return 0
	}

	titleScore := s.stringSimilarity(s.NormalizeString(book1.Title), s.NormalizeString(book2.Title))
	authorScore := s.stringSimilarity(s.NormalizeString(book1.Author), s.NormalizeString(book2.Author))

	// Weighted average: title is more important than author for matching
	finalScore := int((titleScore * 0.7) + (authorScore * 0.3))

	return finalScore
}

// stringSimilarity calculates similarity between two strings (0-100)
// Uses a simple edit distance approach
func (s *BookDeduplicationService) stringSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 100.0
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Simple Levenshtein distance calculation
	distance := s.levenshteinDistance(s1, s2)
	maxLen := float64(max(len(s1), len(s2)))

	similarity := (1.0 - float64(distance)/maxLen) * 100.0

	if similarity < 0 {
		return 0.0
	}

	return similarity
}

// levenshteinDistance calculates the edit distance between two strings
func (s *BookDeduplicationService) levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}

	for j := 1; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			if s1[i-1] == s2[j-1] {
				matrix[i][j] = matrix[i-1][j-1]
			} else {
				matrix[i][j] = min(
					matrix[i-1][j]+1,   // deletion
					min(matrix[i][j-1]+1,   // insertion
					matrix[i-1][j-1]+1), // substitution
				)
			}
		}
	}

	return matrix[len(s1)][len(s2)]
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}