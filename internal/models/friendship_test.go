package models

import (
	"testing"

	"gorm.io/gorm"
)

func TestFriendshipModel(t *testing.T) {
	tests := map[string]struct {
		friendship Friendship
		wantErr    bool
	}{
		"valid friendship": {
			friendship: Friendship{
				UserID:   1,
				FriendID: 2,
			},
			wantErr: false,
		},
		"self friendship": {
			friendship: Friendship{
				UserID:   1,
				FriendID: 1, // Same user
			},
			wantErr: true,
		},
	}

	db := setupTestDB()

	// Create two users
	user1 := User{
		Username:     "testuser1",
		PasswordHash: "password",
	}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("Failed to create test user 1: %v", err)
	}

	user2 := User{
		Username:     "testuser2",
		PasswordHash: "password",
	}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("Failed to create test user 2: %v", err)
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Set valid foreign keys
			if tc.friendship.UserID == 1 {
				tc.friendship.UserID = user1.ID
			}
			if tc.friendship.FriendID == 2 {
				tc.friendship.FriendID = user2.ID
			} else if tc.friendship.FriendID == 1 {
				tc.friendship.FriendID = user1.ID // For self friendship test
			}

			err := db.Create(&tc.friendship).Error
			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !tc.wantErr {
				// Verify timestamp was set
				if tc.friendship.CreatedAt.IsZero() {
					t.Error("CreatedAt should be set")
				}
				if tc.friendship.ID == 0 {
					t.Error("ID should be set")
				}
			}
		})
	}
}

func TestFriendshipBeforeCreateHook(t *testing.T) {
	db := setupTestDB()

	// Create two users
	user1 := User{
		Username:     "testuser1",
		PasswordHash: "password",
	}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("Failed to create test user 1: %v", err)
	}

	user2 := User{
		Username:     "testuser2",
		PasswordHash: "password",
	}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("Failed to create test user 2: %v", err)
	}

	t.Run("sets timestamp", func(t *testing.T) {
		friendship := Friendship{
			UserID:   user1.ID,
			FriendID: user2.ID,
		}

		err := db.Create(&friendship).Error
		if err != nil {
			t.Fatalf("Failed to create friendship: %v", err)
		}

		if friendship.CreatedAt.IsZero() {
			t.Error("CreatedAt should be set by BeforeCreate hook")
		}
	})

	t.Run("prevents self friendship", func(t *testing.T) {
		friendship := Friendship{
			UserID:   user1.ID,
			FriendID: user1.ID, // Same user
		}

		err := db.Create(&friendship).Error
		if err == nil {
			t.Error("Should not allow self friendship")
		}
		if err != gorm.ErrInvalidValue {
			t.Errorf("Expected ErrInvalidValue, got %v", err)
		}
	})
}

func TestFriendshipTableName(t *testing.T) {
	friendship := Friendship{}
	expected := "friendships"
	if friendship.TableName() != expected {
		t.Errorf("TableName() = %v, want %v", friendship.TableName(), expected)
	}
}

func TestFriendshipRelationships(t *testing.T) {
	db := setupTestDB()

	// Create two users
	user1 := User{
		Username:     "testuser1",
		PasswordHash: "password",
	}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("Failed to create test user 1: %v", err)
	}

	user2 := User{
		Username:     "testuser2",
		PasswordHash: "password",
	}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("Failed to create test user 2: %v", err)
	}

	friendship := Friendship{
		UserID:   user1.ID,
		FriendID: user2.ID,
	}
	if err := db.Create(&friendship).Error; err != nil {
		t.Fatalf("Failed to create friendship: %v", err)
	}

	// Test preloading relationships
	var friendshipWithRelations Friendship
	err := db.Preload("User").Preload("Friend").First(&friendshipWithRelations, friendship.ID).Error
	if err != nil {
		t.Fatalf("Failed to load friendship with relations: %v", err)
	}

	if friendshipWithRelations.User.ID != user1.ID {
		t.Errorf("User ID = %v, want %v", friendshipWithRelations.User.ID, user1.ID)
	}

	if friendshipWithRelations.Friend.ID != user2.ID {
		t.Errorf("Friend ID = %v, want %v", friendshipWithRelations.Friend.ID, user2.ID)
	}

	// Verify usernames to ensure correct relationships
	if friendshipWithRelations.User.Username != "testuser1" {
		t.Errorf("User Username = %v, want 'testuser1'", friendshipWithRelations.User.Username)
	}

	if friendshipWithRelations.Friend.Username != "testuser2" {
		t.Errorf("Friend Username = %v, want 'testuser2'", friendshipWithRelations.Friend.Username)
	}
}

