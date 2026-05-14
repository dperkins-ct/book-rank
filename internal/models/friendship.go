package models

import (
	"time"

	"gorm.io/gorm"
)

// Friendship represents a friendship relationship between two users
type Friendship struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	FriendID  uint      `gorm:"not null;index" json:"friend_id"`
	CreatedAt time.Time `gorm:"not null;index" json:"created_at"`

	// Relationships
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Friend User `gorm:"foreignKey:FriendID" json:"friend,omitempty"`
}

// TableName specifies the table name for the Friendship model
func (Friendship) TableName() string {
	return "friendships"
}

// BeforeCreate hook to set timestamp and validate friendship
func (f *Friendship) BeforeCreate(tx *gorm.DB) (err error) {
	f.CreatedAt = time.Now()

	// Validate that UserID and FriendID are different
	if f.UserID == f.FriendID {
		return gorm.ErrInvalidValue
	}

	return
}