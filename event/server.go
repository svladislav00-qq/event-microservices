package event

import (
	"context"
	"time"

	"github.com/svladislav00-qq/event-microservices/event/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EventServices interface {
	CreateEvent(ctx context.Context, name string, description string, filesURLs []string, capacity *int, startTime time.Time, endTime time.Time) (*Event, error)
	UpdateEvent(ctx context.Context, eventID string, updateData map[string]interface{}) (*Event, error)
	DeleteEvent(ctx context.Context, eventID string) error
	GetEvent(ctx context.Context, eventID string) (*Event, error)
	GetEvents(ctx context.Context, filter EventFilter) ([]Event, error)
	UploadFile(ctx context.Context, eventID string, fileName string, data []byte) (string, string, error)
	AttachEventFile(ctx context.Context, eventID string, fileKeys []string) (*Event, error)
}

type serverEvent struct {
	pb.UnimplementedEventServiceServer
	event EventServices
}

func NewEventServer(grpc *grpc.Server, eventServices EventServices) {
	pb.RegisterEventServiceServer(grpc, &serverEvent{event: eventServices})
}

const maxFileSize = 5 << 20

func (s *serverEvent) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*pb.CreateEventResponse, error) {
	if err := validateCreateEvent(req); err != nil {
		return nil, err
	}

	var capacity *int
	if req.Capacity != nil {
		c := int(req.GetCapacity())
		capacity = &c
	}

	res, err := s.event.CreateEvent(
		ctx,
		req.Name,
		req.Description,
		req.FileUrls,
		capacity,
		req.StartTime.AsTime(),
		req.EndTime.AsTime(),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateEventResponse{
		Event: eventToPB(res),
	}, nil
}

func (s *serverEvent) UpdateEvent(ctx context.Context, req *pb.UpdateEventRequest) (*pb.UpdateEventResponse, error) {
	if err := validateUpdateEvent(req); err != nil {
		return nil, err
	}

	updateData := make(map[string]interface{})

	if req.GetName() != "" {
		updateData["name"] = req.GetName()
	}

	if req.GetDescription() != "" {
		updateData["description"] = req.Description
	}

	if req.GetDepartment() != "" {
		updateData["department"] = req.Department
	}

	if req.GetStartTime() != nil {
		updateData["start_time"] = req.StartTime
	}

	if req.GetEndTime() != nil {
		updateData["end_time"] = req.EndTime
	}

	if req.GetCapacity() != 0 {
		updateData["capacity"] = req.Capacity
	}

	updatedEvent, err := s.event.UpdateEvent(ctx, req.Id, updateData)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateEventResponse{Event: eventToPB(updatedEvent)}, nil
}

func (s *serverEvent) DeleteEvent(ctx context.Context, req *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
	if err := validateDeleteEvent(req); err != nil {
		return nil, err
	}

	err := s.event.DeleteEvent(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteEventResponse{}, nil
}

func (s *serverEvent) GetEvent(ctx context.Context, req *pb.GetEventRequest) (*pb.GetEventResponse, error) {
	if err := validateGetEvent(req); err != nil {
		return nil, err
	}

	event, err := s.event.GetEvent(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.GetEventResponse{Event: eventToPB(event)}, nil
}

func (s *serverEvent) GetEvents(ctx context.Context, req *pb.GetEventsRequest) (*pb.GetEventsResponse, error) {
	filter := EventFilter{
		Skip:       req.GetSkip(),
		Take:       req.GetTake(),
		Department: req.GetDepartment(),
		CreatedBy:  req.GetCreatedBy(),
	}

	events, err := s.event.GetEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &pb.GetEventsResponse{
		Events: eventsToPB(events),
	}, nil
}

func (s *serverEvent) UploadFiles(ctx context.Context, req *pb.UploadFilesRequest) (*pb.UploadFilesResponse, error) {
	if err := validateUploadFiles(req); err != nil {
		return nil, err
	}

	var responses []*pb.UploadFileResponse

	for _, file := range req.Files {
		if len(file.FileData) == 0 {
			return nil, status.Error(codes.InvalidArgument, "file_data required")
		}

		if len(file.FileData) > maxFileSize {
			return nil, status.Error(codes.InvalidArgument, "file too large")
		}

		url, fileKey, err := s.event.UploadFile(
			ctx,
			req.Id,
			file.FileName,
			file.FileData,
		)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		responses = append(responses, &pb.UploadFileResponse{
			UploadUrl: url,
			FileKey:   fileKey,
		})
	}

	return &pb.UploadFilesResponse{
		Files: responses,
	}, nil
}

func (s *serverEvent) AttachEventFile(ctx context.Context, req *pb.AttachFileToEventRequest) (*pb.AttachFileToEventResponse, error) {
	if err := validateAttachEventFile(req); err != nil {
		return nil, err
	}

	updatedEvent, err := s.event.AttachEventFile(ctx, req.Id, req.FileKeys)
	if err != nil {
		return nil, err
	}

	var fileURLs []string
	for _, f := range updatedEvent.Files {
		fileURLs = append(fileURLs, f.FileKey)
	}

	return &pb.AttachFileToEventResponse{Event: eventToPB(updatedEvent)}, nil
}

func validateGetEvent(req *pb.GetEventRequest) error {
	if req.Id == "" {
		return status.Error(codes.InvalidArgument, "event ID is required")
	}
	return nil
}

func validateDeleteEvent(req *pb.DeleteEventRequest) error {
	if req.Id == "" {
		return status.Error(codes.InvalidArgument, "event ID is required")
	}
	return nil
}

func validateUpdateEvent(req *pb.UpdateEventRequest) error {
	if req.Id == "" {
		return status.Error(codes.InvalidArgument, "event ID is required")
	}
	return nil
}

func validateAttachEventFile(req *pb.AttachFileToEventRequest) error {
	if req.Id == "" {
		return status.Error(codes.InvalidArgument, "event ID is required")
	}
	if req.FileKeys == nil {
		return status.Error(codes.InvalidArgument, "file keys are required")
	}
	return nil
}

func validateUploadFiles(req *pb.UploadFilesRequest) error {
	if req.Id == "" {
		return status.Error(codes.InvalidArgument, "event ID is required")
	}

	if len(req.Files) == 0 {
		return status.Error(codes.InvalidArgument, "file data is required")
	}
	return nil
}

func validateCreateEvent(req *pb.CreateEventRequest) error {
	if req.GetName() == "" {
		return status.Error(codes.InvalidArgument, "event name is required")
	}
	if req.GetDescription() == "" {
		return status.Error(codes.InvalidArgument, "event description is required")
	}
	if req.GetStartTime() == nil {
		return status.Error(codes.InvalidArgument, "event start time is required")
	}
	if req.GetEndTime() == nil {
		return status.Error(codes.InvalidArgument, "event end time is required")
	}
	return nil
}
