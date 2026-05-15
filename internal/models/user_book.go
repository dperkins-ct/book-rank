package models

import (
	"time"

	"gorm.io/gorm"
)

// UserBookStatus represents the user's relationship with a book
type UserBookStatus string

const (
	StatusWantToRead UserBookStatus = "want_to_read"
	StatusReading    UserBookStatus = "reading"
	StatusRead       UserBookStatus = "read"
	StatusDNF        UserBookStatus = "did_not_finish" // Did Not Finish
)

// UserBook represents the relationship between a user and a book
type UserBook struct {
	UserID    uint           `gorm:"primaryKey;not null" json:"user_id"`
	BookID    uint           `gorm:"primaryKey;not null" json:"book_id"`
	Status    UserBookStatus `gorm:"not null;type:varchar(20);default:'read'" json:"status" validate:"required,oneof=want_to_read reading read did_not_finish"`
	Rating    *int           `gorm:"check:rating >= 1 AND rating <= 5" json:"rating,omitempty" validate:"omitempty,min=1,max=5"`
	Review    string         `gorm:"type:text" json:"review,omitempty"`
	AddedAt   time.Time      `gorm:"not null" json:"added_at"`
	ReadAt    *time.Time     `json:"read_at,omitempty"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Book Book `gorm:"foreignKey:BookID" json:"book,omitempty"`
}

// TableName specifies the table name for the UserBook model
func (UserBook) TableName() string {
	return "user_books"
}

// BeforeCreate hook to set timestamps
func (ub *UserBook) BeforeCreate(tx *gorm.DB) (err error) {
	ub.AddedAt = time.Now()

	// Set ReadAt if status is 'read'
	if ub.Status == StatusRead && ub.ReadAt == nil {
		now := time.Now()
		ub.ReadAt = &now
	}

	return
}

// BeforeUpdate hook to update ReadAt timestamp
func (ub *UserBook) BeforeUpdate(tx *gorm.DB) (err error) {
	// Set ReadAt if status changed to 'read'
	if ub.Status == StatusRead && ub.ReadAt == nil {
		now := time.Now()
		ub.ReadAt = &now
	}

	// Clear ReadAt if status is no longer 'read'
	if ub.Status != StatusRead {
		ub.ReadAt = nil
	}

	return
}

// IsValidStatus checks if the status value is valid
func (ub *UserBook) IsValidStatus() bool {
	return ub.Status == StatusWantToRead ||
		   ub.Status == StatusReading ||
		   ub.Status == StatusRead ||
		   ub.Status == StatusDNF
}