package repository

import (
	"context"
	"ecommerce/services/catalog/internal/domain"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	GetByPublicID(ctx context.Context, publicID string) (*domain.Category, error)
	GetByName(ctx context.Context, name string) (*domain.Category, error)
	GetDescendants(ctx context.Context, id string) ([]*domain.Category, error)
	GetAncestors(ctx context.Context, id string) ([]*domain.Category, error)
	GetAllCategories(ctx context.Context, parentID *uuid.UUID) ([]*domain.Category, error)
}

type categoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) CategoryRepository {
	return &categoryRepo{db: db}
}

func (c *categoryRepo) Create(ctx context.Context, category *domain.Category) error {
	err := gorm.G[*domain.Category](c.db).Create(ctx, &category)
	if err != nil {
		return fmt.Errorf("repository: could not create category %w", err)
	}

	return nil
}

func (c *categoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	category, err := gorm.G[*domain.Category](c.db).Where("id = ?", id).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return category, nil
}

func (c *categoryRepo) GetByPublicID(ctx context.Context, publicID string) (*domain.Category, error) {
	category, err := gorm.G[*domain.Category](c.db).Where("public_id = ?", publicID).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("repository: could not get category by public ID %w", err)
	}

	return category, nil
}

func (c *categoryRepo) GetByName(ctx context.Context, name string) (*domain.Category, error) {
	category, err := gorm.G[*domain.Category](c.db).Where("name = ?", name).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("repository: could not get category by name %w", err)
	}

	return category, nil
}

func (c *categoryRepo) GetDescendants(ctx context.Context, publicId string) ([]*domain.Category, error) {
	categories, err := gorm.G[*domain.Category](c.db).Where(" path::ltree <@ (SELECT path FROM categories WHERE public_id = ?)", publicId).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository: could not get category descendants : %w", err)
	}

	return categories, nil
}

func (c *categoryRepo) GetAncestors(ctx context.Context, publicId string) ([]*domain.Category, error) {
	categories, err := gorm.G[*domain.Category](c.db).Where(" path::ltree @> (SELECT path FROM categories WHERE public_id = ?)", publicId).Order("nlevel(path) ASC").Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository: could not get category ancestors : %w", err)
	}

	return categories, nil
}

func (c *categoryRepo) GetAllCategories(ctx context.Context, parentID *uuid.UUID) ([]*domain.Category, error) {
	var categories []*domain.Category
	var err error
	if parentID == nil {
		categories, err = gorm.G[*domain.Category](c.db).Where("parent_id IS NULL").Find(ctx)
	} else {
		categories, err = gorm.G[*domain.Category](c.db).Where("parent_id = ?", *parentID).Find(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("repository: could not get category all : %w", err)
	}
	return categories, nil
}
