package domain

// ProductDocument represents the Elasticsearch document for a product.
// Price is stored as min/max range since prices live on variants, not products.
// InStock is true if any variant has inventory > 0.
type ProductDocument struct {
	PublicID     string  `json:"public_id"`
	Title        string  `json:"title"`
	Brand        string  `json:"brand"`
	Description  string  `json:"description"`
	CategoryName string  `json:"category_name"`
	SellerName   string  `json:"seller_name"`
	MinPrice     float64 `json:"min_price"`
	MaxPrice     float64 `json:"max_price"`
	InStock      bool    `json:"in_stock"`
	Slug         string  `json:"slug"`
	ImageURL     string  `json:"image_url"`
}

// SearchFilters holds the parsed query parameters for a product search request.
type SearchFilters struct {
	Query    string  `json:"q"`
	Category string  `json:"category"`
	MinPrice float64 `json:"min_price"`
	MaxPrice float64 `json:"max_price"`
	InStock  *bool   `json:"in_stock"`
	Page     int     `json:"page"`
	Limit    int     `json:"limit"`
}

// SearchResult wraps search results with pagination metadata.
type SearchResult struct {
	Products []ProductDocument `json:"products"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
}
