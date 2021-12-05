package main

type User struct {
	id              int64  `db:"id"`
	name            string `db:"name"`
	daysWithoutWeed int64  `db:"days_without_weed"`
	karma           int64  `db:"karma"`
}

type Storage interface {
	AddUser(id int64, name string) error
	UpdateUser(id int64, usedWeed bool) error
	User(id int64) (User, error)
	Users() ([]User, error)
}
