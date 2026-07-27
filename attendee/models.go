package attendee

import "time"

type Attendee struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	EventID      string    `json:"event_id" gorm:"not null;index:idx_attendees_event_status_registered,priority:1"`
	UserID       string    `json:"user_id" gorm:"not null;index:idx_attendees_user_status_registered,priority:1"`
	Status       Status    `json:"status" gorm:"not null;index:idx_attendees_event_status_registered,priority:2;index:idx_attendees_user_status_registered,priority:2"`
	RegisteredAt time.Time `json:"registered_at" gorm:"not null;index:idx_attendees_event_status_registered,priority:3,sort:desc;index:idx_attendees_user_status_registered,priority:3,sort:desc"`
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
