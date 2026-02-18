package service

import (
	"context"
	"ecommerce/services/auth/internal/domain"
	"ecommerce/services/auth/internal/repository"
	"ecommerce/services/auth/internal/utils"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, name, email, password, role, provider, providerId string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, error)
}

type authService struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) AuthService {
	return &authService{repo: repo}
}

func (a *authService) Register(ctx context.Context, name, email, password, role, provider, providerId string) (*domain.User, error) {
	res, err := a.repo.GetUserByEmail(ctx, email)

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

	err = a.repo.CreateUser(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("service: failed to create user: %w", err)
	}

	return &user, nil
}

func (a *authService) Login(ctx context.Context, email, password string) (*domain.User, error) {
	res, err := a.repo.GetUserByEmail(ctx, email)
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
