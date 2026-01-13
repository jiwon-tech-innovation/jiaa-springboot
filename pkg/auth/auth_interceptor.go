package auth

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JWTInterceptor validates JWT tokens for gRPC requests
type JWTInterceptor struct {
	secretKey []byte
}

// NewJWTInterceptor creates a new interceptor with the given secret
func NewJWTInterceptor(secret string) *JWTInterceptor {
	return &JWTInterceptor{
		secretKey: []byte(secret),
	}
}

// NewJWTInterceptorFromEnv creates an interceptor using the JWT_SECRET env var
func NewJWTInterceptorFromEnv() *JWTInterceptor {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("[AUTH] Warning: JWT_SECRET not set, using default (insecure)")
		secret = "jiaa-super-secret-key-for-jwt-token-generation-must-be-at-least-256-bits-long"
	}
	return NewJWTInterceptor(secret)
}

// extractToken extracts the Bearer token from gRPC metadata
func (i *JWTInterceptor) extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata not provided")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization header not provided")
	}

	// Expected format: "Bearer <token>"
	parts := strings.SplitN(authHeader[0], " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	return parts[1], nil
}

// validateToken validates the JWT token
func (i *JWTInterceptor) validateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Error(codes.Unauthenticated, "unexpected signing method")
		}
		return i.secretKey, nil
	})

	if err != nil {
		log.Printf("[AUTH] Token validation failed: %v", err)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	if !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "token is not valid")
	}

	return token, nil
}

// UnaryInterceptor returns a gRPC unary interceptor for JWT validation
func (i *JWTInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth for health check
		if strings.Contains(info.FullMethod, "Health") {
			return handler(ctx, req)
		}

		token, err := i.extractToken(ctx)
		if err != nil {
			return nil, err
		}

		_, err = i.validateToken(token)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// StreamInterceptor returns a gRPC stream interceptor for JWT validation
func (i *JWTInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Skip auth for health check
		if strings.Contains(info.FullMethod, "Health") {
			return handler(srv, ss)
		}

		token, err := i.extractToken(ss.Context())
		if err != nil {
			return err
		}

		_, err = i.validateToken(token)
		if err != nil {
			return err
		}

		return handler(srv, ss)
	}
}
