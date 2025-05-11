package handler

import (
	"context"
	
	"github.com/Zhan028/Music_Service/userService/internal/domain"
	"github.com/Zhan028/Music_Service/userService/internal/usecase"
	"github.com/Zhan028/Music_Service/userService/proto"

	"time"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServiceHandler struct {
	proto.UnimplementedUserServiceServer
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

// RegisterUser handles user registration
func (h *UserServiceHandler) RegisterUser(ctx context.Context, req *proto.UserRequest) (*proto.UserResponse, error) {
	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	id, err := h.userUseCase.Register(user)
	if err != nil {
		if err == usecase.ErrEmailAlreadyExists {
			return nil, status.Errorf(codes.AlreadyExists, "user with this email already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &proto.UserResponse{
		Id:      id,
		Message: "user registered successfully",
	}, nil
}

// AuthenticateUser handles user authentication and generates JWT token
func (h *UserServiceHandler) AuthenticateUser(ctx context.Context, req *proto.AuthRequest) (*proto.AuthResponse, error) {
	user, err := h.userUseCase.Login(req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	// Generate JWT token
	expiresAt := time.Now().Add(h.tokenExp)
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	return &proto.AuthResponse{
		Token:     tokenString,
		UserId:    user.ID,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

// GetUserProfile retrieves user profile by ID
func (h *UserServiceHandler) GetUserProfile(ctx context.Context, req *proto.UserID) (*proto.UserProfile, error) {
	user, err := h.userUseCase.GetByID(req.Id)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return convertUserToProfile(user), nil
}

// GetUserByEmail retrieves user profile by email
func (h *UserServiceHandler) GetUserByEmail(ctx context.Context, req *proto.EmailRequest) (*proto.UserProfile, error) {
	user, err := h.userUseCase.GetByEmail(req.Email)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return convertUserToProfile(user), nil
}

// UpdateUserProfile updates user information
func (h *UserServiceHandler) UpdateUserProfile(ctx context.Context, req *proto.UpdateRequest) (*proto.UserResponse, error) {
	// First get existing user to preserve password
	existingUser, err := h.userUseCase.GetByID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	// Update fields
	user := &domain.User{
		ID:       req.Id,
		Name:     req.Name,
		Email:    req.Email,
		Password: existingUser.Password, // Preserve password
	}

	if err := h.userUseCase.Update(user); err != nil {
		if err == usecase.ErrEmailAlreadyExists {
			return nil, status.Errorf(codes.AlreadyExists, "email already in use")
		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &proto.UserResponse{
		Id:      req.Id,
		Message: "user updated successfully",
	}, nil
}

// ChangePassword handles password changes
func (h *UserServiceHandler) ChangePassword(ctx context.Context, req *proto.PasswordChangeRequest) (*proto.StatusResponse, error) {
	err := h.userUseCase.ChangePassword(req.Id, req.CurrentPassword, req.NewPassword)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			return nil, status.Errorf(codes.NotFound, "user not found")
		case usecase.ErrInvalidCredentials:
			return nil, status.Errorf(codes.PermissionDenied, "current password is incorrect")
		default:
			return nil, status.Errorf(codes.Internal, "failed to change password: %v", err)
		}
	}

	return &proto.StatusResponse{
		Success: true,
		Message: "password changed successfully",
	}, nil
}

// DeleteUser handles user deletion
func (h *UserServiceHandler) DeleteUser(ctx context.Context, req *proto.UserID) (*proto.StatusResponse, error) {
	err := h.userUseCase.Delete(req.Id)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &proto.StatusResponse{
		Success: true,
		Message: "user deleted successfully",
	}, nil
}

// ListUsers retrieves a paginated list of users
func (h *UserServiceHandler) ListUsers(ctx context.Context, req *proto.ListRequest) (*proto.UserList, error) {
	users, err := h.userUseCase.List(req.Page, req.Limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	userProfiles := make([]*proto.UserProfile, 0, len(users))
	for _, user := range users {
		userProfiles = append(userProfiles, convertUserToProfile(user))
	}

	return &proto.UserList{
		Users:      userProfiles,
		TotalCount: int64(len(userProfiles)),
		Page:       req.Page,
		Limit:      req.Limit,
	}, nil
}

// Helper function to convert domain.User to pb.UserProfile
func convertUserToProfile(user *domain.User) *proto.UserProfile {
	var createdAt, updatedAt int64

	// Handle time fields if they exist in your domain model
	if user.CreatedAt != nil {
		createdAt = user.CreatedAt.Unix()
	}
	if user.UpdatedAt != nil {
		updatedAt = user.UpdatedAt.Unix()
	}

	return &proto.UserProfile{
		Id:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
