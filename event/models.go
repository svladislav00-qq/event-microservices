package event

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	ID          string `json:"id" gorm:"primaryKey"`
	EventName   string `json:"event_name" gorm:"not null"`
	Description string `json:"description"`

	Department string `json:"department" gorm:"index"`
	CreatedBy  string `json:"created_by" gorm:"index"`

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	Files []EventFile `json:"files,omitempty" gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE"`
}

type EventFile struct {
	ID       string `json:"id" gorm:"primaryKey"`
	EventID  string `json:"event_id" gorm:"index;not null"`
	FileKey  string `json:"file_key" gorm:"not null"`
	FileType string `json:"file_type" gorm:"index"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type EventFilter struct {
	Department string
	CreatedBy  string

	Skip uint64
	Take uint64
}
