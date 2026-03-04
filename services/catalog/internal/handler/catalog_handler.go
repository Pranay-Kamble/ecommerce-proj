package handler

import "ecommerce/services/catalog/internal/service"

type CatalogHandler struct {
	category service.CategoryService
	product  service.ProductService
	seller   service.SellerService
}
