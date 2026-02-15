package service

import (
	"context"
	"ecommerce/services/auth/internal/domain"
	"ecommerce/services/auth/internal/repository"
	"ecommerce/services/auth/internal/utils"
	"errors"
	"fmt"
)

type AuthService interface {
	Register(ctx context.Context, name, email, password, role, provider, providerId string) error
}

type authService struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) AuthService {
	return &authService{repo: repo}
}

func (a *authService) Register(ctx context.Context, name, email, password, role, provider, providerId string) error {
	res, err := a.repo.GetUserByEmail(ctx, email)

	if err != nil {
		return fmt.Errorf("service: failed to check email existence: %w", err)
	}

	if res != nil {
		return errors.New("service: email already exists")
	}

	var hashedPassword string
	hashedPassword, err = utils.HashPassword(password)

	if err != nil {
		return err
	}

	user := domain.User{Name: name, Email: email, Password: hashedPassword, Role: role, Provider: provider, ProviderID: providerId}

	err = a.repo.CreateUser(ctx, &user)
	if err != nil {
		return fmt.Errorf("service: failed to create user: %w", err)
	}

	return nil
}
