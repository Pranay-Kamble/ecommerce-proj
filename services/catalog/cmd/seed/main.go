package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"ecommerce/services/catalog/internal/domain"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	godotenv.Load(".env")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=admin password=password dbname=catalog_db port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Starting database ingestion")

	log.Println("🧹 Cleaning old data")
	db.Exec("TRUNCATE TABLE variants, products, categories, sellers CASCADE;")

	gofakeit.Seed(0)

	log.Println("Seeding Sellers")
	var sellers []domain.Seller
	for i := 0; i < 5; i++ {
		seller := domain.Seller{
			UserID:            gofakeit.LetterN(21),
			Name:              gofakeit.Company(),
			Description:       gofakeit.CompanySuffix(),
			LogoURL:           gofakeit.ImageURL(200, 200),
			SupportEmail:      gofakeit.Email(),
			SupportPhone:      gofakeit.Phone(),
			GSTIN:             strings.ToUpper(gofakeit.LetterN(15)),
			RegisteredAddress: gofakeit.Address().Address,
			Status:            "active",
			IsVerified:        true,
		}
		db.Create(&seller)
		sellers = append(sellers, seller)
	}

	log.Println("Seeding Categories")
	var leafCategories []domain.Category
	parentNames := []string{"Electronics", "Clothing", "Home"}

	for _, pName := range parentNames {
		cleanParentPath := strings.ReplaceAll(strings.ToLower(pName), " ", "_")
		parent := domain.Category{
			Name: pName,
			Path: cleanParentPath,
		}
		db.Create(&parent)

		for i := 0; i < 3; i++ {
			cName := gofakeit.Word()
			cleanChildPath := cleanParentPath + "." + strings.ReplaceAll(strings.ToLower(cName), " ", "_")

			child := domain.Category{
				Name:     strings.Title(cName),
				Path:     cleanChildPath,
				ParentID: &parent.ID,
			}
			db.Create(&child)
			leafCategories = append(leafCategories, child)
		}
	}

	log.Println("Seeding Products and Variants")
	for i := 0; i < 20; i++ {
		seller := sellers[gofakeit.Number(0, len(sellers)-1)]
		category := leafCategories[gofakeit.Number(0, len(leafCategories)-1)]

		product := domain.Product{
			CategoryID:  category.ID,
			SellerID:    seller.ID,
			Title:       gofakeit.ProductName(),
			Brand:       gofakeit.Company(),
			Description: gofakeit.ProductDescription(),
			Highlights:  []string{gofakeit.Sentence(3), gofakeit.Sentence(4)},
			Dimensions: map[string]interface{}{
				"weight": gofakeit.Float32Range(0.5, 10.0),
				"length": gofakeit.Float32Range(10, 100),
			},
			Images: []*domain.Image{
				{URL: gofakeit.ImageURL(800, 800), AltText: "Main View", IsPrimary: true},
				{URL: gofakeit.ImageURL(800, 800), AltText: "Side View", IsPrimary: false},
			},
		}
		db.Create(&product)

		numVariants := gofakeit.Number(1, 3)
		for v := 0; v < numVariants; v++ {
			variant := domain.Variant{
				ProductID: product.ID,
				Title:     fmt.Sprintf("%s - %s", product.Title, gofakeit.Color()),
				SKU:       fmt.Sprintf("SKU-%s-%d", strings.ToUpper(gofakeit.LetterN(5)), v),
				Price:     gofakeit.Price(10, 1000),
				Inventory: gofakeit.Number(0, 100),
				Specifications: map[string]interface{}{
					"Color": gofakeit.Color(),
					"Size":  []string{"S", "M", "L", "XL"}[gofakeit.Number(0, 3)],
				},
				Images: []*domain.Image{
					{URL: gofakeit.ImageURL(400, 400), AltText: "Variant Image", IsPrimary: true},
				},
			}
			db.Create(&variant)
		}
	}

	log.Println("Database successfully seeded.")
}
