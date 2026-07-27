package attendee

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/svladislav00-qq/event-microservices/pkg/logger/sl"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository interface {
	Close()
	RegisterToEvent(ctx context.Context, a Attendee) error
	CancelRegistration(ctx context.Context, userID string, eventID string) error
	GetUserRegistrations(ctx context.Context, af AttendeeFilter, userID string) ([]Attendee, error)
	GetEventRegistrations(ctx context.Context, skip uint64, take uint64, eventID string) ([]Attendee, error)
	GetAttendeeById(ctx context.Context, attendeeID string) (*Attendee, error)
	GetAttendeeByUserID(ctx context.Context, userID string) (*Attendee, error)
	GetRegisteredUserIDsByEvent(ctx context.Context, eventID string) ([]string, error)
	RepeatRegistration(ctx context.Context, userID string, eventID string) error
}

type postgresRepository struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewPostgresRepository(log *slog.Logger, databaseURL string) (Repository, error) {
	const op = "attendee.repository.NewPostgresRepository"

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.AutoMigrate(&Attendee{})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to migrate: %w", op, err)
	}

	return &postgresRepository{
		db:  db,
		log: log,
	}, nil
}

func (r *postgresRepository) Close() {
	sqlDB, err := r.db.DB()
	if err != nil {
		return
	}

	sqlDB.Close()
}

func (r *postgresRepository) RegisterToEvent(ctx context.Context, a Attendee) error {
	const op = "attendee.repository.RegisterToEvent"

	err := r.db.WithContext(ctx).Create(a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			r.log.Error("attendee already exists", slog.String("op", op), sl.Err(err))
			return ErrAttendeeAlreadyExists
		}
		r.log.Error("failed to register to event", slog.String("op", op), sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *postgresRepository) CancelRegistration(ctx context.Context, userID string, eventID string) error {
	const op = "attendee.repository.CancelRegistration"

	result := r.db.WithContext(ctx).Model(&Attendee{}).Where("user_id = ? AND event_id = ? AND status = ?", userID, eventID, StatusRegistered).Update("status", StatusCanceled)
	if result.Error != nil {
		r.log.Error("failed to cancel registration", slog.String("op", op), slog.String("event id", eventID), sl.Err(result.Error))
		return fmt.Errorf("%s: %w", op, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("%s: attendee not found", op)
	}

	return nil
}

func (r *postgresRepository) GetUserRegistrations(ctx context.Context, af AttendeeFilter, userID string) ([]Attendee, error) {
	const op = "attendee.repository.GetUserRegistrations"

	db := r.db.WithContext(ctx).Model(&Attendee{}).Where("user_id = ?", userID)

	if af.Status != "" {
		db = db.Where("status = ?", af.Status)
	}

	if af.Take == 0 {
		af.Take = 10
	}

	db = db.Order("registered_at DESC, id DESC").Offset(int(af.Skip)).Limit(int(af.Take))

	var attendees []Attendee
	if err := db.Find(&attendees).Error; err != nil {
		r.log.Error("failed to get user registrations", slog.String("op", op), sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return attendees, nil
}

func (r *postgresRepository) GetEventRegistrations(ctx context.Context, skip uint64, take uint64, eventID string) ([]Attendee, error) {
	const op = "attendee.repository.GetEventAttendee"

	db := r.db.WithContext(ctx).Model(&Attendee{}).Where("event_id = ? AND status = ?", eventID, StatusRegistered)

	if take == 0 {
		take = 10
	}

	db = db.Order("registered_at DESC, id DESC").Limit(int(take)).Offset(int(skip))

	var attendees []Attendee
	err := db.Find(&attendees).Error
	if err != nil {
		r.log.Error("failed to get attendees for event", slog.String("op", op), sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return attendees, nil
}

func (r *postgresRepository) GetAttendeeById(ctx context.Context, attendeeID string) (*Attendee, error) {
	const op = "attendee.repository.GetAttendeeById"

	var attendee Attendee
	err := r.db.WithContext(ctx).Preload("Attendees").Where("id = ?", attendeeID).First(&attendee).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("attendee not found", slog.String("op", op), slog.String("attendeeID", attendeeID), sl.Err(err))
			return nil, ErrAttendeeNotFound
		}
		r.log.Error("failed to get attendee", slog.String("op", op), slog.String("attendeeID", attendeeID), sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &attendee, nil
}

func (r *postgresRepository) GetAttendeeByUserID(ctx context.Context, userID string) (*Attendee, error) {
	const op = "attendee.repository.GetAttendeeByUserID"

	var attendee Attendee
	err := r.db.WithContext(ctx).Preload("Attendees").Where("user_id = ?", userID).First(&attendee).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("attendee not found", slog.String("op", op), slog.String("userID", userID), sl.Err(err))
			return nil, ErrAttendeeNotFound
		}
		r.log.Error("failed to get attendee", slog.String("op", op), slog.String("userID", userID), sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &attendee, nil
}

func (r *postgresRepository) GetAttendeeByEvent(ctx context.Context, eventID string) (*Attendee, error) {
	const op = "attendee.repository.GetAttendeeByEvent"

	var attendee Attendee
	err := r.db.WithContext(ctx).Preload("Attendees").Where("event_id = ?", eventID).First(&attendee).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("attendee not found", slog.String("op", op), slog.String("eventID", eventID), sl.Err(err))
			return nil, ErrAttendeeNotFound
		}
		r.log.Error("failed to get attendee", slog.String("op", op), slog.String("userID", eventID), sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &attendee, nil
}

func (r *postgresRepository) RepeatRegistration(ctx context.Context, userID string, eventID string) error {
	const op = "attendee.repository.RepeatRegistration"

	result := r.db.WithContext(ctx).Where("event_id = ? AND user_id=? AND status = ?", eventID, userID, StatusCanceled).Update("status", StatusRegistered)
	if result.Error != nil {
		r.log.Error("failed to repeat registration", slog.String("op", op), slog.String("event_id", eventID), sl.Err(result.Error))
		return fmt.Errorf("%s: %w", op, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("%s: attendee not found", op)
	}
	return nil
}

func (r *postgresRepository) GetRegisteredUserIDsByEvent(ctx context.Context, eventID string) ([]string, error) {
	const op = "attendee.repository.GetRegisteredUserIDsByEvent"

	var userIDs []string
	err := r.db.WithContext(ctx).Model(&Attendee{}).Where("event_id = ? AND status = ?", eventID, StatusRegistered).Pluck("user_id", &userIDs).Error
	if err != nil {
		slog.Error("failed to get userIDs",
			slog.String("op", op),
			slog.String("eventOD", eventID),
			sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return userIDs, nil
}
