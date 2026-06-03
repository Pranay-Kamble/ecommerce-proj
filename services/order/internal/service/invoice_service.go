package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"ecommerce/services/order/internal/client"
	"ecommerce/services/order/internal/domain"
	"ecommerce/services/order/internal/repository"

	"github.com/jung-kurt/gofpdf"
	"github.com/sixafter/nanoid"
)

type InvoiceService interface {
	GenerateAndStore(ctx context.Context, order *domain.Order) error
	GetInvoiceURL(ctx context.Context, orderPublicID string, userID string) (string, error)
}

type invoiceService struct {
	invoiceRepo repository.InvoiceRepository
	orderRepo   repository.OrderRepository
	s3Client    client.S3Client
	nanoGen     nanoid.Interface
}

func NewInvoiceService(invoiceRepo repository.InvoiceRepository, orderRepo repository.OrderRepository, s3Client client.S3Client) (InvoiceService, error) {
	gen, err := nanoid.NewGenerator(nanoid.WithAlphabet("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	if err != nil {
		return nil, fmt.Errorf("service: failed to initialize nanoid generator: %w", err)
	}
	return &invoiceService{
		invoiceRepo: invoiceRepo,
		orderRepo:   orderRepo,
		s3Client:    s3Client,
		nanoGen:     gen,
	}, nil
}

func (s *invoiceService) GenerateAndStore(ctx context.Context, order *domain.Order) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(40, 10, "INVOICE")

	pdf.SetFont("Arial", "", 12)
	pdf.Ln(15)
	pdf.Cell(40, 10, fmt.Sprintf("Order Number: %s", order.PublicID))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Invoice Date: %s", time.Now().Format("2006-01-02")))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Billed To: %s", order.ShippingName))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Address: %s, %s", order.ShippingAddress, order.ShippingCity))

	pdf.Ln(15)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(80, 10, "Item", "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 10, "Qty", "1", 0, "C", false, 0, "")
	pdf.CellFormat(40, 10, "Price (INR)", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, "Total (INR)", "1", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "", 12)
	for _, item := range order.Items {
		itemTotal := float64(item.Quantity) * item.Price
		pdf.CellFormat(80, 10, item.ProductID, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", item.Price), "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", itemTotal), "1", 1, "R", false, 0, "")
	}

	pdf.Ln(10)
	gstAmount := order.TotalAmount * 0.18
	grandTotal := order.TotalAmount + gstAmount

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(110, 10, "", "", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, "Subtotal", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", order.TotalAmount), "1", 1, "R", false, 0, "")

	pdf.CellFormat(110, 10, "", "", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, "GST (18%)", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", gstAmount), "1", 1, "R", false, 0, "")

	pdf.CellFormat(110, 10, "", "", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, "Grand Total", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", grandTotal), "1", 1, "R", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return fmt.Errorf("service: failed to generate PDF: %w", err)
	}

	nanoID, _ := s.nanoGen.NewWithLength(8)
	objectKey := fmt.Sprintf("%s_invoice_%s.pdf", order.PublicID, nanoID)

	fileURL, err := s.s3Client.UploadPDF(ctx, &buf, objectKey)
	if err != nil {
		return fmt.Errorf("service: failed to upload invoice to S3: %w", err)
	}

	invID, _ := s.nanoGen.NewWithLength(12)
	invoice := &domain.Invoice{
		PublicID:  fmt.Sprintf("INV-%s", invID),
		OrderID:   order.ID,
		ObjectKey: objectKey,
		FileURL:   fileURL,
	}

	if err := s.invoiceRepo.CreateInvoice(ctx, invoice); err != nil {
		return fmt.Errorf("service: failed to save invoice to DB: %w", err)
	}

	return nil
}

func (s *invoiceService) GetInvoiceURL(ctx context.Context, orderPublicID string, userID string) (string, error) {
	order, err := s.orderRepo.GetOrderByPublicID(ctx, orderPublicID)
	if err != nil {
		return "", fmt.Errorf("service: failed to get order: %w", err)
	}
	if order == nil || order.UserID != userID {
		return "", fmt.Errorf("service: unauthorized or order not found")
	}

	invoice, err := s.invoiceRepo.GetInvoiceByOrderID(ctx, order.ID)
	if err != nil {
		return "", fmt.Errorf("service: failed to get invoice: %w", err)
	}
	if invoice == nil {
		return "", fmt.Errorf("service: invoice not generated yet")
	}

	url, err := s.s3Client.GeneratePresignedURL(ctx, invoice.ObjectKey, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("service: failed to generate presigned URL: %w", err)
	}

	return url, nil
}
