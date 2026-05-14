package service

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"bookrank/internal/models"
	"bookrank/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles business logic for user operations
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(req *models.UserCreateRequest) (*models.User, error) {
	// Validate input
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) validateCreateRequest(req *models.UserCreateRequest) error {
	if strings.TrimSpace(req.Username) == "" {
		return errors.New("username is required")
	}
	if len(req.Username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	return s.userRepo.GetByUsername(username)
}

// UpdateUser updates a user's information
func (s *UserService) UpdateUser(id uint, req *models.UserUpdateRequest) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Validate update request
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Update fields
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Username != nil {
		user.Username = *req.Username
	}

	// Save changes
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) validateUpdateRequest(req *models.UserUpdateRequest) error {
	if req.Email != nil {
		if _, err := mail.ParseAddress(*req.Email); err != nil {
			return errors.New("invalid email format")
		}
	}
	if req.Username != nil {
		if len(*req.Username) < 3 {
			return errors.New("username must be at least 3 characters")
		}
	}
	return nil
}

// GetUsers retrieves all users with pagination
func (s *UserService) GetUsers(limit, offset int) ([]*models.User, error) {
	return s.userRepo.GetAll(limit, offset)
}

// SearchUsers searches for users
func (s *UserService) SearchUsers(query string, limit, offset int) ([]*models.User, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("search query cannot be empty")
	}
	return s.userRepo.Search(query, limit, offset)
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(id uint) error {
	return s.userRepo.Delete(id)
}