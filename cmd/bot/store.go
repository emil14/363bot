package main

import "context"

type User struct {
	id              int64  `db:"id"`
	name            string `db:"name"`
	daysWithoutWeed int64  `db:"days_without_weed"`
	karma           int64  `db:"karma"`
}

type Storage interface {
	AddUser(ctx context.Context, id int64, name string) error
	UpdateUser(ctx context.Context, id int64, usedWeed bool) error
	User(ctx context.Context, id int64) (User, error)
	Users(ctx context.Context) ([]User, error)

	Close(ctx context.Context) error
}
