package main

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	ID         string         `json:"id"`
	Email      string         `json:"email"`
	Password   string         `json:"-"`
	Username   string         `json:"username"`
	Role       string         `json:"role"`
	Department *string        `json:"department,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`
}
