package repository

import (
	"context"
	"fmt"

	"ecommerce/services/order/internal/domain"

	"gorm.io/gorm"
)

type InvoiceRepository interface {
	CreateInvoice(ctx context.Context, invoice *domain.Invoice) error
	GetInvoiceByOrderID(ctx context.Context, orderID string) (*domain.Invoice, error)
}

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) CreateInvoice(ctx context.Context, invoice *domain.Invoice) error {
	if err := gorm.G[domain.Invoice](r.db).Create(ctx, invoice); err != nil {
		return fmt.Errorf("repository: failed to create invoice: %w", err)
	}
	return nil
}

func (r *invoiceRepository) GetInvoiceByOrderID(ctx context.Context, orderID string) (*domain.Invoice, error) {
	invoice, err := gorm.G[*domain.Invoice](r.db).Where("order_id = ?", orderID).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: failed to find invoice by order_id: %w", err)
	}
	return invoice, nil
}
