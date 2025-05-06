package usecase

import (
	"errors"
	"github.com/facelessEmptiness/user_service/internal/domain"
	"github.com/facelessEmptiness/user_service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(r repository.UserRepository) *UserUseCase {
	return &UserUseCase{repo: r}
}

// Register creates a new user with hashed password
func (u *UserUseCase) Register(user *domain.User) (string, error) {
	// Check if user with this email already exists
	existingUser, err := u.repo.GetByEmail(user.Email)
	if err == nil && existingUser != nil {
		return "", ErrEmailAlreadyExists
	}

	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	user.Password = string(hashedPassword)

	return u.repo.Create(user)
}

// Login authenticates a user
func (u *UserUseCase) Login(email, password string) (*domain.User, error) {
	user, err := u.repo.GetByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (u *UserUseCase) GetByEmail(email string) (*domain.User, error) {
	user, err := u.repo.GetByEmail(email)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetByID retrieves a user by ID
func (u *UserUseCase) GetByID(id string) (*domain.User, error) {
	user, err := u.repo.GetByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Update updates an existing user
func (u *UserUseCase) Update(user *domain.User) error {
	// Check if user exists
	existingUser, err := u.repo.GetByID(user.ID)
	if err != nil {
		return ErrUserNotFound
	}

	// If password is being updated, hash it
	if user.Password != "" && user.Password != existingUser.Password {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	} else {
		// Keep the original password if not changed
		user.Password = existingUser.Password
	}

	// Check if updating email and if new email already exists
	if user.Email != existingUser.Email {
		emailUser, err := u.repo.GetByEmail(user.Email)
		if err == nil && emailUser != nil && emailUser.ID != user.ID {
			return ErrEmailAlreadyExists
		}
	}

	return u.repo.Update(user)
}

// Delete removes a user by ID
func (u *UserUseCase) Delete(id string) error {
	// Check if user exists
	_, err := u.repo.GetByID(id)
	if err != nil {
		return ErrUserNotFound
	}

	return u.repo.Delete(id)
}

// List retrieves a paginated list of users
func (u *UserUseCase) List(page, limit int64) ([]*domain.User, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10 // Default limit
	}

	return u.repo.List(page, limit)
}

// ChangePassword updates a user's password
func (u *UserUseCase) ChangePassword(id, currentPassword, newPassword string) error {
	user, err := u.repo.GetByID(id)
	if err != nil {
		return ErrUserNotFound
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user with new password
	user.Password = string(hashedPassword)
	return u.repo.Update(user)
}
