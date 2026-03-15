package event

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
	EventRepository
	FileRepository
	Close()
}

type EventRepository interface {
	CreateEvent(ctx context.Context, e *Event) error
	UpdateEvent(ctx context.Context, eventID string, data map[string]interface{}) (*Event, error)
	DeleteEvent(ctx context.Context, eventID string) error
	GetEventByID(ctx context.Context, eventID string) (*Event, error)
	GetEvents(ctx context.Context, f EventFilter) ([]Event, error)
}

type FileRepository interface {
	UploadFileData(ctx context.Context, f *EventFile) error
	GetFileImageKeys(ctx context.Context, eventID string) ([]string, error)
	DeleteFile(ctx context.Context, fileKey string) error
}

type postgresRepository struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewPostgresRepository(log *slog.Logger, databaseURL string) (Repository, error) {
	const op = "event.repository.NewPostgresRepository"

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

	err = db.AutoMigrate(&Event{}, &EventFile{})
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
func (r *postgresRepository) CreateEvent(ctx context.Context, e *Event) error {
	const op = "event.repository.PostEvent"

	db := r.db.WithContext(ctx)

	if err := db.Create(e).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			r.log.Warn("event already exists", slog.String("op", op), sl.Err(err))
			return ErrEventAlreadyExists
		}
		r.log.Error("failed to save event", slog.String("op", op), sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *postgresRepository) DeleteEvent(ctx context.Context, eventID string) error {
	return r.db.WithContext(ctx).Delete(&Event{}, "id = ?", eventID).Error
}

func (r *postgresRepository) UpdateEvent(ctx context.Context, eventID string, data map[string]interface{}) (*Event, error) {
	const op = "event.repository.UpdateEvent"

	db := r.db.WithContext(ctx)

	var event Event
	if err := db.Preload("Event", "Files").Model(&Event{}).
		Where("id = ?", eventID).
		Updates(data).Error; err != nil {

		r.log.Error("failed to update event",
			slog.String("op", op),
			slog.String("event_id", eventID),
			sl.Err(err),
		)

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &event, nil
}

func (r *postgresRepository) GetEventByID(ctx context.Context, eventID string) (*Event, error) {
	const op = "event.repository.GetEventById"

	var event Event
	err := r.db.WithContext(ctx).Preload("Files").Where("id = ?", eventID).First(&event).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("event not found", slog.String("op", op), slog.String("eventID", eventID), sl.Err(err))
			return nil, ErrEventNotFound
		}
		r.log.Error("failed to get event", slog.String("op", op), slog.String("eventID", eventID), sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &event, nil
}

func (r *postgresRepository) GetEvents(ctx context.Context, filter EventFilter) ([]Event, error) {
	const op = "auth.repository.GetEvents"

	db := r.db.WithContext(ctx).Model(&Event{})

	if filter.Department != "" {
		db = db.Where("department = ?", filter.Department)
	}

	if filter.CreatedBy != "" {
		db = db.Where("created_by = ?", filter.CreatedBy)
	}

	if filter.Take == 0 {
		filter.Take = 10
	}

	db = db.Offset(int(filter.Skip)).Limit(int(filter.Take))

	var events []Event
	if err := db.Preload("Files").Find(&events).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return events, nil
}

func (r *postgresRepository) UploadFileData(ctx context.Context, f *EventFile) error {
	const op = "event.repository.UploadFileData"

	db := r.db.WithContext(ctx)

	if err := db.Create(f).Error; err != nil {
		r.log.Error("failed to save file data", slog.String("op", op), sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *postgresRepository) GetFileImageKeys(ctx context.Context, eventID string) ([]string, error) {
	const op = "event.repository.GetFileImageKeys"

	var fileKeys []string
	err := r.db.WithContext(ctx).Select("file_key").Where("event_id = ?", eventID).First(&EventFile{}).Error
	if err != nil {
		r.log.Error("failed to get file key", slog.String("op", op), sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(fileKeys) == 0 {
		r.log.Error("file not found", slog.String("op", op))
		return nil, ErrFileNotFound
	}

	return fileKeys, nil
}

func (r *postgresRepository) DeleteFile(ctx context.Context, eventID string) error {
	return r.db.WithContext(ctx).Where("event_id = ?", eventID).Delete(&EventFile{}).Error
}
