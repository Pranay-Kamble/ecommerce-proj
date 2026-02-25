package service

import (
	"context"
	"ecommerce/services/auth/internal/domain"
	"ecommerce/services/auth/internal/repository"
	"ecommerce/services/auth/internal/utils"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sixafter/nanoid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, name, email, password, role, provider, providerId string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, error)
	RotateRefreshToken(ctx context.Context, refreshToken string) (string, *domain.User, error)
	Logout(ctx context.Context, refreshToken string) error
	SaveRefreshToken(ctx context.Context, userID, hashedRefreshToken, familyId string) (*domain.Token, error)
	VerifyEmail(ctx context.Context, email string, otp string) (*domain.User, error)
	CreateOTP(ctx context.Context, email string, ttl time.Duration) (string, error)
	ResendOTP(ctx context.Context, email string) (string, error)
	OAuthLogin(ctx context.Context, email string, providerID string, name string) (*domain.User, error)
}

type authService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	otpRepo   repository.OTPRepository
}

func (a *authService) ResendOTP(ctx context.Context, email string) (string, error) {
	user, err := a.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("service: could not get user by email : %w", err)
	}

	if user == nil {
		return "", errors.New("service: user not found")
	}

	if user.IsVerified {
		return "", errors.New("service: user is already verified")
	}

	otp, err := a.CreateOTP(ctx, email, time.Minute*10)
	if err != nil {
		return "", fmt.Errorf("service: failed to create new OTP: %w", err)
	}

	return otp, nil
}

func (a *authService) CreateOTP(ctx context.Context, email string, ttl time.Duration) (string, error) {
	otp, err := utils.GetOTP()

	if err != nil {
		return "", err
	}

	err = a.otpRepo.Create(ctx, otp, email, ttl)
	if err != nil {
		return "", fmt.Errorf("service: failed to write OTP into redis: %w", err)
	}

	return otp, nil
}

func (a *authService) VerifyEmail(ctx context.Context, email string, otp string) (*domain.User, error) {
	dbOTP, err := a.otpRepo.Get(ctx, email)

	if err != nil {
		if strings.Contains(err.Error(), "repository: otp not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("service: failed to get OTP from OTP repository %w", err)
	}

	if dbOTP != otp {
		return nil, nil
	}

	user, err := a.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get user by email %w", err)
	}

	err = a.userRepo.UpdateVerified(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to update user by email %w", err)
	}

	err = a.otpRepo.Delete(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("service: failed to delete OTP from OTP repository %w", err)
	}

	return user, nil
}

func (a *authService) Logout(ctx context.Context, refreshToken string) error {
	hashToken := utils.HashUsingSHA256(nanoid.ID(refreshToken))
	fullToken, err := a.tokenRepo.FindByTokenHash(ctx, hashToken)
	if err != nil {
		return fmt.Errorf("service: could not find token %w", err)
	}

	if fullToken == nil {
		return nil
	}

	err = a.tokenRepo.RevokeTokenFamily(ctx, fullToken.FamilyID)
	if err != nil {
		return fmt.Errorf("service: could not revoke token family %w", err)
	}
	return nil
}

func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, otpRepo repository.OTPRepository) AuthService {
	return &authService{userRepo: userRepo, tokenRepo: tokenRepo, otpRepo: otpRepo}
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

	if !res.IsVerified {
		return nil, errors.New("service: user is not verified")
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

func (a *authService) OAuthLogin(ctx context.Context, email string, providerID string, name string) (*domain.User, error) {
	res, err := a.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("service: failed to check email existence: %w", err)
	}

	if res == nil {
		res, err = a.Register(ctx, name, email, "", "buyer", "google", providerID)
		if err != nil {
			return nil, fmt.Errorf("service: failed to register OAuth user: %w", err)
		}
	}

	if !res.IsVerified {
		err = a.userRepo.UpdateVerified(ctx, res.ID)
		if err != nil {
			return nil, fmt.Errorf("service: failed to update verified user: %w", err)
		}
	}

	return res, nil
}
