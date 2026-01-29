package grpc

import (
	account "account/internal"
	"account/pb"
	"context"

	"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

type GrpcAccountServer struct {
	pb.UnimplementedAccountServiceServer
	repo account.AccountRepository
}

func NewGrpcAccountServer(repo account.AccountRepository) *GrpcAccountServer {
	return &GrpcAccountServer{
		repo: repo,
	}
}

func (s *GrpcAccountServer) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	account, err := s.repo.GetByID(ctx, req.AccountId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	return &pb.GetAccountResponse{
		Id:       account.ID,
		Name:     account.Name,
		IsActive: true,
		IsExists: true,
	}, nil
}
