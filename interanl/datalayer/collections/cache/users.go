package cache

import (
	"context"
	"fmt"
	"testbot/interanl/datalayer/models"
)

type UsersCollection struct {
	users map[int64]models.Users
}

func NewUsersCollection() *UsersCollection {
	return &UsersCollection{
		users: make(map[int64]models.Users),
	}
}

func (c *UsersCollection) GetByChatID(ctx context.Context, chatID int64) (*models.Users, error) {
	user, ok := c.users[chatID]
	if ok {
		return &user, nil
	}

	return nil, nil
}

func (c *UsersCollection) Create(ctx context.Context, userRequest models.Users) error {
	_, ok := c.users[userRequest.ChatID]
	if ok {
		return nil
	}

	c.users[userRequest.ChatID] = userRequest

	fmt.Println("user created: ", userRequest)

	return nil
}
