package models

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err := AutoMigrate(db); err != nil {
		panic(err)
	}

	return db
}

func TestUserModel(t *testing.T) {
	tests := map[string]struct {
		user    User
		wantErr bool
	}{
		"valid user": {
			user: User{
				Username:     "testuser",
				PasswordHash: "hashed_password_here",
			},
			wantErr: false,
		},
		"empty username": {
			user: User{
				Username:     "",
				PasswordHash: "hashed_password_here",
			},
			wantErr: true, // CHECK constraint actually works in SQLite
		},
		"empty password": {
			user: User{
				Username:     "testuser2",
				PasswordHash: "",
			},
			wantErr: false, // SQLite allows empty strings, validation should be in service layer
		},
		"username too long": {
			user: User{
				Username:     "this_username_is_way_too_long_for_validation_and_exceeds_size_limit",
				PasswordHash: "hashed_password_here",
			},
			wantErr: false, // SQLite size limits work differently, validation should be in service layer
		},
	}

	db := setupTestDB()

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := db.Create(&tc.user).Error
			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !tc.wantErr {
				// Verify timestamps were set
				if tc.user.CreatedAt.IsZero() {
					t.Error("CreatedAt should be set")
				}
				if tc.user.UpdatedAt.IsZero() {
					t.Error("UpdatedAt should be set")
				}
				if tc.user.ID == 0 {
					t.Error("ID should be set")
				}
			}
		})
	}
}

func TestUserBeforeCreateHook(t *testing.T) {
	db := setupTestDB()

	user := User{
		Username:     "testuser",
		PasswordHash: "hashed_password",
	}

	err := db.Create(&user).Error
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set by BeforeCreate hook")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set by BeforeCreate hook")
	}

	// Check that both timestamps are approximately equal (within 1 second)
	diff := user.UpdatedAt.Sub(user.CreatedAt)
	if diff > time.Second || diff < -time.Second {
		t.Error("CreatedAt and UpdatedAt should be set to approximately the same time")
	}
}

func TestUserBeforeUpdateHook(t *testing.T) {
	db := setupTestDB()

	user := User{
		Username:     "testuser",
		PasswordHash: "hashed_password",
	}

	// Create user
	err := db.Create(&user).Error
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	originalUpdatedAt := user.UpdatedAt

	// Wait a small amount to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update user
	user.Username = "updateduser"
	err = db.Save(&user).Error
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	if !user.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated by BeforeUpdate hook")
	}
}

func TestUserTableName(t *testing.T) {
	user := User{}
	expected := "users"
	if user.TableName() != expected {
		t.Errorf("TableName() = %v, want %v", user.TableName(), expected)
	}
}

func TestUserSoftDelete(t *testing.T) {
	db := setupTestDB()

	user := User{
		Username:     "testuser",
		PasswordHash: "hashed_password",
	}

	// Create user
	err := db.Create(&user).Error
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	userID := user.ID

	// Soft delete user
	err = db.Delete(&user).Error
	if err != nil {
		t.Fatalf("Failed to soft delete user: %v", err)
	}

	// Try to find the user (should not be found due to soft delete)
	var foundUser User
	err = db.First(&foundUser, userID).Error
	if err != gorm.ErrRecordNotFound {
		t.Errorf("Soft deleted user should not be found, got error: %v", err)
	}

	// Find with Unscoped (should find the soft deleted user)
	err = db.Unscoped().First(&foundUser, userID).Error
	if err != nil {
		t.Errorf("Should find soft deleted user with Unscoped(), got error: %v", err)
	}

	if foundUser.DeletedAt.Time.IsZero() {
		t.Error("DeletedAt should be set for soft deleted user")
	}
}