package models

import (
	"time"

	"gorm.io/gorm"
)

// Ranking represents a user's rating for a specific book
type Ranking struct {
	UserID    uint      `gorm:"primaryKey;not null" json:"user_id"`
	BookID    uint      `gorm:"primaryKey;not null" json:"book_id"`
	Score     int       `gorm:"not null;default:1500;check:score >= 0 AND score <= 3000" json:"score" validate:"min=0,max=3000"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Book Book `gorm:"foreignKey:BookID" json:"book,omitempty"`
}

// TableName specifies the table name for the Ranking model
func (Ranking) TableName() string {
	return "rankings"
}

// BeforeCreate hook to set initial score and timestamp
func (r *Ranking) BeforeCreate(tx *gorm.DB) (err error) {
	if r.Score == 0 {
		r.Score = 1500 // Default ELO rating
	}
	r.UpdatedAt = time.Now()
	return
}

// BeforeUpdate hook to update timestamp
func (r *Ranking) BeforeUpdate(tx *gorm.DB) (err error) {
	r.UpdatedAt = time.Now()
	return
}