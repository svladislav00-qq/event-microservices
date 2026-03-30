package attendee

import (
	"context"
	"fmt"

	"github.com/svladislav00-qq/event-microservices/attendee/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.AttendeeServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := pb.NewAttendeeServiceClient(conn)
	return &Client{
		conn:    conn,
		service: c,
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) RegisterToEvent(ctx context.Context, eventID string) (*Attendee, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}

	r, err := c.service.RegisterToEvent(
		ctx,
		&pb.RegisterToEventRequest{
			EventId: eventID,
			UserId:  userID,
		},
	)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, fmt.Errorf("empty attendee in response")
	}

	return pbAttendeeToModel(r.Attendee), nil
}

func (c *Client) CancelRegistration(ctx context.Context, eventID string) error {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return status.Error(codes.Unauthenticated, "no user in context")
	}

	_, err := c.service.CancelRegistration(
		ctx,
		&pb.CancelRegistrationRequest{
			EventId: eventID,
			UserId:  userID,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetUserRegistrations(ctx context.Context, skip uint32, take uint32) ([]Attendee, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}

	resp, err := c.service.GetUserRegistrations(
		ctx,
		&pb.GetUserRegistrationsRequest{
			Skip:   uint64(skip),
			Take:   uint64(take),
			UserId: userID,
		},
	)
	if err != nil {
		return nil, err
	}

	result := make([]Attendee, 0, len(resp.Events))

	for _, e := range resp.Events {
		result = append(result, Attendee{
			EventID:      e.EventId,
			Status:       mapStatusFromPB(e.Status),
			RegisteredAt: e.RegisteredAt.AsTime(),
		})
	}

	return result, nil
}

func (c *Client) GetEventAttendees(ctx context.Context, eventID string, skip uint32, take uint32) ([]Attendee, error) {
	resp, err := c.service.GetEventAttendees(
		ctx,
		&pb.GetEventAttendeesRequest{
			EventId: eventID,
			Skip:    uint64(skip),
			Take:    uint64(take),
		},
	)
	if err != nil {
		return nil, err
	}

	result := make([]Attendee, 0, len(resp.Attendees))

	for _, a := range resp.Attendees {
		result = append(result, Attendee{
			ID:           a.Id,
			EventID:      a.EventId,
			UserID:       a.UserId,
			Status:       mapStatusFromPB(a.Status),
			RegisteredAt: a.RegisteredAt.AsTime(),
		})
	}
	return result, nil
}

func (c *Client) ExportAttendeeTable(ctx context.Context, eventID string) ([]byte, string, error) {
	resp, err := c.service.ExportAttendeesTable(
		ctx,
		&pb.ExportAttendeeTableRequest{
			EventId: eventID,
		},
	)
	if err != nil {
		return nil, "", err
	}

	return resp.File, resp.Filename, nil
}
