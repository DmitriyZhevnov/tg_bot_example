package collections

import (
	"context"
	"testbot/interanl/datalayer/models"
)

type (
	Users interface {
		GetByChatID(ctx context.Context, chatID int64) (*models.Users, error)
		Create(ctx context.Context, user models.Users) error
	}
)
