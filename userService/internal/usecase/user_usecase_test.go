package usecase_test

import (
	"errors"
	"github.com/Zhan028/Music_Service/userService/internal/domain"
	"github.com/Zhan028/Music_Service/userService/internal/usecase"
	"github.com/stretchr/testify/assert"
	"testing"
)

type fakeUserRepo struct {
	users map[string]*domain.User
}

func (f *fakeUserRepo) GetByID(id string) (*domain.User, error) {
	user, exists := f.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (f *fakeUserRepo) GetByEmail(email string) (*domain.User, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeUserRepo) Create(user *domain.User) (string, error) {
	return "", errors.New("not implemented")
}

func (f *fakeUserRepo) Update(user *domain.User) error {
	return errors.New("not implemented")
}
func (f *fakeUserRepo) Delete(id string) error {
	return errors.New("not implemented")
}

func (f *fakeUserRepo) List(page int64, limit int64) ([]*domain.User, error) {
	return nil, errors.New("not implemented")
}

func TestGetByID(t *testing.T) {
	repo := &fakeUserRepo{
		users: map[string]*domain.User{
			"123": {ID: "123", Name: "John"},
			"456": {ID: "456", Name: "Alice"},
		},
	}
	uc := usecase.NewUserUseCase(repo, nil)

	tests := []struct {
		name     string
		inputID  string
		wantUser *domain.User
		wantErr  error
	}{
		{
			name:     "existing user",
			inputID:  "123",
			wantUser: &domain.User{ID: "123", Name: "John"},
			wantErr:  nil,
		},
		{
			name:     "another existing user",
			inputID:  "456",
			wantUser: &domain.User{ID: "456", Name: "Alice"},
			wantErr:  nil,
		},
		{
			name:     "user not found",
			inputID:  "999",
			wantUser: nil,
			wantErr:  usecase.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := uc.GetByID(tt.inputID)

			assert.Equal(t, tt.wantUser, user)

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
