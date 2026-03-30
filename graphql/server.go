package graphql

import (
	"github.com/svladislav00-qq/event-microservices/attendee"
	"github.com/svladislav00-qq/event-microservices/auth"
	"github.com/svladislav00-qq/event-microservices/event"
)

type Server struct {
	authClient     *auth.Client
	eventClient    *event.Client
	attendeeClient *attendee.Client
}

func NewServer(authURL, eventURL, attendeeURL string) (*Server, error) {
	authClient, err := auth.NewClient(authURL)
	if err != nil {
		return nil, err
	}

	eventClient, err := event.NewClient(eventURL)
	if err != nil {
		return nil, err
	}

	attendeeClient, err := attendee.NewClient(attendeeURL)
	if err != nil {
		return nil, err
	}

	return &Server{
		authClient:     authClient,
		eventClient:    eventClient,
		attendeeClient: attendeeClient,
	}, nil
}
