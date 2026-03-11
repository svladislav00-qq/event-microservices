package event

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/svladislav00-qq/event-microservices/pkg/logger/sl"
)

type EventService struct {
	log          *slog.Logger
	evntSaver    EventSaver
	evntProvider EventProvider
	flSaver      FileSaver
}

type EventSaver interface {
	CreateEvent(ctx context.Context, e *Event) error
}
type EventProvider interface {
	UpdateEvent(ctx context.Context, eventID string, data map[string]interface{}) (*Event, error)
	DeleteEvent(ctx context.Context, eventID string) error
	GetEventByID(ctx context.Context, eventID string) (*Event, error)
	GetEventsByDepartment(ctx context.Context, department string, skip uint64, take uint64) ([]Event, error)
	GetEvents(ctx context.Context, filter EventFilter) ([]Event, error)
}
type FileSaver interface {
	UploadFileData(ctx context.Context, f *EventFile) error
	GetFileImageKeys(ctx context.Context, eventID string) ([]string, error)
	UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
	GetURL(objectName string) string
	DeleteFile(ctx context.Context, objectName string) error
}

func NewEventService(log *slog.Logger, evntSaver EventSaver, evntProvider EventProvider, flSaver FileSaver) *EventService {
	return &EventService{
		log:          log,
		evntSaver:    evntSaver,
		evntProvider: evntProvider,
		flSaver:      flSaver,
	}
}

func (e *EventService) CreateEvent(ctx context.Context, name string, description string, filesURLs string, startTime time.Time, endTime time.Time) (*Event, error) {
	const op = "event.service.CreateEvent"

	log := e.log.With(
		slog.String("op", op),
		slog.String("name", name),
	)

	log.Info("creating event")

	newEvent := &Event{
		ID:          ksuid.New().String(),
		EventName:   name,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := e.evntSaver.CreateEvent(ctx, newEvent); err != nil {
		if errors.Is(err, ErrEventAlreadyExists) {
			log.Warn("event already exists", sl.Err(err))
			return nil, ErrEventAlreadyExists
		}
		log.Error("failed to save event", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return newEvent, nil
}

func (e *EventService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, eventID string) (string, error) {
	const op = "event.service.UploadFile"

	log := e.log.With(
		slog.String("op", op),
		slog.String("eventID", eventID),
	)

	log.Info("starting upload files")

	_, err := e.evntProvider.GetEventByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, ErrEventNotFound) {
			log.Error("event not found", sl.Err(err))
			return "", ErrEventNotFound
		}
		log.Error("failed to get event", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if header.Size > 5<<20 {
		log.Error("file too large", sl.Err(err))
		return "", ErrFileTooLarge
	}

	contentType := filepath.Ext(header.Filename)
	fileKey := fmt.Sprintf("%s%s", ksuid.New().String(), contentType)

	_, err = e.flSaver.UploadFile(
		ctx,
		fileKey,
		file,
		header.Size,
		header.Header.Get("Content-Type"),
	)
	if err != nil {
		log.Error("failed to upload file in cloud", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	fileData := &EventFile{
		ID:       ksuid.New().String(),
		EventID:  eventID,
		FileKey:  fileKey,
		FileType: header.Header.Get("Content-Type"),
	}

	err = e.flSaver.UploadFileData(ctx, fileData)
	if err != nil {
		_ = e.flSaver.DeleteFile(ctx, fileKey)
		log.Error("failed to save file data", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	url := e.flSaver.GetURL(fileKey)

	return url, nil
}

func (e *EventService) AttachEventFile(ctx context.Context, eventID string, fileKeys []string) (*Event, error) {
	const op = "event.service.AttachEventFile"

	log := e.log.With(
		slog.String("op", op),
		slog.String("eventID", eventID),
	)

	log.Info("attaching file to event")

	event, err := e.evntProvider.GetEventByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, ErrEventNotFound) {
			log.Error("event not found", sl.Err(err))
			return nil, ErrEventNotFound
		}

		log.Error("failed to get event", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	changes := map[string]any{
		"files": fileKeys,
	}

	updatedEvent, err := e.evntProvider.UpdateEvent(ctx, event.ID, changes)
	if err != nil {
		log.Error("failed to update events files", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("files successfully attached to event")

	return updatedEvent, nil
}

func (e *EventService) DeleteEvent(ctx context.Context, eventID string) error {
	const op = "event.service.DeleteEvent"

	log := e.log.With(
		slog.String("op", op),
		slog.String("eventID", eventID),
	)

	log.Info("starting delete event")

	event, err := e.evntProvider.GetEventByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, ErrEventNotFound) {
			log.Error("event not found", sl.Err(err))
			return ErrEventNotFound
		}

		log.Error("failed to get event", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	for _, file := range event.Files {
		err := e.flSaver.DeleteFile(ctx, file.FileKey)
		if err != nil {
			log.Error("failed to delete file", slog.String("fileKey", file.FileKey), sl.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := e.evntProvider.DeleteEvent(ctx, eventID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("event and file successfully deleted")

	return nil
}

func (e *EventService) GetEvent(ctx context.Context, eventID string) (*Event, error) {
	const op = "event.service.GetEvent"

	log := e.log.With(
		slog.String("op", op),
		slog.String("eventID", eventID),
	)

	log.Info("starting to get event")

	event, err := e.evntProvider.GetEventByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, ErrEventNotFound) {
			log.Error("event not found", sl.Err(err))
			return nil, ErrEventNotFound
		}

		log.Error("failed to get event", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("event got successfully")
	return event, nil
}

func (e *EventService) GetEvents(ctx context.Context, filter EventFilter) ([]Event, error) {
	const op = "event.service.GetEvents"

	log := e.log.With(
		slog.String("op", op),
	)

	log.Info("starting to get events with pagination")

	events, err := e.evntProvider.GetEvents(ctx, filter)
	if err != nil {
		log.Error("failed to get events", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return events, nil
}
