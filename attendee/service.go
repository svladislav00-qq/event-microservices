package attendee

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/svladislav00-qq/event-microservices/pkg/logger/sl"
	"github.com/svladislav00-qq/event-microservices/pkg/models"
	"google.golang.org/grpc/metadata"
)

type AttendeeService struct {
	log           *slog.Logger
	attendeeSaver AttendeeSaver
	userProvider  UserProvider
}

type AttendeeSaver interface {
	RegisterToEvent(ctx context.Context, a Attendee) error
	CancelRegistration(ctx context.Context, userID string, eventID string) error
	GetUserRegistrations(ctx context.Context, af AttendeeFilter, userID string) ([]Attendee, error)
	GetEventRegistrations(ctx context.Context, skip uint64, take uint64, eventID string) ([]Attendee, error)
	GetAttendeeById(ctx context.Context, attendeeID string) (*Attendee, error)
	GetAttendeeByUserID(ctx context.Context, userID string) (*Attendee, error)
	RepeatRegistration(ctx context.Context, userID string, eventID string) error
	GetRegisteredUserIDsByEvent(ctx context.Context, eventID string) ([]string, error)
}

type UserProvider interface {
	GetUsersByIDs(ctx context.Context, ids []string) ([]models.Account, error)
}

func NewAttendeeService(log *slog.Logger, attendeeSaver AttendeeSaver, userProvider UserProvider) *AttendeeService {
	return &AttendeeService{
		log:           log,
		attendeeSaver: attendeeSaver,
		userProvider:  userProvider,
	}
}

func (a *AttendeeService) RegisterToEvent(ctx context.Context, userID string, eventID string) (*Attendee, error) {
	const op = "attendee.service.RegisterToEvent"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("registering to event")

	newAttendee := &Attendee{
		ID:           ksuid.New().String(),
		EventID:      eventID,
		UserID:       userID,
		Status:       StatusRegistered,
		RegisteredAt: time.Now(),
	}

	if err := a.attendeeSaver.RegisterToEvent(ctx, *newAttendee); err != nil {
		if errors.Is(err, ErrAttendeeAlreadyExists) {
			log.Error("user already registered status changed to registered")
			if err := a.attendeeSaver.RepeatRegistration(ctx, newAttendee.UserID, newAttendee.EventID); err != nil {
				log.Error("failed to repeat registration", sl.Err(err))
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			log.Info("user already registered, status restored")
			return newAttendee, nil
		}
		log.Error("failed to registered user", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return newAttendee, nil
}

func (a *AttendeeService) CancelRegistration(ctx context.Context, userID string, eventID string) error {
	const op = "attendee.service.CancelRegistration"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("cancel registration to event")

	err := a.attendeeSaver.CancelRegistration(ctx, userID, eventID)
	if err != nil {
		log.Error("failed to cancel registration", sl.Err(err))
		return err
	}

	log.Info("registration successfully canceled")
	return nil
}

func (a *AttendeeService) GetUserRegistrations(ctx context.Context, af AttendeeFilter, userID string) ([]Attendee, error) {
	const op = "attendee.service.GetUserRegistrations"

	log := a.log.With(
		slog.String("op", op),
		slog.String("userID", userID),
	)

	log.Info("starting to get user's registrations")

	userRegistrations, err := a.attendeeSaver.GetUserRegistrations(ctx, af, userID)
	if err != nil {
		log.Error("failed to get user's registrations", slog.String("op", op), sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return userRegistrations, nil
}

func (a *AttendeeService) GetEventRegistrations(ctx context.Context, skip uint64, take uint64, eventID string) ([]Attendee, error) {
	const op = "attendee.service.GetEventRegistrations"

	log := a.log.With(
		slog.String("op", op),
		slog.String("eventID", eventID),
	)

	log.Info("starting to get attendees for event")

	eventRegistrations, err := a.attendeeSaver.GetEventRegistrations(ctx, skip, take, eventID)
	if err != nil {
		log.Error("failed to get attendees for event", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return eventRegistrations, nil
}

func (a *AttendeeService) ExportAttendeesTable(ctx context.Context, eventID string) ([]string, error) {
	const op = "attendee.service.ExportAttendeesTable"

	log := a.log.With(
		slog.String("op", op),
		slog.String("eventID", eventID),
	)

	log.Info("starting to export attendees table")

	userIDs, err := a.attendeeSaver.GetRegisteredUserIDsByEvent(ctx, eventID)
	if err != nil {
		log.Error("failed to get userIDs", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(userIDs) == 0 {
		return []string{}, nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata")
	}

	outCtx := metadata.NewOutgoingContext(ctx, md)

	users, err := a.userProvider.GetUsersByIDs(outCtx, userIDs)
	if err != nil {
		log.Error("failed  to get accounts", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	table := make([]string, 0, len(users)+1)
	table = append(table, "id,username")

	userMap := make(map[string]models.Account, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}

	for _, id := range userIDs {
		u, ok := userMap[id]
		if !ok {
			log.Warn("user not found", slog.String("userID", id))
			continue
		}

		row := fmt.Sprintf("%s,%s", u.ID, u.Username)
		table = append(table, row)
	}

	return table, nil
}
