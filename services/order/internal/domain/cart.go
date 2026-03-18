package domain

type CartItem struct {
	ProductVariantID string  `json:"product_variant_id"`
	Quantity         int     `json:"quantity"`
	Price            float64 `json:"price"`
}

type Cart struct {
	UserID string     `json:"user_id"`
	Items  []CartItem `json:"items"`
}
