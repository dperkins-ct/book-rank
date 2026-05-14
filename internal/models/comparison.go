package models

import (
	"time"

	"gorm.io/gorm"
)

// ComparisonPreference represents the user's preference in a book comparison
type ComparisonPreference string

const (
	PreferenceBookA ComparisonPreference = "book_a"
	PreferenceBookB ComparisonPreference = "book_b"
	PreferenceTie   ComparisonPreference = "tie"
)

// Comparison represents a pairwise comparison between two books by a user
type Comparison struct {
	ID         uint                 `gorm:"primaryKey" json:"id"`
	UserID     uint                 `gorm:"not null;index" json:"user_id"`
	BookAID    uint                 `gorm:"not null;index" json:"book_a_id"`
	BookBID    uint                 `gorm:"not null;index" json:"book_b_id"`
	Preference ComparisonPreference `gorm:"not null;type:varchar(10);check:preference IN ('book_a','book_b','tie')" json:"preference" validate:"required,oneof=book_a book_b tie"`
	CreatedAt  time.Time            `gorm:"not null;index" json:"created_at"`

	// Relationships
	User  User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	BookA Book `gorm:"foreignKey:BookAID" json:"book_a,omitempty"`
	BookB Book `gorm:"foreignKey:BookBID" json:"book_b,omitempty"`
}

// TableName specifies the table name for the Comparison model
func (Comparison) TableName() string {
	return "comparisons"
}

// BeforeCreate hook to set timestamp and validate comparison
func (c *Comparison) BeforeCreate(tx *gorm.DB) (err error) {
	c.CreatedAt = time.Now()

	// Validate that BookAID and BookBID are different
	if c.BookAID == c.BookBID {
		return gorm.ErrInvalidValue
	}

	return
}

// IsValidPreference checks if the preference value is valid
func (c *Comparison) IsValidPreference() bool {
	return c.Preference == PreferenceBookA || c.Preference == PreferenceBookB || c.Preference == PreferenceTie
}