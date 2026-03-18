package attendee

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/svladislav00-qq/event-microservices/attendee/pb"
	"github.com/xuri/excelize/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AttendeeServices interface {
	RegisterToEvent(ctx context.Context, userID string, eventID string) (*Attendee, error)
	CancelRegistration(ctx context.Context, userID string, eventID string) error
	GetUserRegistrations(ctx context.Context, af AttendeeFilter, userID string) ([]Attendee, error)
	GetEventRegistrations(ctx context.Context, skip uint64, take uint64, eventID string) ([]Attendee, error)
	ExportAttendeesTable(ctx context.Context, eventID string) ([]string, error)
}

type serverAttendee struct {
	pb.UnimplementedAttendeeServiceServer
	attendee *AttendeeService
}

func NewServer(gRPC *grpc.Server, attendee *AttendeeService) *serverAttendee {
	return &serverAttendee{
		attendee: attendee,
	}
}

func (s *serverAttendee) RegisterToEvent(ctx context.Context, req *pb.RegisterToEventRequest) (*pb.RegisterToEventResponse, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.GetEventId() == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id is required")
	}

	attendee, err := s.attendee.RegisterToEvent(ctx, userID, req.GetEventId())
	if err != nil {
		return nil, err
	}
	return &pb.RegisterToEventResponse{Attendee: &pb.Attendee{
		Id:           attendee.ID,
		EventId:      attendee.EventID,
		UserId:       attendee.UserID,
		Status:       mapStatusToPB(attendee.Status),
		RegisteredAt: timestamppb.New(attendee.RegisteredAt),
	}}, nil
}

func (s *serverAttendee) CancelRegistration(ctx context.Context, req *pb.CancelRegistrationRequest) (*pb.CancelRegistrationResponse, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.GetEventId() == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id is required")
	}

	err := s.attendee.CancelRegistration(ctx, userID, req.EventId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to cancel registration")
	}

	return &pb.CancelRegistrationResponse{}, nil
}

func (s *serverAttendee) GetUserRegistrations(ctx context.Context, req *pb.GetUserRegistrationsRequest) (*pb.GetUserRegistrationsResponse, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	filter := AttendeeFilter{
		Skip:   req.Skip,
		Take:   req.Take,
		Status: string(StatusRegistered),
	}

	registrations, err := s.attendee.GetUserRegistrations(ctx, filter, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user's registrations")
	}

	pbEvents := make([]*pb.MyEvent, 0, len(registrations))

	for _, r := range registrations {
		pbEvents = append(pbEvents, &pb.MyEvent{
			EventId:      r.EventID,
			Status:       mapStatusToPB(r.Status),
			RegisteredAt: timestamppb.New(r.RegisteredAt),
		})
	}

	return &pb.GetUserRegistrationsResponse{
		Events: pbEvents,
	}, nil
}

func (s *serverAttendee) GetEventAttendees(ctx context.Context, req *pb.GetEventAttendeesRequest) (*pb.GetEventAttendeesResponse, error) {
	if req.GetEventId() == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id is required")
	}

	take := req.GetTake()
	if take == 0 {
		take = 10
	}

	skip := req.GetSkip()

	attendees, err := s.attendee.GetEventRegistrations(ctx, skip, take, req.GetEventId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get attendees from event")
	}

	pbAttendees := make([]*pb.Attendee, 0, len(attendees))

	for _, a := range attendees {
		pbAttendees = append(pbAttendees, &pb.Attendee{
			Id:           a.ID,
			EventId:      a.EventID,
			UserId:       a.UserID,
			Status:       mapStatusToPB(a.Status),
			RegisteredAt: timestamppb.New(a.RegisteredAt),
		})
	}

	return &pb.GetEventAttendeesResponse{Attendees: pbAttendees}, nil
}

func (s *serverAttendee) ExportAttendeesTable(ctx context.Context, req *pb.ExportAttendeeTableRequest) (*pb.ExportAttendeeTableResponse, error) {
	if req.GetEventId() == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id is required")
	}

	table, err := s.attendee.ExportAttendeesTable(ctx, req.GetEventId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to export attendees")
	}

	f := excelize.NewFile()
	sheet := "Sheet1"

	for i, row := range table {
		cols := strings.Split(row, ",")

		for j, col := range cols {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
			f.SetCellValue(sheet, cell, col)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, status.Error(codes.Internal, "failed to generate excel file")
	}

	return &pb.ExportAttendeeTableResponse{
		File:     buf.Bytes(),
		Filename: fmt.Sprintf("attendees_%s.xlsx", req.GetEventId()),
	}, nil
}
