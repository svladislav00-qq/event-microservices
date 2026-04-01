package event

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
	GetEvents(ctx context.Context, filter EventFilter) ([]Event, error)
	GetFileImageKeys(ctx context.Context, eventID string) ([]string, error)
	UploadFileData(ctx context.Context, f *EventFile) error
	DeleteFile(ctx context.Context, eventID string) error
}
type FileSaver interface {
	UploadFile(ctx context.Context, fileKey string, reader io.Reader, size int64, contentType string) (string, error)
	GetURL(fileKey string) string
	DeleteFile(ctx context.Context, fileKey string) error
}

func NewEventService(log *slog.Logger, evntSaver EventSaver, evntProvider EventProvider, flSaver FileSaver) *EventService {
	return &EventService{
		log:          log,
		evntSaver:    evntSaver,
		evntProvider: evntProvider,
		flSaver:      flSaver,
	}
}

func (e *EventService) CreateEvent(ctx context.Context, name string, description string, fileKeys []string, capacity *int, startTime time.Time, endTime time.Time, department string, createdBy string, role string) (*Event, error) {
	const op = "event.service.CreateEvent"

	log := e.log.With(
		slog.String("op", op),
		slog.String("name", name),
	)

	log.Info("creating event")

	if roleKey == "user" {
		return nil, ErrPermissionDenied
	}

	newEvent := &Event{
		ID:          ksuid.New().String(),
		EventName:   name,
		Description: description,
		Department:  department,
		CreatedBy:   createdBy,
		StartTime:   startTime,
		EndTime:     endTime,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Capacity:    capacity,
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

func (e *EventService) UploadFile(ctx context.Context, eventID string, fileName string, data []byte) (string, string, error) {
	const op = "event.service.UploadFile"

	log := e.log.With(
		slog.String("op", op),
		slog.String("eventID", eventID),
	)

	log.Info("starting upload files")

	if len(data) == 0 {
		return "", "", fmt.Errorf("%s: empty file", op)
	}

	_, err := e.evntProvider.GetEventByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, ErrEventNotFound) {
			log.Error("event not found", sl.Err(err))
			return "", "", ErrEventNotFound
		}
		log.Error("failed to get event", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	reader := bytes.NewReader(data)

	fileKey := ksuid.New().String() + filepath.Ext(fileName)
	contentType := http.DetectContentType(data)

	_, err = e.flSaver.UploadFile(
		ctx,
		fileKey,
		reader,
		int64(len(data)),
		contentType,
	)
	if err != nil {
		log.Error("failed to upload file", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	url := e.flSaver.GetURL(fileKey)

	file := &EventFile{
		ID:       ksuid.New().String(),
		EventID:  eventID,
		FileKey:  fileKey,
		FileType: http.DetectContentType(data),
	}

	err = e.evntProvider.UploadFileData(ctx, file)
	if err != nil {
		e.flSaver.DeleteFile(ctx, fileKey)
		log.Error("failed to save file metadata", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return url, fileKey, nil

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
		return fmt.Errorf("%s: %w", op, err)
	}

	for _, file := range event.Files {

		if err := e.flSaver.DeleteFile(ctx, file.FileKey); err != nil {
			log.Error("failed to delete file from storage",
				slog.String("fileKey", file.FileKey),
				sl.Err(err),
			)

			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := e.evntProvider.DeleteFile(ctx, eventID); err != nil {
		log.Error("failed to delete event files", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := e.evntProvider.DeleteEvent(ctx, eventID); err != nil {
		log.Error("failed to delete event", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("event and files successfully deleted")

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

	for i := range event.Files {
		event.Files[i].FileKey = e.flSaver.GetURL(event.Files[i].FileKey)
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

	for i := range events {
		for j := range events[i].Files {
			events[i].Files[j].FileKey = e.flSaver.GetURL(events[i].Files[j].FileKey)
		}
	}

	return events, nil
}

func (e *EventService) UpdateEvent(ctx context.Context, eventID string, updateData map[string]interface{}) (*Event, error) {
	const op = "event.service.UpdateEvent"

	log := e.log.With(
		slog.String("op", op),
		slog.String("eventID", eventID),
		slog.Any("updateData", updateData),
	)

	log.Info("starting to update event data")

	event, err := e.evntProvider.GetEventByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, ErrEventNotFound) {
			log.Error("event not found", sl.Err(err))
			return nil, ErrEventNotFound
		}

		log.Error("failed to get event", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = e.evntProvider.UpdateEvent(ctx, event.ID, updateData)
	if err != nil {
		log.Error("failed to update event", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return e.evntProvider.GetEventByID(ctx, eventID)
}
