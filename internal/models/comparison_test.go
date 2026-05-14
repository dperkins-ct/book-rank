package models

import (
	"testing"

	"gorm.io/gorm"
)

func TestComparisonModel(t *testing.T) {
	tests := map[string]struct {
		comparison Comparison
		wantErr    bool
	}{
		"valid comparison prefer book_a": {
			comparison: Comparison{
				UserID:     1,
				BookAID:    1,
				BookBID:    2,
				Preference: PreferenceBookA,
			},
			wantErr: false,
		},
		"valid comparison prefer book_b": {
			comparison: Comparison{
				UserID:     1,
				BookAID:    1,
				BookBID:    2,
				Preference: PreferenceBookB,
			},
			wantErr: false,
		},
		"valid comparison tie": {
			comparison: Comparison{
				UserID:     1,
				BookAID:    1,
				BookBID:    2,
				Preference: PreferenceTie,
			},
			wantErr: false,
		},
		"same book comparison": {
			comparison: Comparison{
				UserID:     1,
				BookAID:    1,
				BookBID:    1, // Same book
				Preference: PreferenceBookA,
			},
			wantErr: true,
		},
		"invalid preference": {
			comparison: Comparison{
				UserID:     1,
				BookAID:    1,
				BookBID:    2,
				Preference: ComparisonPreference("invalid"),
			},
			wantErr: true,
		},
	}

	db := setupTestDB()

	// Create a user and books first
	user := User{
		Username:     "testuser",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	book1 := Book{
		Title:     "Test Book 1",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book1).Error; err != nil {
		t.Fatalf("Failed to create test book 1: %v", err)
	}

	book2 := Book{
		Title:     "Test Book 2",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book2).Error; err != nil {
		t.Fatalf("Failed to create test book 2: %v", err)
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Set valid foreign keys
			tc.comparison.UserID = user.ID
			if tc.comparison.BookAID == 1 {
				tc.comparison.BookAID = book1.ID
			}
			if tc.comparison.BookBID == 2 {
				tc.comparison.BookBID = book2.ID
			} else if tc.comparison.BookBID == 1 {
				tc.comparison.BookBID = book1.ID // For same book test
			}

			err := db.Create(&tc.comparison).Error
			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !tc.wantErr {
				// Verify timestamp was set
				if tc.comparison.CreatedAt.IsZero() {
					t.Error("CreatedAt should be set")
				}
				if tc.comparison.ID == 0 {
					t.Error("ID should be set")
				}
			}
		})
	}
}

func TestComparisonBeforeCreateHook(t *testing.T) {
	db := setupTestDB()

	// Create user and books
	user := User{
		Username:     "testuser",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	book1 := Book{
		Title:     "Test Book 1",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book1).Error; err != nil {
		t.Fatalf("Failed to create test book 1: %v", err)
	}

	book2 := Book{
		Title:     "Test Book 2",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book2).Error; err != nil {
		t.Fatalf("Failed to create test book 2: %v", err)
	}

	t.Run("sets timestamp", func(t *testing.T) {
		comparison := Comparison{
			UserID:     user.ID,
			BookAID:    book1.ID,
			BookBID:    book2.ID,
			Preference: PreferenceBookA,
		}

		err := db.Create(&comparison).Error
		if err != nil {
			t.Fatalf("Failed to create comparison: %v", err)
		}

		if comparison.CreatedAt.IsZero() {
			t.Error("CreatedAt should be set by BeforeCreate hook")
		}
	})

	t.Run("prevents same book comparison", func(t *testing.T) {
		comparison := Comparison{
			UserID:     user.ID,
			BookAID:    book1.ID,
			BookBID:    book1.ID, // Same book
			Preference: PreferenceBookA,
		}

		err := db.Create(&comparison).Error
		if err == nil {
			t.Error("Should not allow comparison of same book")
		}
		if err != gorm.ErrInvalidValue {
			t.Errorf("Expected ErrInvalidValue, got %v", err)
		}
	})
}

func TestComparisonTableName(t *testing.T) {
	comparison := Comparison{}
	expected := "comparisons"
	if comparison.TableName() != expected {
		t.Errorf("TableName() = %v, want %v", comparison.TableName(), expected)
	}
}

func TestComparisonIsValidPreference(t *testing.T) {
	tests := map[string]struct {
		preference ComparisonPreference
		want       bool
	}{
		"book_a valid": {
			preference: PreferenceBookA,
			want:       true,
		},
		"book_b valid": {
			preference: PreferenceBookB,
			want:       true,
		},
		"tie valid": {
			preference: PreferenceTie,
			want:       true,
		},
		"invalid preference": {
			preference: ComparisonPreference("invalid"),
			want:       false,
		},
		"empty preference": {
			preference: ComparisonPreference(""),
			want:       false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			comparison := Comparison{Preference: tc.preference}
			got := comparison.IsValidPreference()
			if got != tc.want {
				t.Errorf("IsValidPreference() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestComparisonRelationships(t *testing.T) {
	db := setupTestDB()

	// Create user and books
	user := User{
		Username:     "testuser",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	book1 := Book{
		Title:     "Test Book 1",
		Author:    "Test Author 1",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book1).Error; err != nil {
		t.Fatalf("Failed to create test book 1: %v", err)
	}

	book2 := Book{
		Title:     "Test Book 2",
		Author:    "Test Author 2",
		Genre:     "Non-Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book2).Error; err != nil {
		t.Fatalf("Failed to create test book 2: %v", err)
	}

	comparison := Comparison{
		UserID:     user.ID,
		BookAID:    book1.ID,
		BookBID:    book2.ID,
		Preference: PreferenceBookA,
	}
	if err := db.Create(&comparison).Error; err != nil {
		t.Fatalf("Failed to create comparison: %v", err)
	}

	// Test preloading relationships
	var comparisonWithRelations Comparison
	err := db.Preload("User").Preload("BookA").Preload("BookB").First(&comparisonWithRelations, comparison.ID).Error
	if err != nil {
		t.Fatalf("Failed to load comparison with relations: %v", err)
	}

	if comparisonWithRelations.User.ID != user.ID {
		t.Errorf("User ID = %v, want %v", comparisonWithRelations.User.ID, user.ID)
	}

	if comparisonWithRelations.BookA.ID != book1.ID {
		t.Errorf("BookA ID = %v, want %v", comparisonWithRelations.BookA.ID, book1.ID)
	}

	if comparisonWithRelations.BookB.ID != book2.ID {
		t.Errorf("BookB ID = %v, want %v", comparisonWithRelations.BookB.ID, book2.ID)
	}

	// Verify book details to ensure correct relationships
	if comparisonWithRelations.BookA.Title != "Test Book 1" {
		t.Errorf("BookA Title = %v, want 'Test Book 1'", comparisonWithRelations.BookA.Title)
	}

	if comparisonWithRelations.BookB.Title != "Test Book 2" {
		t.Errorf("BookB Title = %v, want 'Test Book 2'", comparisonWithRelations.BookB.Title)
	}
}

func TestComparisonPreferenceConstants(t *testing.T) {
	tests := map[string]struct {
		constant ComparisonPreference
		expected string
	}{
		"book_a": {
			constant: PreferenceBookA,
			expected: "book_a",
		},
		"book_b": {
			constant: PreferenceBookB,
			expected: "book_b",
		},
		"tie": {
			constant: PreferenceTie,
			expected: "tie",
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