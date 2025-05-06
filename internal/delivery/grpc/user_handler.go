package grpc

import (
	"context"
	"time"

	"github.com/facelessEmptiness/user_service/internal/domain"
	"github.com/facelessEmptiness/user_service/internal/usecase"
	pb "github.com/facelessEmptiness/user_service/proto"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServiceHandler struct {
	pb.UnimplementedUserServiceServer
	userUseCase *usecase.UserUseCase
	jwtSecret   []byte
	tokenExp    time.Duration
}

func NewUserServiceHandler(userUseCase *usecase.UserUseCase, jwtSecret string, tokenExp time.Duration) *UserServiceHandler {
	return &UserServiceHandler{
		userUseCase: userUseCase,
		jwtSecret:   []byte(jwtSecret),
		tokenExp:    tokenExp,
	}
}

// RegisterUser обрабатывает регистрацию пользователя
func (h *UserServiceHandler) RegisterUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	id, err := h.userUseCase.Register(user)
	if err != nil {
		if err == usecase.ErrEmailAlreadyExists {
			return nil, status.Errorf(codes.AlreadyExists, "пользователь с этим email уже существует")
		}
		return nil, status.Errorf(codes.Internal, "не удалось создать пользователя: %v", err)
	}

	return &pb.UserResponse{
		Id:      id,
		Message: "пользователь успешно зарегистрирован",
	}, nil
}

// AuthenticateUser обрабатывает аутентификацию пользователя и генерирует JWT токен
func (h *UserServiceHandler) AuthenticateUser(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	user, err := h.userUseCase.Login(req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "неверные учетные данные")
	}

	// Генерация JWT токена
	expiresAt := time.Now().Add(h.tokenExp)
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось сгенерировать токен")
	}

	return &pb.AuthResponse{
		Token: tokenString,
	}, nil
}

// GetUserProfile получает профиль пользователя по ID
func (h *UserServiceHandler) GetUserProfile(ctx context.Context, req *pb.UserID) (*pb.UserProfile, error) {
	user, err := h.userUseCase.GetByID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "пользователь не найден")
	}

	return &pb.UserProfile{
		Id:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
