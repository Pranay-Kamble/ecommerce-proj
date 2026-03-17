package handler

import (
	"context"

	pb "ecommerce/pkg/protobufs/catalog"
	"ecommerce/services/catalog/internal/service"
)

type CatalogGrpcServer struct {
	pb.UnimplementedCatalogServiceServer
	productService service.ProductService
}

func NewCatalogGrpcServer(productService service.ProductService) *CatalogGrpcServer {
	return &CatalogGrpcServer{productService: productService}
}

func (s *CatalogGrpcServer) CheckPrices(ctx context.Context, req *pb.CheckPricesRequest) (*pb.CheckPricesResponse, error) {
	variants, err := s.productService.VerifyVariants(ctx, req.ProductIds)
	if err != nil {
		return nil, err
	}

	var verifiedProducts []*pb.ProductCheck
	for _, v := range variants {
		verifiedProducts = append(verifiedProducts, &pb.ProductCheck{
			ProductId:   v.PublicID,
			Price:       v.Price,
			IsAvailable: v.Inventory > 0,
		})
	}

	return &pb.CheckPricesResponse{
		Products: verifiedProducts,
	}, nil
}
