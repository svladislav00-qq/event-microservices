package attendee

import "time"

type Attendee struct {
	ID           string    `json:"id"`
	EventID      string    `json:"event_id"`
	UserID       string    `json:"user_id"`
	Status       Status    `json:"status"`
	RegisteredAt time.Time `json:"registered_at"`
}

type Status string

const (
	StatusRegistered Status = "registered"
	StatusCanceled   Status = "canceled"
)

type AttendeeFilter struct {
	Skip uint64
	Take uint64

	Status string
}

