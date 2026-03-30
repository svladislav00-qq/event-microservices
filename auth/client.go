package auth

import (
	"context"
	"fmt"

	"github.com/svladislav00-qq/event-microservices/auth/pb"
	"github.com/svladislav00-qq/event-microservices/pkg/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.AuthServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := pb.NewAuthServiceClient(conn)
	return &Client{
		conn:    conn,
		service: c,
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) Register(ctx context.Context, email string, password string, username string) (*Account, error) {
	r, err := c.service.Register(
		ctx,
		&pb.RegisterRequest{
			Email:    email,
			Password: password,
			Username: username,
		},
	)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, fmt.Errorf("empty account in response")
	}
	return &Account{
		ID:        r.Account.Id,
		Email:     r.Account.Email,
		Username:  r.Account.Username,
		Role:      r.Account.Role.String(),
		CreatedAt: r.Account.CreatedAt.AsTime(),
	}, nil
}

func (c *Client) Login(ctx context.Context, email, password string) (string, error) {
	r, err := c.service.Login(ctx, &pb.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", err
	}

	return r.Token, nil
}

func (c *Client) PromoteToModerator(ctx context.Context, userID string, departmentID string) (*Account, error) {
	r, err := c.service.PromoteToModerator(ctx, &pb.PromoteToModeratorRequest{
		UserId:     userID,
		Department: departmentID,
	})
	if err != nil {
		return nil, err
	}

	acc := r.GetAccount()
	if acc == nil {
		return nil, fmt.Errorf("empty account in response")
	}
	return &Account{
		ID:         acc.Id,
		Email:      acc.Email,
		Username:   acc.Username,
		Role:       acc.Role.String(),
		Department: acc.Department,
		CreatedAt:  acc.CreatedAt.AsTime(),
	}, nil
}

func (c *Client) GetUsersByIDs(ctx context.Context, ids []string) ([]models.Account, error) {
	resp, err := c.service.GetUsersByIDs(ctx, &pb.GetUsersByIDsRequest{
		UserIds: ids,
	})
	if err != nil {
		return nil, err
	}

	accounts := make([]models.Account, 0, len(resp.Users))

	for _, a := range resp.Users {
		accounts = append(accounts, models.Account{
			ID:       a.Id,
			Username: a.Name,
		})
	}

	return accounts, nil
}
