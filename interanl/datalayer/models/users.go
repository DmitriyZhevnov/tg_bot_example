package models

import "github.com/google/uuid"

type Users struct {
	ID     uuid.UUID
	Name   string
	ChatID int64
	Role   int // 0 - гость, 1 - юзер, 2 - админ, 3 - суперАдмин
}
