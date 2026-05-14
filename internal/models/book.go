package models

import (
	"time"

	"gorm.io/gorm"
)

// Book represents a book in the BookRank application
type Book struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Title           string         `gorm:"not null;size:255;index" json:"title" validate:"required,max=255"`
	Author          string         `gorm:"not null;size:255;index" json:"author" validate:"required,max=255"`
	Genre           string         `gorm:"size:100;index" json:"genre" validate:"omitempty,max=100"`
	PublicationDate *time.Time     `gorm:"index" json:"publication_date,omitempty"`
	Description     string         `gorm:"type:text" json:"description,omitempty"`
	CreatedBy       uint           `gorm:"not null;index" json:"created_by"`
	CreatedAt       time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Creator    User           `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Rankings   []Ranking      `gorm:"foreignKey:BookID" json:"rankings,omitempty"`
	Metadata   []BookMetadata `gorm:"foreignKey:BookID" json:"metadata,omitempty"`

	// Comparison relationships
	ComparisonsAsA []Comparison `gorm:"foreignKey:BookAID" json:"-"`
	ComparisonsAsB []Comparison `gorm:"foreignKey:BookBID" json:"-"`
}

// TableName specifies the table name for the Book model
func (Book) TableName() string {
	return "books"
}

// BeforeCreate hook to set timestamps
func (b *Book) BeforeCreate(tx *gorm.DB) (err error) {
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now
	return
}

// BeforeUpdate hook to update timestamp
func (b *Book) BeforeUpdate(tx *gorm.DB) (err error) {
	b.UpdatedAt = time.Now()
	return
}