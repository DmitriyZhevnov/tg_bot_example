package businesslayer

import (
	"context"
	"testbot/interanl/businesslayer/dto"
)

type (
	ChatBot interface {
		SendMessage()
	}

	Users interface {
		CreateIfNotExist(ctx context.Context, user dto.User) error
		LoadByChatID(ctx context.Context, chatID int64) (*dto.User, error)
		Work(userName string) string
	}

	Executor interface {
		SaveValues(a, b int) error
	}
)
