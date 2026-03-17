package handler

import (
	"context"

	pb "ecommerce/pkg/protobufs/catalog"
	"ecommerce/services/catalog/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CatalogGrpcServer struct {
	pb.UnimplementedCatalogServiceServer
	productService service.ProductService
}

func NewCatalogGrpcServer(productService service.ProductService) *CatalogGrpcServer {
	return &CatalogGrpcServer{productService: productService}
}

func (s *CatalogGrpcServer) CheckPrices(ctx context.Context, req *pb.CheckPricesRequest) (*pb.CheckPricesResponse, error) {

	if req == nil || len(req.ProductIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "product_ids array cannot be empty")
	}

	if len(req.ProductIds) > 100 {
		return nil, status.Error(codes.InvalidArgument, "cannot process more than 100 variants per request")
	}

	variants, err := s.productService.VerifyVariants(ctx, req.ProductIds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database error while verifying variants: %v", err)
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
