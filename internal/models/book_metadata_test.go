package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBookMetadataModel(t *testing.T) {
	tests := map[string]struct {
		metadata BookMetadata
		wantErr  bool
	}{
		"valid metadata with external_id": {
			metadata: BookMetadata{
				BookID:     1,
				ExternalID: "OL123456M",
				Source:     SourceOpenLibrary,
				AdditionalData: JSON{
					"isbn":        "9780123456789",
					"cover_url":   "https://example.com/cover.jpg",
					"description": "A great book",
				},
			},
			wantErr: false,
		},
		"valid metadata without external_id": {
			metadata: BookMetadata{
				BookID: 1,
				Source: SourceManual,
				AdditionalData: JSON{
					"notes": "Added manually",
				},
			},
			wantErr: false,
		},
		"google books metadata": {
			metadata: BookMetadata{
				BookID:     1,
				ExternalID: "abc123",
				Source:     SourceGoogleBooks,
				AdditionalData: JSON{
					"volume_id":   "abc123",
					"thumbnail":   "https://books.google.com/thumbnail.jpg",
					"page_count":  250,
					"categories":  []string{"Fiction", "Adventure"},
				},
			},
			wantErr: false,
		},
		"empty source": {
			metadata: BookMetadata{
				BookID:     1,
				ExternalID: "OL123456M",
				Source:     "",
			},
			wantErr: false, // SQLite doesn't enforce NOT NULL on empty strings the same way
		},
		"invalid source": {
			metadata: BookMetadata{
				BookID:     1,
				ExternalID: "OL123456M",
				Source:     MetadataSource("invalid"),
			},
			wantErr: false, // Custom validation would need to be in service layer
		},
	}

	db := setupTestDB()

	// Create a user and book first
	user := User{
		Username:     "testuser",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	book := Book{
		Title:     "Test Book",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("Failed to create test book: %v", err)
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Set valid foreign key
			tc.metadata.BookID = book.ID

			err := db.Create(&tc.metadata).Error
			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !tc.wantErr {
				// Verify timestamps were set
				if tc.metadata.CreatedAt.IsZero() {
					t.Error("CreatedAt should be set")
				}
				if tc.metadata.UpdatedAt.IsZero() {
					t.Error("UpdatedAt should be set")
				}
				if tc.metadata.ID == 0 {
					t.Error("ID should be set")
				}
			}
		})
	}
}

func TestBookMetadataJSONField(t *testing.T) {
	db := setupTestDB()

	// Create a user and book first
	user := User{
		Username:     "testuser",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	book := Book{
		Title:     "Test Book",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("Failed to create test book: %v", err)
	}

	// Test complex JSON data
	complexData := JSON{
		"isbn":       "9780123456789",
		"cover_url":  "https://example.com/cover.jpg",
		"page_count": 250,
		"authors":    []string{"Author 1", "Author 2"},
		"categories": []string{"Fiction", "Adventure"},
		"ratings": map[string]interface{}{
			"average": 4.5,
			"count":   123,
		},
		"published_date": "2023-01-01",
		"language":       "en",
	}

	metadata := BookMetadata{
		BookID:         book.ID,
		ExternalID:     "test123",
		Source:         SourceOpenLibrary,
		AdditionalData: complexData,
	}

	err := db.Create(&metadata).Error
	if err != nil {
		t.Fatalf("Failed to create metadata: %v", err)
	}

	// Retrieve and verify JSON data
	var retrievedMetadata BookMetadata
	err = db.First(&retrievedMetadata, metadata.ID).Error
	if err != nil {
		t.Fatalf("Failed to retrieve metadata: %v", err)
	}

	if retrievedMetadata.AdditionalData == nil {
		t.Fatal("AdditionalData should not be nil")
	}

	// Verify specific fields
	if retrievedMetadata.AdditionalData["isbn"] != "9780123456789" {
		t.Errorf("ISBN = %v, want '9780123456789'", retrievedMetadata.AdditionalData["isbn"])
	}

	if retrievedMetadata.AdditionalData["page_count"] != float64(250) {
		t.Errorf("PageCount = %v, want 250", retrievedMetadata.AdditionalData["page_count"])
	}

	// Test array field
	authors, ok := retrievedMetadata.AdditionalData["authors"].([]interface{})
	if !ok {
		t.Errorf("Authors should be an array")
	} else if len(authors) != 2 {
		t.Errorf("Authors length = %d, want 2", len(authors))
	}

	// Test nested object
	ratings, ok := retrievedMetadata.AdditionalData["ratings"].(map[string]interface{})
	if !ok {
		t.Errorf("Ratings should be an object")
	} else if ratings["average"] != 4.5 {
		t.Errorf("Ratings average = %v, want 4.5", ratings["average"])
	}
}

func TestJSONValueAndScan(t *testing.T) {
	tests := map[string]struct {
		json    JSON
		wantErr bool
	}{
		"nil json": {
			json:    nil,
			wantErr: false,
		},
		"empty json": {
			json:    JSON{},
			wantErr: false,
		},
		"simple json": {
			json: JSON{
				"key": "value",
			},
			wantErr: false,
		},
		"complex json": {
			json: JSON{
				"string":  "value",
				"number":  42,
				"boolean": true,
				"array":   []string{"a", "b", "c"},
				"object": map[string]interface{}{
					"nested": "value",
				},
			},
			wantErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Test Value() method
			value, err := tc.json.Value()
			if (err != nil) != tc.wantErr {
				t.Errorf("Value() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !tc.wantErr && tc.json != nil {
				// Test that value can be marshaled back to JSON
				var testJSON JSON
				if value != nil {
					bytes, ok := value.([]byte)
					if !ok {
						t.Errorf("Value() should return []byte, got %T", value)
					} else {
						err := json.Unmarshal(bytes, &testJSON)
						if err != nil {
							t.Errorf("Failed to unmarshal Value() result: %v", err)
						}
					}
				}
			}

			// Test Scan() method
			if !tc.wantErr && value != nil {
				var scannedJSON JSON
				err := scannedJSON.Scan(value)
				if err != nil {
					t.Errorf("Scan() error = %v", err)
				}

				// Compare original and scanned JSON
				if len(tc.json) > 0 {
					originalBytes, _ := json.Marshal(tc.json)
					scannedBytes, _ := json.Marshal(scannedJSON)
					if string(originalBytes) != string(scannedBytes) {
						t.Errorf("Scanned JSON doesn't match original")
					}
				}
			}
		})
	}
}

func TestBookMetadataBeforeCreateHook(t *testing.T) {
	db := setupTestDB()

	// Create a user and book first
	user := User{
		Username:     "testuser",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	book := Book{
		Title:     "Test Book",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("Failed to create test book: %v", err)
	}

	metadata := BookMetadata{
		BookID:     book.ID,
		ExternalID: "test123",
		Source:     SourceOpenLibrary,
	}

	err := db.Create(&metadata).Error
	if err != nil {
		t.Fatalf("Failed to create metadata: %v", err)
	}

	if metadata.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set by BeforeCreate hook")
	}

	if metadata.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set by BeforeCreate hook")
	}

	// Check that both timestamps are approximately equal (within 1 second)
	diff := metadata.UpdatedAt.Sub(metadata.CreatedAt)
	if diff > time.Second || diff < -time.Second {
		t.Error("CreatedAt and UpdatedAt should be set to approximately the same time")
	}
}

