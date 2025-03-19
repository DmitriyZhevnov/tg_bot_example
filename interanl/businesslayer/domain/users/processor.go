package users

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"testbot/interanl/businesslayer/dto"
	"testbot/interanl/datalayer/collections"
	"testbot/interanl/datalayer/models"
)

type Processor struct {
	logger          zerolog.Logger
	usersCollection collections.Users
}

func NewProcessor(logger zerolog.Logger, usersCollection collections.Users) *Processor {
	return &Processor{logger: logger, usersCollection: usersCollection}
}

func (p *Processor) CreateIfNotExist(ctx context.Context, userRequest dto.User) error {
	user, err := p.usersCollection.GetByChatID(ctx, userRequest.ChatID)
	if err != nil {
		p.logger.Err(err).Send()
		return err
	}

	if user == nil {
		if err := p.usersCollection.Create(
			ctx,
			models.Users{
				ID:     uuid.New(),
				Name:   userRequest.Name,
				ChatID: userRequest.ChatID,
			},
		); err != nil {
			p.logger.Err(err).Send()
			return err
		}
	}

	return nil
}

func (p *Processor) LoadByChatID(ctx context.Context, chatID int64) (*dto.User, error) {
	user, err := p.usersCollection.GetByChatID(ctx, chatID)
	if err != nil {
		p.logger.Err(err).Send()
		return nil, err
	}

	if user == nil {
		return nil, errors.New("not found")
	}

	return &dto.User{
		ID:     user.ID,
		Name:   user.Name,
		ChatID: user.ChatID,
	}, nil
}

func (p *Processor) Work(userName string) string {
	return fmt.Sprintf("Дорогой %s, я работаю над этим", userName)
}
