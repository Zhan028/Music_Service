package middleware

import (
	"context"
	"github.com/Zhan028/Music_Service/playlistService/internal/middleware/jwt"
	"strings"
		
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	jwtManager      *jwt.JWTManager
	accessibleRoles map[string][]string
}

func NewAuthInterceptor(jwtManager *jwt.JWTManager, accessibleRoles map[string][]string) *AuthInterceptor {
	return &AuthInterceptor{jwtManager, accessibleRoles}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth if method doesn't require it
		if !i.needsAuth(info.FullMethod) {
			return handler(ctx, req)
		}

		// Get auth metadata
		userID, err := i.authorize(ctx)
		if err != nil {
			return nil, err
		}

		// Add userID to context
		ctx = context.WithValue(ctx, "user_id", userID)

		return handler(ctx, req)
	}
}

func (i *AuthInterceptor) needsAuth(method string) bool {
	if roles, ok := i.accessibleRoles[method]; ok {
		return len(roles) > 0
	}
	return true // Require auth by default
}

func (i *AuthInterceptor) authorize(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	authHeader := values[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", status.Errorf(codes.Unauthenticated, "invalid auth format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := i.jwtManager.Verify(token)
	if err != nil {
		if err == jwt.ErrExpiredToken {
			return "", status.Errorf(codes.Unauthenticated, "token has expired")
		}
		return "", status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	return claims.UserID, nil
}
