package service

import (
	"context"
	"ecommerce/services/catalog/internal/domain"
	"ecommerce/services/catalog/internal/repository"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sixafter/nanoid"
)

type CategoryService interface {
	CreateCategory(c context.Context, name string, parentPublicID string) error
	GetAllCategories(c context.Context, parentPublicID string) ([]*domain.Category, error)
	GetCategoryBreadCrumbs(c context.Context, publicID string) ([]*domain.Category, error)
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func (p *categoryService) GetCategoryBreadCrumbs(c context.Context, parentPublicID string) ([]*domain.Category, error) {
	category, err := p.categoryRepo.GetByPublicID(c, parentPublicID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get category by public id: %w", err)
	} else if category == nil {
		return nil, fmt.Errorf("service: category not found")
	}

	upperCategories, err := p.categoryRepo.GetAncestors(c, parentPublicID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get ancestors: %w", err)
	}

	return upperCategories, nil
}

func (p *categoryService) GetAllCategories(c context.Context, publicID string) ([]*domain.Category, error) {
	var categories []*domain.Category
	var err error
	if publicID == "" {
		categories, err = p.categoryRepo.GetAllCategories(c, nil)
	} else {
		category, innerErr := p.categoryRepo.GetByPublicID(c, publicID)
		if innerErr != nil {
			return nil, fmt.Errorf("service: failed to get category by public id: %w", innerErr)
		}
		if category == nil {
			return nil, fmt.Errorf("service: category not found")
		}
		categories, err = p.categoryRepo.GetAllCategories(c, &category.ID)
	}

	if err != nil {
		return nil, fmt.Errorf("service: failed to get all categories: %w", err)
	}
	return categories, nil
}

func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

func (p *categoryService) CreateCategory(c context.Context, name string, parentPublicID string) error {
	var finalPath string
	var parentID *uuid.UUID
	publicCategoryID, err := nanoid.New()

	if err != nil {
		return fmt.Errorf("repository: could not generate category id: %w", err)
	}

	safePublicID := "cat_" + strings.ReplaceAll(publicCategoryID.String(), "-", "_")

	if parentPublicID != "" {
		parentCategory, err := p.categoryRepo.GetByPublicID(c, parentPublicID)
		if err != nil {
			return fmt.Errorf("service: failed to get parent category by public id: %w", err)
		} else if parentCategory == nil {
			return fmt.Errorf("service: invalid parent category")
		}

		finalPath = parentCategory.Path + "." + safePublicID
		parentID = &parentCategory.ID
	} else {
		finalPath = safePublicID
	}

	category := &domain.Category{
		Path:     finalPath,
		Name:     name,
		ParentID: parentID,
		PublicID: safePublicID,
	}

	err = p.categoryRepo.Create(c, category)
	if err != nil {
		return fmt.Errorf("service: failed to create category: %w", err)
	}

	return nil
}
