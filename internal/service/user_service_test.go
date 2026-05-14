package service

import (
	"bookrank/internal/models"
	"bookrank/internal/repository"
	"testing"

	"gorm.io/gorm"
)

// MockUserRepository for testing
type MockUserRepository struct {
	users     map[uint]*models.User
	usersByEmail map[string]*models.User
	usersByUsername map[string]*models.User
	nextID    uint
	createErr error
	getErr    error
	updateErr error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:           make(map[uint]*models.User),
		usersByEmail:    make(map[string]*models.User),
		usersByUsername: make(map[string]*models.User),
		nextID:          1,
	}
}

func (m *MockUserRepository) Create(user *models.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	m.usersByEmail[user.Email] = user
	m.usersByUsername[user.Username] = user
	return nil
}

func (m *MockUserRepository) GetByID(id uint) (*models.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	user, exists := m.users[id]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	user, exists := m.usersByEmail[email]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	user, exists := m.usersByUsername[username]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (m *MockUserRepository) Update(user *models.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Delete(id uint) error {
	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) GetAll(limit, offset int) ([]*models.User, error) {
	var result []*models.User
	for _, user := range m.users {
		result = append(result, user)
	}
	return result, nil
}

func (m *MockUserRepository) Search(query string, limit, offset int) ([]*models.User, error) {
	return m.GetAll(limit, offset)
}

func (m *MockUserRepository) UpdateLastLogin(userID uint) error {
	return nil
}

func (m *MockUserRepository) GetUserStats(userID uint) (*repository.UserStats, error) {
	return &repository.UserStats{UserID: userID}, nil
}

func TestUserService_CreateUser(t *testing.T) {
	tests := map[string]struct {
		request *models.UserCreateRequest
		repoErr error
		wantErr bool
	}{
		"valid_user": {
			request: &models.UserCreateRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		"duplicate_email": {
			request: &models.UserCreateRequest{
				Username: "testuser2",
				Email:    "test@example.com",
				Password: "password123",
			},
			repoErr: gorm.ErrDuplicatedKey,
			wantErr: true,
		},
		"empty_username": {
			request: &models.UserCreateRequest{
				Username: "",
				Email:    "test2@example.com",
				Password: "password123",
			},
			wantErr: true,
		},
		"invalid_email": {
			request: &models.UserCreateRequest{
				Username: "testuser3",
				Email:    "invalid-email",
				Password: "password123",
			},
			wantErr: true,
		},
		"weak_password": {
			request: &models.UserCreateRequest{
				Username: "testuser4",
				Email:    "test4@example.com",
				Password: "123",
			},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockRepo := NewMockUserRepository()
			mockRepo.createErr = tc.repoErr
			service := NewUserService(mockRepo)

			user, err := service.CreateUser(tc.request)
			if (err != nil) != tc.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if user.Username != tc.request.Username {
					t.Errorf("Expected username %s, got %s", tc.request.Username, user.Username)
				}
				if user.Email != tc.request.Email {
					t.Errorf("Expected email %s, got %s", tc.request.Email, user.Email)
				}
				if user.PasswordHash == tc.request.Password {
					t.Error("Password should be hashed")
				}
			}
		})
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewUserService(mockRepo)

	// Create test user
	testUser := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}
	testUser.ID = 1
	mockRepo.users[1] = testUser

	tests := map[string]struct {
		id      uint
		repoErr error
		wantErr bool
	}{
		"existing_user": {
			id:      1,
			wantErr: false,
		},
		"nonexistent_user": {
			id:      999,
			wantErr: true,
		},
		"repository_error": {
			id:      1,
			repoErr: gorm.ErrInvalidDB,
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockRepo.getErr = tc.repoErr
			user, err := service.GetUserByID(tc.id)
			mockRepo.getErr = nil

			if (err != nil) != tc.wantErr {
				t.Errorf("GetUserByID() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr && user.ID != tc.id {
				t.Errorf("Expected user ID %d, got %d", tc.id, user.ID)
			}
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewUserService(mockRepo)

	// Create test user
	testUser := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}
	testUser.ID = 1
	mockRepo.users[1] = testUser

	newEmail := "newemail@example.com"
	tests := map[string]struct {
		id      uint
		request *models.UserUpdateRequest
		wantErr bool
	}{
		"valid_update": {
			id: 1,
			request: &models.UserUpdateRequest{
				Email: &newEmail,
			},
			wantErr: false,
		},
		"invalid_email": {
			id: 1,
			request: &models.UserUpdateRequest{
				Email: func() *string { s := "invalid-email"; return &s }(),
			},
			wantErr: true,
		},
		"nonexistent_user": {
			id: 999,
			request: &models.UserUpdateRequest{
				Email: &newEmail,
			},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			user, err := service.UpdateUser(tc.id, tc.request)
			if (err != nil) != tc.wantErr {
				t.Errorf("UpdateUser() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr && tc.request.Email != nil && user.Email != *tc.request.Email {
				t.Errorf("Expected email %s, got %s", *tc.request.Email, user.Email)
			}
		})
	}
}

func TestUserService_GetUsers(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewUserService(mockRepo)

	// Add test users
	for i := 1; i <= 5; i++ {
		user := &models.User{
			Username:     "testuser" + string(rune('0'+i)),
			Email:        "test" + string(rune('0'+i)) + "@example.com",
			PasswordHash: "hashedpassword",
		}
		user.ID = uint(i)
		mockRepo.users[uint(i)] = user
	}

	tests := map[string]struct {
		limit   int
		offset  int
		wantErr bool
	}{
		"valid_pagination": {
			limit:   10,
			offset:  0,
			wantErr: false,
		},
		"zero_limit": {
			limit:   0,
			offset:  0,
			wantErr: false,
		},
		"negative_offset": {
			limit:   10,
			offset:  -1,
			wantErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			users, err := service.GetUsers(tc.limit, tc.offset)
			if (err != nil) != tc.wantErr {
				t.Errorf("GetUsers() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr && users == nil {
				t.Error("Expected users slice to be non-nil")
			}
		})
	}
}

func TestUserService_SearchUsers(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewUserService(mockRepo)

	// Add test users
	for i := 1; i <= 3; i++ {
		user := &models.User{
			Username:     "testuser" + string(rune('0'+i)),
			Email:        "test" + string(rune('0'+i)) + "@example.com",
			PasswordHash: "hashedpassword",
		}
		user.ID = uint(i)
		mockRepo.users[uint(i)] = user
		mockRepo.usersByUsername[user.Username] = user
		mockRepo.usersByEmail[user.Email] = user
	}

	tests := map[string]struct {
		query   string
		limit   int
		offset  int
		wantErr bool
	}{
		"valid_search": {
			query:   "test",
			limit:   10,
			offset:  0,
			wantErr: false,
		},
		"empty_query": {
			query:   "",
			limit:   10,
			offset:  0,
			wantErr: true,
		},
		"whitespace_query": {
			query:   "   ",
			limit:   10,
			offset:  0,
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			users, err := service.SearchUsers(tc.query, tc.limit, tc.offset)
			if (err != nil) != tc.wantErr {
				t.Errorf("SearchUsers() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr && users == nil {
				t.Error("Expected users slice to be non-nil")
			}
		})
	}
}