package service

import (
	"context"
	"ecommerce/services/order/internal/client"
	"fmt"
	"math"

	"ecommerce/services/order/internal/domain"
	"ecommerce/services/order/internal/repository"

	pb "ecommerce/pkg/protobufs/catalog"

	"github.com/sixafter/nanoid"
)

type OrderService interface {
	Checkout(ctx context.Context, userID string, name, phone string, address domain.Address) (*domain.Order, string, error)
	GetOrder(ctx context.Context, publicID string, userID string) (*domain.Order, error)
	GetUserOrders(ctx context.Context, userID string) ([]domain.Order, error)
	UpdateOrderStatus(ctx context.Context, id string, status string) error
}

type orderService struct {
	orderRepo repository.OrderRepository
	cartRepo  repository.CartRepository

	nanoGen       nanoid.Interface
	catalogClient pb.CatalogServiceClient
	paymentClient client.PaymentService
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	catalogClient pb.CatalogServiceClient,
	paymentClient client.PaymentService,
) (OrderService, error) {

	gen, err := nanoid.NewGenerator(nanoid.WithAlphabet("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	if err != nil {
		return nil, fmt.Errorf("service: failed to initialize nanoid generator: %w", err)
	}

	return &orderService{
		orderRepo:     orderRepo,
		cartRepo:      cartRepo,
		nanoGen:       gen,
		catalogClient: catalogClient,
		paymentClient: paymentClient,
	}, nil
}

func (s *orderService) Checkout(ctx context.Context, userID string, name, phone string, address domain.Address) (*domain.Order, string, error) {
	cart, err := s.cartRepo.GetCart(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("service: failed to get cart for checkout: %w", err)
	}

	if len(cart.Items) == 0 {
		return nil, "", fmt.Errorf("service: cannot checkout with an empty cart")
	}

	var productIDs []string
	for _, item := range cart.Items {
		productIDs = append(productIDs, item.ProductVariantID)
	}

	//gRPC call to get real product info if any of them are updated
	catalogResp, err := s.catalogClient.CheckPrices(ctx, &pb.CheckPricesRequest{
		ProductIds: productIDs,
	})
	if err != nil {
		return nil, "", fmt.Errorf("service: failed to communicate with catalog: %w", err)
	}

	verifiedProducts := make(map[string]*pb.ProductCheck)
	for _, p := range catalogResp.Products {
		verifiedProducts[p.ProductId] = p
	}

	var totalAmount float64 = 0
	var orderItems []domain.OrderItem

	for _, item := range cart.Items {
		vp, exists := verifiedProducts[item.ProductVariantID]

		if !exists || !vp.IsAvailable {
			return nil, "", fmt.Errorf("service: product %s is currently unavailable", item.ProductVariantID)
		}
		totalAmount += vp.Price * float64(item.Quantity)

		orderItems = append(orderItems, domain.OrderItem{
			ProductID: item.ProductVariantID,
			Quantity:  item.Quantity,
			Price:     vp.Price,
		})
	}

	id, err := s.nanoGen.NewWithLength(8)
	if err != nil {
		return nil, "", fmt.Errorf("service: failed to generate order number: %w", err)
	}
	orderNumber := fmt.Sprintf("ORD-%s", id)

	order := &domain.Order{
		PublicID:        orderNumber,
		UserID:          userID,
		TotalAmount:     totalAmount,
		Status:          "pending",
		ShippingName:    name,
		ShippingPhone:   phone,
		ShippingAddress: address.AddressLine,
		ShippingCity:    address.City,
		ShippingState:   address.State,
		ShippingZip:     address.ZipCode,
		Items:           orderItems,
	}

	if err = s.orderRepo.CreateOrder(ctx, order); err != nil {
		return nil, "", fmt.Errorf("service: failed to save order: %w", err)
	}

	amountInPaise := int64(math.Round(totalAmount * 100))

	paymentURL, err := s.paymentClient.InitiatePayment(ctx, orderNumber, userID, amountInPaise, "inr")
	if err != nil {
		return nil, "", fmt.Errorf("service: failed to initiate payment gateway: %w", err)
	}

	return order, paymentURL, nil
}

func (s *orderService) GetOrder(ctx context.Context, publicID string, userID string) (*domain.Order, error) {
	order, err := s.orderRepo.GetOrderByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to fetch order: %w", err)
	}
	if order == nil || order.UserID != userID {
		return nil, fmt.Errorf("service: order not found")
	}
	return order, nil
}

func (s *orderService) GetUserOrders(ctx context.Context, userID string) ([]domain.Order, error) {
	orders, err := s.orderRepo.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to fetch user orders: %w", err)
	}
	return orders, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, orderId string, status string) error {
	order, err := s.orderRepo.GetOrderByPublicID(ctx, orderId)
	if err != nil {
		return fmt.Errorf("service: failed to fetch order: %w", err)
	}
	order.Status = status
	err = s.orderRepo.UpdateOrder(ctx, order)
	if err != nil {
		return fmt.Errorf("service: failed to update order: %w", err)
	}
	return nil
}
