package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"server/internal/models"
	"server/internal/proto"
	"server/internal/repository"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type UserServiceServer struct {
	proto.UnimplementedUserServiceServer
	userRepo repository.UserRepository
	jwtSecret string
}

func NewUserServiceServer(userRepo repository.UserRepository, jwtSecret string) *UserServiceServer {
	return &UserServiceServer{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *UserServiceServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	user := &models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     req.Role,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, status.Errorf(codes.AlreadyExists, "user with this email already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &proto.RegisterResponse{Message: "User registered successfully"}, nil
}

func (s *UserServiceServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	token, err := s.generateToken(user.ID, user.Role)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	return &proto.LoginResponse{Token: token}, nil
}

func (s *UserServiceServer) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	token, err := jwt.Parse(req.Token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token claims")
	}

	userID := claims["user_id"].(string)
	role := claims["role"].(string)

	return &proto.ValidateTokenResponse{
		IsValid: true,
		UserId:  userID,
		Role:    role,
	}, nil
}

func (s *UserServiceServer) generateToken(userID uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": fmt.Sprintf("%d", userID),
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Токен действует 24 часа
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}