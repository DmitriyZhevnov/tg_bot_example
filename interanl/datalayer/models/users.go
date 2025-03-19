package models

import "github.com/google/uuid"

type Users struct {
	ID     uuid.UUID
	Name   string
	ChatID int64
}
