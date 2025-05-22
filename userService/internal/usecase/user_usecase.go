package usecase

import (
	"errors"
	"log"

	"github.com/Zhan028/Music_Service/userService/internal/domain"
	"github.com/Zhan028/Music_Service/userService/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserUseCase struct {
	repo        repository.UserRepository
	emailSender EmailSender
}

func NewUserUseCase(r repository.UserRepository, sender EmailSender) *UserUseCase {
	return &UserUseCase{
		repo:        r,
		emailSender: sender,
	}
}

// Register creates a new user with hashed password and sends a welcome email
func (u *UserUseCase) Register(user *domain.User) (string, error) {
	// Check if user with this email already exists
	existingUser, err := u.repo.GetByEmail(user.Email)
	if err == nil && existingUser != nil {
		log.Printf("Регистрация: email %s уже существует", user.Email)
		return "", ErrEmailAlreadyExists
	}

	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Регистрация: ошибка хэширования пароля для email %s: %v", user.Email, err)
		return "", err
	}
	user.Password = string(hashedPassword)

	id, err := u.repo.Create(user)
	if err != nil {
		log.Printf("Регистрация: ошибка создания пользователя с email %s: %v", user.Email, err)
		return "", err
	}

	log.Printf("Регистрация: пользователь с email %s успешно создан с ID %s", user.Email, id)

	// Отправка приветственного письма
	subject := "Добро пожаловать в Music Service!"
	body := "Здравствуйте, " + user.Name + "!<br><br>Спасибо за регистрацию в нашем сервисе."

	err = u.emailSender.SendEmail(user.Email, subject, body)
	if err != nil {
		log.Printf("Регистрация: ошибка при отправке email на %s: %v", user.Email, err)
	} else {
		log.Printf("Регистрация: письмо успешно отправлено на %s", user.Email)
	}

	return id, nil
}

// Login authenticates a user
func (u *UserUseCase) Login(email, password string) (*domain.User, error) {
	user, err := u.repo.GetByEmail(email)
	if err != nil {
		log.Printf("Логин: пользователь с email %s не найден", email)
		return nil, ErrInvalidCredentials
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Printf("Логин: неверный пароль для email %s", email)
		return nil, ErrInvalidCredentials
	}

	log.Printf("Логин: пользователь с email %s успешно аутентифицирован", email)
	return user, nil
}

// GetByEmail retrieves a user by email
func (u *UserUseCase) GetByEmail(email string) (*domain.User, error) {
	user, err := u.repo.GetByEmail(email)
	if err != nil {
		log.Printf("GetByEmail: пользователь с email %s не найден", email)
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetByID retrieves a user by ID
func (u *UserUseCase) GetByID(id string) (*domain.User, error) {
	user, err := u.repo.GetByID(id)
	if err != nil {
		log.Printf("GetByID: пользователь с ID %s не найден", id)
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Update updates an existing user
func (u *UserUseCase) Update(user *domain.User) error {
	// Check if user exists
	existingUser, err := u.repo.GetByID(user.ID)
	if err != nil {
		log.Printf("Update: пользователь с ID %s не найден", user.ID)
		return ErrUserNotFound
	}

	// If password is being updated, hash it
	if user.Password != "" && user.Password != existingUser.Password {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Update: ошибка хэширования пароля для пользователя ID %s: %v", user.ID, err)
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
			log.Printf("Update: email %s уже существует (попытка обновления пользователя ID %s)", user.Email, user.ID)
			return ErrEmailAlreadyExists
		}
	}

	err = u.repo.Update(user)
	if err != nil {
		log.Printf("Update: ошибка обновления пользователя ID %s: %v", user.ID, err)
		return err
	}

	log.Printf("Update: пользователь ID %s успешно обновлен", user.ID)
	return nil
}

// Delete removes a user by ID
func (u *UserUseCase) Delete(id string) error {
	// Check if user exists
	_, err := u.repo.GetByID(id)
	if err != nil {
		log.Printf("Delete: пользователь с ID %s не найден", id)
		return ErrUserNotFound
	}

	err = u.repo.Delete(id)
	if err != nil {
		log.Printf("Delete: ошибка удаления пользователя ID %s: %v", id, err)
		return err
	}

	log.Printf("Delete: пользователь ID %s успешно удален", id)
	return nil
}

// List retrieves a paginated list of users
func (u *UserUseCase) List(page, limit int64) ([]*domain.User, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10 // Default limit
	}

	users, err := u.repo.List(page, limit)
	if err != nil {
		log.Printf("List: ошибка получения списка пользователей: %v", err)
		return nil, err
	}

	log.Printf("List: получено %d пользователей (страница %d, лимит %d)", len(users), page, limit)
	return users, nil
}

// ChangePassword updates a user's password
func (u *UserUseCase) ChangePassword(id, currentPassword, newPassword string) error {
	user, err := u.repo.GetByID(id)
	if err != nil {
		log.Printf("ChangePassword: пользователь с ID %s не найден", id)
		return ErrUserNotFound
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword))
	if err != nil {
		log.Printf("ChangePassword: неверный текущий пароль для пользователя ID %s", id)
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ChangePassword: ошибка хэширования нового пароля для пользователя ID %s: %v", id, err)
		return err
	}

	// Update user with new password
	user.Password = string(hashedPassword)
	err = u.repo.Update(user)
	if err != nil {
		log.Printf("ChangePassword: ошибка обновления пароля для пользователя ID %s: %v", id, err)
		return err
	}

	log.Printf("ChangePassword: пароль успешно изменен для пользователя ID %s", id)
	return nil
}
