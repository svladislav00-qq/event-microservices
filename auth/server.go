package auth

import (
	"context"
	"errors"

	"github.com/svladislav00-qq/event-microservices/auth/pb"
	"github.com/svladislav00-qq/event-microservices/pkg/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthService interface {
	Register(ctx context.Context, email, password, username string) (*Account, error)
	Login(ctx context.Context, email, password string) (string, error)
	PromoteToModerator(ctx context.Context, userID string, departmentId string) (*Account, error)
	GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
	GetAccountById(ctx context.Context, userID string) (*Account, error)
	GetUsersByIDs(ctx context.Context, ids []string) ([]models.Account, error)
}

type serverAuth struct {
	pb.UnimplementedAuthServiceServer
	auth AuthService
}

func NewAuthServer(gRPC *grpc.Server, authService AuthService) {
	pb.RegisterAuthServiceServer(gRPC, &serverAuth{auth: authService})
}

func (s *serverAuth) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	res, err := s.auth.Register(ctx, req.Email, req.Password, req.Username)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.RegisterResponse{Account: &pb.Account{
		Id:        res.ID,
		Email:     res.Email,
		Username:  res.Username,
		Role:      roleToProto(res.Role),
		CreatedAt: timestamppb.New(res.CreatedAt),
		UpdatedAt: timestamppb.New(res.UpdatedAt),
	}}, nil
}

func (s *serverAuth) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}
		return nil, err
	}

	return &pb.LoginResponse{Token: token}, nil
}

func (s *serverAuth) PromoteToModerator(ctx context.Context, req *pb.PromoteToModeratorRequest) (*pb.PromoteToModeratorResponse, error) {
	if err := validatePromoteToModerator(req); err != nil {
		return nil, err
	}

	account, err := s.auth.PromoteToModerator(ctx, req.UserId, req.Department)
	if err != nil {
		return nil, err
	}

	return &pb.PromoteToModeratorResponse{Account: &pb.Account{
		Id:         account.ID,
		Email:      account.Email,
		Username:   account.Username,
		Role:       roleToProto(account.Role),
		Department: account.Department,
		CreatedAt:  timestamppb.New(account.CreatedAt),
	}}, nil
}

func (s *serverAuth) GetMe(ctx context.Context, req *pb.GetMeRequest) (*pb.GetMeResponse, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	account, err := s.auth.GetAccountById(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.GetMeResponse{
		Account: &pb.Account{
			Id:        account.ID,
			Email:     account.Email,
			Username:  account.Username,
			Role:      roleToProto(account.Role),
			CreatedAt: timestamppb.New(account.CreatedAt),
		},
	}, nil
}

func (s *serverAuth) GetUsersByIDs(ctx context.Context, req *pb.GetUsersByIDsRequest) (*pb.GetUsersByIDsResponse, error) {
	accounts, err := s.auth.GetUsersByIDs(ctx, req.UserIds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get users: %v", err)
	}

	var users []*pb.User
	for _, acc := range accounts {
		users = append(users, &pb.User{
			Id:   acc.ID,
			Name: acc.Username,
		})
	}

	return &pb.GetUsersByIDsResponse{
		Users: users,
	}, nil
}

func validateRegister(req *pb.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetUsername() == "" {
		return status.Error(codes.InvalidArgument, "username is required")
	}

	return nil
}

func validateLogin(req *pb.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	return nil
}

func validatePromoteToModerator(req *pb.PromoteToModeratorRequest) error {
	if req.GetUserId() == "" {
		return status.Error(codes.InvalidArgument, "userID is required")
	}
	if req.GetDepartment() == "" {
		return status.Error(codes.InvalidArgument, "departmentID is required")
	}
	return nil
}

func roleToProto(role string) pb.Role {
	switch role {
	case "admin":
		return pb.Role_ROLE_ADMIN
	case "moderator":
		return pb.Role_ROLE_MODERATOR
	default:
		return pb.Role_ROLE_USER
	}
}
