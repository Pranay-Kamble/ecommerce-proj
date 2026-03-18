package service

import (
	"context"
	"fmt"

	"ecommerce/pkg/logger"
	"ecommerce/services/order/internal/domain"
	"ecommerce/services/order/internal/repository"

	pb "ecommerce/pkg/protobufs/catalog"

	"github.com/sixafter/nanoid"
)

type OrderService interface {
	Checkout(ctx context.Context, userID string, name, phone string, address domain.Address) (*domain.Order, error)
	GetOrder(ctx context.Context, publicID string, userID string) (*domain.Order, error)
	GetUserOrders(ctx context.Context, userID string) ([]domain.Order, error)
}

type orderService struct {
	orderRepo repository.OrderRepository
	cartRepo  repository.CartRepository

	nanoGen       nanoid.Interface
	catalogClient pb.CatalogServiceClient
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	catalogClient pb.CatalogServiceClient,
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
	}, nil
}

func (s *orderService) Checkout(ctx context.Context, userID string, name, phone string, address domain.Address) (*domain.Order, error) {
	cart, err := s.cartRepo.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get cart for checkout: %w", err)
	}

	if len(cart.Items) == 0 {
		return nil, fmt.Errorf("service: cannot checkout with an empty cart")
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
		return nil, fmt.Errorf("service: failed to communicate with catalog: %w", err)
	}

	verifiedProducts := make(map[string]*pb.ProductCheck)
	for _, p := range catalogResp.Products {
		verifiedProducts[p.ProductId] = p
	}

	var totalAmount float64
	var orderItems []domain.OrderItem

	for _, item := range cart.Items {
		vp, exists := verifiedProducts[item.ProductVariantID]

		if !exists || !vp.IsAvailable {
			return nil, fmt.Errorf("service: product %s is currently unavailable", item.ProductVariantID)
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
		return nil, fmt.Errorf("service: failed to generate order number: %w", err)
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
		return nil, fmt.Errorf("service: failed to save order: %w", err)
	}

	if err = s.cartRepo.ClearCart(ctx, userID); err != nil {
		logger.Error(fmt.Sprintf("warning: failed to clear cart after successful checkout for user %s", userID))
	}

	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, publicID string, userID string) (*domain.Order, error) {
	order, err := s.orderRepo.GetOrderByPublicID(ctx, publicID, userID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to fetch order: %w", err)
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
