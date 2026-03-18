package attendee

import "errors"

var (
	ErrAttendeeNotFound      = errors.New("attendee not found")
	ErrAttendeeAlreadyExists = errors.New("attendee already exists")
)
