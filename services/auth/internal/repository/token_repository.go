package repository

import (
	"context"
	"ecommerce/services/auth/internal/domain"
	"fmt"

	"gorm.io/gorm"
)

type TokenRepository interface {
	Create(ctx context.Context, token *domain.Token) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*domain.Token, error)
	MarkAsUsed(ctx context.Context, tokenHash string) error
	RevokeTokenFamily(ctx context.Context, familyID string) error
	RevokeUser(ctx context.Context, userID string) error
}

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (t *tokenRepository) Create(ctx context.Context, token *domain.Token) error {
	err := gorm.G[domain.Token](t.db).Create(ctx, token)

	if err != nil {
		return fmt.Errorf("repository: could not create token: %w", err)
	}

	return nil
}

func (t *tokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.Token, error) {
	res, err := gorm.G[*domain.Token](t.db).Where("token_hash = ?", tokenHash).First(ctx)

	if err != nil {
		return nil, fmt.Errorf("repository: could not find token: %w", err)
	}

	return res, nil
}

func (t *tokenRepository) MarkAsUsed(ctx context.Context, tokenHash string) error {
	_, err := gorm.G[domain.Token](t.db).Where("token_hash = ?", tokenHash).Update(ctx, "is_used", true)

	if err != nil {
		return fmt.Errorf("repository: could not mark as used: %w", err)
	}

	return nil
}

func (t *tokenRepository) RevokeTokenFamily(ctx context.Context, familyID string) error {
	_, err := gorm.G[domain.Token](t.db).Where("family_id = ?", familyID).Update(ctx, "is_revoked", true)
	if err != nil {
		return fmt.Errorf("repository: could not revoke token family: %w", err)
	}
	return nil
}

func (t *tokenRepository) RevokeUser(ctx context.Context, userID string) error {
	_, err := gorm.G[domain.Token](t.db).Where("user_id = ?", userID).Update(ctx, "is_revoked", true)
	if err != nil {
		return fmt.Errorf("repository: could not revoke user: %w", err)
	}
	return nil
}
