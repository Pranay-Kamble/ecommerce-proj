package service

import (
	"context"
	"ecommerce/services/auth/internal/domain"
	"ecommerce/services/auth/internal/repository"
	"ecommerce/services/auth/internal/utils"
	"errors"
	"fmt"
	"time"

	"github.com/sixafter/nanoid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, name, email, password, role, provider, providerId string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, error)
	RotateRefreshToken(ctx context.Context, refreshToken string) (string, *domain.User, error)
	//Logout(ctx context.Context) error
	SaveRefreshToken(ctx context.Context, userID, hashedRefreshToken, familyId string) (*domain.Token, error)
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

func (a *authService) SaveRefreshToken(ctx context.Context, userID, hashedRefreshToken, familyID string) (*domain.Token, error) {
	refreshToken := domain.NewToken(userID, hashedRefreshToken, familyID, time.Now().Add(time.Hour*24*7))

	err := a.tokenRepo.Create(ctx, refreshToken)

	if err != nil {
		return nil, fmt.Errorf("service: failed to save refresh token: %w", err)
	}
	return refreshToken, nil
}

func (a *authService) RotateRefreshToken(ctx context.Context, refreshToken string) (string, *domain.User, error) {
	hashToken := utils.HashUsingSHA256(nanoid.ID(refreshToken))

	fullToken, err := a.tokenRepo.FindByTokenHash(ctx, hashToken)

	if err != nil {
		return "", nil, fmt.Errorf("service: failed to rotate refresh token: %w", err)
	}

	if fullToken == nil {
		return "", nil, errors.New("service: refresh token not found")
	}

	if fullToken.IsUsed || fullToken.IsRevoked {
		err = a.tokenRepo.RevokeTokenFamily(ctx, fullToken.FamilyID)
		if err != nil {
			return "", nil, fmt.Errorf("service: failed to revoke refresh token family: %w", err)
		}
		return "", nil, errors.New("service: refresh token already used or revoked")
	}

	if time.Now().After(fullToken.ExpiresOn) {
		return "", nil, errors.New("service: refresh token is expired")
	}

	err = a.tokenRepo.MarkAsUsed(ctx, fullToken.TokenHash)
	if err != nil {
		return "", nil, fmt.Errorf("service: failed to mark refresh token as used: %w", err)
	}

	newTokenString, newTokenHash, err := utils.GetRefreshTokenStringWithFamilyID(fullToken.FamilyID)
	if err != nil {
		return "", nil, fmt.Errorf("service: failed to rotate refresh token: %w", err)
	}

	_, err = a.SaveRefreshToken(ctx, fullToken.UserID, newTokenHash, fullToken.FamilyID)
	if err != nil {
		return "", nil, fmt.Errorf("service: failed to rotate refresh token: %w", err)
	}

	tokenUser, err := a.userRepo.GetUserByID(ctx, fullToken.UserID)
	if err != nil {
		return "", nil, fmt.Errorf("service: failed to get user by ID: %w", err)
	}

	if tokenUser == nil {
		return "", nil, errors.New("service: user not found")
	}

	return newTokenString, tokenUser, nil
}