func TestBookMetadataBeforeUpdateHook(t *testing.T) {
	db := setupTestDB()

	// Create a user and book first
	user := User{
		Username:     "testuser",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	book := Book{
		Title:     "Test Book",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("Failed to create test book: %v", err)
	}

	metadata := BookMetadata{
		BookID:     book.ID,
		ExternalID: "test123",
		Source:     SourceOpenLibrary,
	}

	// Create metadata
	err := db.Create(&metadata).Error
	if err != nil {
		t.Fatalf("Failed to create metadata: %v", err)
	}

	originalUpdatedAt := metadata.UpdatedAt

	// Wait a small amount to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update metadata
	metadata.ExternalID = "updated123"
	err = db.Save(&metadata).Error
	if err != nil {
		t.Fatalf("Failed to update metadata: %v", err)
	}

	if !metadata.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated by BeforeUpdate hook")
	}
}

func TestBookMetadataTableName(t *testing.T) {
	metadata := BookMetadata{}
	expected := "book_metadata"
	if metadata.TableName() != expected {
		t.Errorf("TableName() = %v, want %v", metadata.TableName(), expected)
	}
}

func TestBookMetadataIsValidSource(t *testing.T) {
	tests := map[string]struct {
		source MetadataSource
		want   bool
	}{
		"openlibrary": {
			source: SourceOpenLibrary,
			want:   true,
		},
		"googlebooks": {
			source: SourceGoogleBooks,
			want:   true,
		},
		"manual": {
			source: SourceManual,
			want:   true,
		},
		"invalid": {
			source: MetadataSource("invalid"),
			want:   false,
		},
		"empty": {
			source: MetadataSource(""),
			want:   false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			metadata := BookMetadata{Source: tc.source}
			got := metadata.IsValidSource()
			if got != tc.want {
				t.Errorf("IsValidSource() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMetadataSourceConstants(t *testing.T) {
	tests := map[string]struct {
		constant MetadataSource
		expected string
	}{
		"openlibrary": {
			constant: SourceOpenLibrary,
			expected: "openlibrary",
		},
		"googlebooks": {
			constant: SourceGoogleBooks,
			expected: "googlebooks",
		},
		"manual": {
			constant: SourceManual,
			expected: "manual",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if string(tc.constant) != tc.expected {
				t.Errorf("Constant %s = %v, want %v", name, tc.constant, tc.expected)
			}
		})
	}
}