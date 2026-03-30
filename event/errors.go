package event

import "errors"

var (
	ErrFileNotFound       = errors.New("file not found")
	ErrEventNotFound      = errors.New("event not found")
	ErrEventAlreadyExists = errors.New("event already exists")
	ErrFileTooLarge       = errors.New("file too large")
	ErrUnauthorized       = errors.New("user unauthorized")
	ErrPermissionDenied   = errors.New("forbidden")
)
