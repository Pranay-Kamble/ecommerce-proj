package handler

import (
	"ecommerce/pkg/logger"
	"net/http"
	"strings"

	"ecommerce/services/order/internal/domain"
	"ecommerce/services/order/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

type checkoutRequest struct {
	Name        string `json:"name" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
	AddressLine string `json:"address_line" binding:"required"`
	City        string `json:"city" binding:"required"`
	State       string `json:"state" binding:"required"`
	ZipCode     string `json:"zip_code" binding:"required"`
}

func (h *OrderHandler) Checkout(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req checkoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shipping details", "details": err.Error()})
		return
	}

	shippingAddress := domain.Address{
		AddressLine: req.AddressLine,
		City:        req.City,
		State:       req.State,
		ZipCode:     req.ZipCode,
	}

	order, paymentURL, err := h.orderService.Checkout(c.Request.Context(), userID, req.Name, req.Phone, shippingAddress)
	if err != nil {
		if strings.Contains(err.Error(), "empty cart") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your cart is empty. Please add items before checking out."})
			return
		}

		if strings.Contains(err.Error(), "unavailable") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "One or more items in your cart are currently unavailable.",
				"details": err.Error(),
			})
			return
		}
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.Unavailable:
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Product catalog is temporarily offline. Please try again in a few minutes."})
				return
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checkout payload.", "details": st.Message()})
				return
			case codes.DeadlineExceeded:
				c.JSON(http.StatusGatewayTimeout, gin.H{"error": " Checkout timed out while verifying prices. Please try again."})
				return
			}
		}

		logger.Error("Failed to checkout.", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "An unexpected error occurred during checkout.",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Order created successfully. Please complete your payment.",
		"order_id":    order.PublicID,
		"payment_url": paymentURL,
	})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	publicID := c.Param("public_id")
	if publicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order id is required"})
		return
	}

	order, err := h.orderService.GetOrder(c.Request.Context(), publicID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	orders, err := h.orderService.GetUserOrders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}

	if orders == nil {
		orders = []domain.Order{}
	}

	c.JSON(http.StatusOK, orders)
}
