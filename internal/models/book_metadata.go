package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// JSON is a custom type for handling JSON data in GORM
type JSON map[string]interface{}

// Value implements the driver.Valuer interface for JSON
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSON
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

// MetadataSource represents the source of external metadata
type MetadataSource string

const (
	SourceOpenLibrary MetadataSource = "openlibrary"
	SourceGoogleBooks MetadataSource = "googlebooks"
	SourceManual      MetadataSource = "manual"
)

// BookMetadata represents external metadata for books
type BookMetadata struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	BookID         uint           `gorm:"not null;index" json:"book_id"`
	ExternalID     string         `gorm:"size:255;index" json:"external_id,omitempty"`
	Source         MetadataSource `gorm:"not null;type:varchar(50);index" json:"source" validate:"required,oneof=openlibrary googlebooks manual"`
	AdditionalData JSON           `gorm:"type:jsonb" json:"additional_data,omitempty"`
	CreatedAt      time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"not null" json:"updated_at"`

	// Relationships
	Book Book `gorm:"foreignKey:BookID" json:"book,omitempty"`
}

// TableName specifies the table name for the BookMetadata model
func (BookMetadata) TableName() string {
	return "book_metadata"
}

// BeforeCreate hook to set timestamps
func (bm *BookMetadata) BeforeCreate(tx *gorm.DB) (err error) {
	now := time.Now()
	bm.CreatedAt = now
	bm.UpdatedAt = now
	return
}

// BeforeUpdate hook to update timestamp
func (bm *BookMetadata) BeforeUpdate(tx *gorm.DB) (err error) {
	bm.UpdatedAt = time.Now()
	return
}

// IsValidSource checks if the source value is valid
func (bm *BookMetadata) IsValidSource() bool {
	return bm.Source == SourceOpenLibrary || bm.Source == SourceGoogleBooks || bm.Source == SourceManual
}