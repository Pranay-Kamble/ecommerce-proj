package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ecommerce/pkg/logger"
	"ecommerce/services/search/internal/domain"

	"github.com/elastic/go-elasticsearch/v8"
	"go.uber.org/zap"
)

const indexName = "products"

// ElasticRepository wraps the Elasticsearch client for product indexing and search.
type ElasticRepository interface {
	EnsureIndex(ctx context.Context) error
	IndexProduct(ctx context.Context, doc domain.ProductDocument) error
	DeleteProduct(ctx context.Context, publicID string) error
	SearchProducts(ctx context.Context, filters domain.SearchFilters) (*domain.SearchResult, error)
}

type elasticRepository struct {
	client *elasticsearch.Client
}

// NewElasticRepository creates a new ElasticRepository.
func NewElasticRepository(client *elasticsearch.Client) ElasticRepository {
	return &elasticRepository{client: client}
}

// EnsureIndex creates the "products" index with proper mappings if it doesn't already exist.
func (r *elasticRepository) EnsureIndex(ctx context.Context) error {
	res, err := r.client.Indices.Exists([]string{indexName})
	if err != nil {
		return fmt.Errorf("repository: failed to check if index exists: %w", err)
	}
	defer res.Body.Close()

	// Index already exists
	if res.StatusCode == 200 {
		logger.Info("Elasticsearch index already exists", zap.String("index", indexName))
		return nil
	}

	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"analysis": {
				"analyzer": {
					"product_analyzer": {
						"type": "custom",
						"tokenizer": "standard",
						"filter": ["lowercase", "asciifolding"]
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"public_id":     { "type": "keyword" },
				"title":         { "type": "text", "analyzer": "product_analyzer", "fields": { "keyword": { "type": "keyword" } } },
				"brand":         { "type": "text", "analyzer": "product_analyzer", "fields": { "keyword": { "type": "keyword" } } },
				"description":   { "type": "text", "analyzer": "product_analyzer" },
				"category_name": { "type": "text", "analyzer": "product_analyzer", "fields": { "keyword": { "type": "keyword" } } },
				"seller_name":   { "type": "text", "analyzer": "product_analyzer", "fields": { "keyword": { "type": "keyword" } } },
				"min_price":     { "type": "float" },
				"max_price":     { "type": "float" },
				"in_stock":      { "type": "boolean" },
				"slug":          { "type": "keyword" },
				"image_url":     { "type": "keyword", "index": false }
			}
		}
	}`

	res, err = r.client.Indices.Create(
		indexName,
		r.client.Indices.Create.WithBody(strings.NewReader(mapping)),
		r.client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("repository: failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("repository: error response when creating index: %s", res.String())
	}

	logger.Info("Created Elasticsearch index", zap.String("index", indexName))
	return nil
}

// IndexProduct indexes or updates a product document in Elasticsearch.
// Uses public_id as the document ID for upsert behavior.
func (r *elasticRepository) IndexProduct(ctx context.Context, doc domain.ProductDocument) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("repository: failed to marshal product document: %w", err)
	}

	res, err := r.client.Index(
		indexName,
		bytes.NewReader(body),
		r.client.Index.WithDocumentID(doc.PublicID),
		r.client.Index.WithContext(ctx),
		r.client.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("repository: failed to index product: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("repository: error response when indexing product: %s", res.String())
	}

	return nil
}

// DeleteProduct removes a product document from Elasticsearch by its public_id.
func (r *elasticRepository) DeleteProduct(ctx context.Context, publicID string) error {
	res, err := r.client.Delete(
		indexName,
		publicID,
		r.client.Delete.WithContext(ctx),
		r.client.Delete.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("repository: failed to delete product: %w", err)
	}
	defer res.Body.Close()

	// 404 is acceptable — document may not exist
	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("repository: error response when deleting product: %s", res.String())
	}

	return nil
}

// SearchProducts performs a full-text search with optional filters.
func (r *elasticRepository) SearchProducts(ctx context.Context, filters domain.SearchFilters) (*domain.SearchResult, error) {
	query := r.buildQuery(filters)

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to marshal search query: %w", err)
	}

	from := (filters.Page - 1) * filters.Limit

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(indexName),
		r.client.Search.WithBody(bytes.NewReader(body)),
		r.client.Search.WithFrom(from),
		r.client.Search.WithSize(filters.Limit),
		r.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("repository: error response from search: %s", res.String())
	}

	var searchResponse struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.ProductDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err = json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("repository: failed to decode search response: %w", err)
	}

	products := make([]domain.ProductDocument, 0, len(searchResponse.Hits.Hits))
	for _, hit := range searchResponse.Hits.Hits {
		products = append(products, hit.Source)
	}

	return &domain.SearchResult{
		Products: products,
		Total:    searchResponse.Hits.Total.Value,
		Page:     filters.Page,
		Limit:    filters.Limit,
	}, nil
}

// buildQuery constructs the Elasticsearch query DSL from filters.
func (r *elasticRepository) buildQuery(filters domain.SearchFilters) map[string]interface{} {
	must := make([]map[string]interface{}, 0)
	filterClauses := make([]map[string]interface{}, 0)

	// Full-text search across title, brand, description
	if filters.Query != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  filters.Query,
				"fields": []string{"title^3", "brand^2", "description"},
				"type":   "best_fields",
				"fuzziness": "AUTO",
			},
		})
	}

	// Category filter (case-insensitive match via analyzed field)
	if filters.Category != "" {
		filterClauses = append(filterClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"category_name": filters.Category,
			},
		})
	}

	// Price range filters:
	// min_price param → at least one variant >= this price → max_price >= min_price_param
	if filters.MinPrice > 0 {
		filterClauses = append(filterClauses, map[string]interface{}{
			"range": map[string]interface{}{
				"max_price": map[string]interface{}{
					"gte": filters.MinPrice,
				},
			},
		})
	}

	// max_price param → at least one variant <= this price → min_price <= max_price_param
	if filters.MaxPrice > 0 {
		filterClauses = append(filterClauses, map[string]interface{}{
			"range": map[string]interface{}{
				"min_price": map[string]interface{}{
					"lte": filters.MaxPrice,
				},
			},
		})
	}

	// In-stock filter
	if filters.InStock != nil {
		filterClauses = append(filterClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"in_stock": *filters.InStock,
			},
		})
	}

	// If no query or filters, match all
	if len(must) == 0 && len(filterClauses) == 0 {
		return map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
	}

	boolQuery := map[string]interface{}{}
	if len(must) > 0 {
		boolQuery["must"] = must
	}
	if len(filterClauses) > 0 {
		boolQuery["filter"] = filterClauses
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": boolQuery,
		},
	}
}