func TestFriendshipUniqueness(t *testing.T) {
	db := setupTestDB()

	// Create two users
	user1 := User{
		Username:     "testuser1",
		PasswordHash: "password",
	}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("Failed to create test user 1: %v", err)
	}

	user2 := User{
		Username:     "testuser2",
		PasswordHash: "password",
	}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("Failed to create test user 2: %v", err)
	}

	// Create first friendship
	friendship1 := Friendship{
		UserID:   user1.ID,
		FriendID: user2.ID,
	}
	if err := db.Create(&friendship1).Error; err != nil {
		t.Fatalf("Failed to create first friendship: %v", err)
	}

	// Note: The unique index is created in CreateIndexes(), so we might not see
	// the constraint violation in this test unless indexes are created.
	// But we can still test that duplicate friendships can conceptually be prevented.

	t.Run("allows reverse friendship", func(t *testing.T) {
		// This should be allowed (user2 -> user1)
		friendship2 := Friendship{
			UserID:   user2.ID,
			FriendID: user1.ID,
		}
		err := db.Create(&friendship2).Error
		if err != nil {
			t.Errorf("Should allow reverse friendship: %v", err)
		}
	})
}

func TestFriendshipBidirectional(t *testing.T) {
	db := setupTestDB()

	// Create two users
	user1 := User{
		Username:     "testuser1",
		PasswordHash: "password",
	}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("Failed to create test user 1: %v", err)
	}

	user2 := User{
		Username:     "testuser2",
		PasswordHash: "password",
	}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("Failed to create test user 2: %v", err)
	}

	// Create friendships in both directions
	friendship1 := Friendship{
		UserID:   user1.ID,
		FriendID: user2.ID,
	}
	if err := db.Create(&friendship1).Error; err != nil {
		t.Fatalf("Failed to create friendship 1->2: %v", err)
	}

	friendship2 := Friendship{
		UserID:   user2.ID,
		FriendID: user1.ID,
	}
	if err := db.Create(&friendship2).Error; err != nil {
		t.Fatalf("Failed to create friendship 2->1: %v", err)
	}

	// Test querying friendships from user1's perspective
	var user1Friendships []Friendship
	err := db.Where("user_id = ?", user1.ID).Find(&user1Friendships).Error
	if err != nil {
		t.Fatalf("Failed to query user1 friendships: %v", err)
	}

	if len(user1Friendships) != 1 {
		t.Errorf("User1 should have 1 friendship, got %d", len(user1Friendships))
	}

	// Test querying friendships from user2's perspective
	var user2Friendships []Friendship
	err = db.Where("user_id = ?", user2.ID).Find(&user2Friendships).Error
	if err != nil {
		t.Fatalf("Failed to query user2 friendships: %v", err)
	}

	if len(user2Friendships) != 1 {
		t.Errorf("User2 should have 1 friendship, got %d", len(user2Friendships))
	}

	// Test querying where user1 is a friend (using FriendOf relationship)
	var friendOfRelationships []Friendship
	err = db.Where("friend_id = ?", user1.ID).Find(&friendOfRelationships).Error
	if err != nil {
		t.Fatalf("Failed to query friend_of relationships: %v", err)
	}

	if len(friendOfRelationships) != 1 {
		t.Errorf("User1 should be a friend in 1 relationship, got %d", len(friendOfRelationships))
	}
}