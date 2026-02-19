package service

import (
	"context"
	"ecommerce/services/auth/internal/domain"
	"ecommerce/services/auth/internal/repository"
	"ecommerce/services/auth/internal/utils"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, name, email, password, role, provider, providerId string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, error)
	//Refresh(ctx context.Context, refreshToken string) (*domain.User, error)
	//Logout(ctx context.Context) error
	SaveRefreshToken(ctx context.Context, userID, hashedRefreshToken, familyId string) error
}

type authService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
}

func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository) AuthService {
	return &authService{userRepo: userRepo, tokenRepo: tokenRepo}
}

func (a *authService) Register(ctx context.Context, name, email, password, role, provider, providerId string) (*domain.User, error) {
	res, err := a.userRepo.GetUserByEmail(ctx, email)

	if err != nil {
		return nil, fmt.Errorf("service: failed to check email existence: %w", err)
	}

	if res != nil {
		return nil, errors.New("service: email already exists")
	}

	var hashedPassword string
	hashedPassword, err = utils.HashPassword(password)

	if err != nil {
		return nil, err
	}

	user := domain.User{Name: name, Email: email, Password: hashedPassword, Role: role, Provider: provider, ProviderID: providerId}

	err = a.userRepo.CreateUser(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("service: failed to create user: %w", err)
	}

	return &user, nil
}

func (a *authService) Login(ctx context.Context, email, password string) (*domain.User, error) {
	res, err := a.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("service: failed to check email existence: %w", err)
	}
	if res == nil {
		return nil, errors.New("service: email does not exist")
	}

	err = bcrypt.CompareHashAndPassword([]byte(res.Password), []byte(password))

	if err != nil {
		return nil, errors.New("service: invalid password")
	}

	return res, nil
}

func (a *authService) SaveRefreshToken(ctx context.Context, userID, hashedRefreshToken, familyID string) error {
	refreshToken := domain.NewToken(userID, hashedRefreshToken, familyID, time.Now().Add(time.Hour*24*7))

	err := a.tokenRepo.Create(ctx, refreshToken)

	if err != nil {
		return fmt.Errorf("service: failed to save refresh token: %w", err)
	}
	return nil
}
