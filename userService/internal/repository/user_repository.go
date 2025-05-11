package repository

import (
	"github.com/Zhan028/Music_Service/userService/internal/domain"
)

type UserRepository interface {
	Create(user *domain.User) (string, error)
	GetByEmail(email string) (*domain.User, error)
	GetByID(id string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id string) error
	List(page, limit int64) ([]*domain.User, error)
}
