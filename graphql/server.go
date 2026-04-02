package graphql

import (
	attendeepb "github.com/svladislav00-qq/event-microservices/attendee/pb"
	authpb "github.com/svladislav00-qq/event-microservices/auth/pb"
	eventpb "github.com/svladislav00-qq/event-microservices/event/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	authClient     authpb.AuthServiceClient
	eventClient    eventpb.EventServiceClient
	attendeeClient attendeepb.AttendeeServiceClient
}

func NewServer(authURL, eventURL, attendeeURL string) (*Server, error) {
	authConn, err := grpc.NewClient(authURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	eventConn, err := grpc.NewClient(eventURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	attendeeConn, err := grpc.NewClient(attendeeURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Server{
		authClient:     authpb.NewAuthServiceClient(authConn),
		eventClient:    eventpb.NewEventServiceClient(eventConn),
		attendeeClient: attendeepb.NewAttendeeServiceClient(attendeeConn),
	}, nil
}
