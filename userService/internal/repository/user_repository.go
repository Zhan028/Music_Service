package repository

import (
	"github.com/facelessEmptiness/user_service/internal/domain"
)

type UserRepository interface {
	Create(user *domain.User) (string, error)
	GetByEmail(email string) (*domain.User, error)
	GetByID(id string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id string) error
	List(page, limit int64) ([]*domain.User, error)
}
