package event

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/svladislav00-qq/event-microservices/event/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.EventServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := pb.NewEventServiceClient(conn)
	return &Client{
		conn:    conn,
		service: c,
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) CreateEvent(ctx context.Context, name string, description string, startTime *timestamppb.Timestamp, endTime *timestamppb.Timestamp, capacity *uint32) (*Event, error) {
	r, err := c.service.CreateEvent(
		ctx,
		&pb.CreateEventRequest{
			Name:        name,
			Description: description,
			StartTime:   startTime,
			EndTime:     endTime,
			Capacity:    capacity,
		},
	)
	if err != nil {
		return nil, err
	}

	if r == nil || r.Event == nil {
		return nil, fmt.Errorf("empty event in response")
	}

	return pbEventToModel(r.Event), nil
}

func (c *Client) UpdateEvent(ctx context.Context, eventID string, updateData map[string]interface{}) (*Event, error) {
	req := &pb.UpdateEventRequest{
		Id: eventID,
	}

	if v, ok := updateData["name"].(string); ok {
		req.Name = v
	}

	if v, ok := updateData["description"].(string); ok {
		req.Description = v
	}

	if v, ok := updateData["department"].(string); ok {
		req.Department = v
	}

	if v, ok := updateData["start_time"].(time.Time); ok {
		req.StartTime = timestamppb.New(v)
	}

	if v, ok := updateData["end_time"].(time.Time); ok {
		req.EndTime = timestamppb.New(v)
	}
	if v, ok := updateData["capacity"].(int); ok {
		val := uint32(v)
		req.Capacity = &val
	}

	resp, err := c.service.UpdateEvent(ctx, req)
	if err != nil {
		return nil, err
	}

	return pbEventToModel(resp.Event), nil
}

func (c *Client) DeleteEvent(ctx context.Context, eventID string) error {
	r, err := c.service.DeleteEvent(
		ctx,
		&pb.DeleteEventRequest{
			Id: eventID,
		},
	)
	if err != nil {
		return err
	}

	if r == nil {
		return fmt.Errorf("failed to delete event")
	}

	return nil
}

func (c *Client) GetEvent(ctx context.Context, eventID string) (*Event, error) {
	r, err := c.service.GetEvent(
		ctx,
		&pb.GetEventRequest{
			Id: eventID,
		},
	)
	if err != nil {
		return nil, err
	}
	if r == nil || r.Event == nil {
		return nil, fmt.Errorf("empty event in response")
	}

	return pbEventToModel(r.Event), nil
}

func (c *Client) GetEvents(ctx context.Context, filter EventFilter) ([]*Event, error) {
	r, err := c.service.GetEvents(
		ctx,
		&pb.GetEventsRequest{
			Skip:       filter.Skip,
			Take:       filter.Take,
			Department: filter.Department,
			CreatedBy:  filter.CreatedBy,
		},
	)
	if err != nil {
		return nil, err
	}
	if r == nil || r.Events == nil {
		return nil, fmt.Errorf("empty response")
	}

	return eventsFromPB(r.Events), nil
}

func (c *Client) AttachEventFile(ctx context.Context, eventID string, fileKeys []string) (*Event, error) {
	res, err := c.service.AttachFileToEvent(
		ctx,
		&pb.AttachFileToEventRequest{
			Id:       eventID,
			FileKeys: fileKeys,
		},
	)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Event == nil {
		return nil, fmt.Errorf("empty response")
	}
	return pbEventToModel(res.Event), nil
}

func (c *Client) UploadFiles(ctx context.Context, eventID string, files []*multipart.FileHeader) ([]*pb.UploadFileResponse, error) {
	var uploads []*pb.FileUpload

	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			return nil, err
		}

		data, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		uploads = append(uploads, &pb.FileUpload{
			FileName: fh.Filename,
			FileData: data,
		})
	}

	resp, err := c.service.UploadFiles(
		ctx, &pb.UploadFilesRequest{
			Id:    eventID,
			Files: uploads,
		},
	)
	if err != nil {
		return nil, err
	}

	return resp.Files, nil
}
