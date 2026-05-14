package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the BookRank application
type User struct {
	gorm.Model
	Username     string `gorm:"unique;not null;size:20;check:length(username) >= 3" json:"username"`
	Email        string `gorm:"size:100" json:"email,omitempty"` // Optional for demo
	PasswordHash string `gorm:"not null;size:255" json:"-"`
}

// UserRegisterRequest represents the request body for user registration
type UserRegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Password string `json:"password" validate:"required"`
}

// UserCreateRequest represents the request body for creating a user
type UserCreateRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// UserUpdateRequest represents the request body for updating a user
type UserUpdateRequest struct {
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=20"`
}

// UserLoginRequest represents the request body for user login
type UserLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UserResponse represents the response body for user operations
type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// TokenResponse represents the response body for authentication operations
type TokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresAt    int64  `json:"expires_at"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to set timestamps
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	return
}

// BeforeUpdate hook to update timestamp
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	u.UpdatedAt = time.Now()
	return
}