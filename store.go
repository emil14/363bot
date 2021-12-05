package main

import (
	"fmt"
)

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

type memoryStorage struct {
	uu []User
}

func (m *memoryStorage) AddUser(id int64, name string) error {
	m.uu = append(m.uu, User{
		id:   id,
		name: name,
	})

	return nil
}

func (m *memoryStorage) UpdateUser(id int64, smokedYesteday bool) error {
	for i, u := range m.uu {
		if u.id != id {
			continue
		}

		if smokedYesteday {
			if u.daysWithoutWeed > 0 {
				u.daysWithoutWeed = 0
			} else {
				u.daysWithoutWeed -= 1
			}
		} else {
			if u.daysWithoutWeed < 0 {
				u.daysWithoutWeed = 1
			} else {
				u.daysWithoutWeed += 1
			}
		}

		m.uu[i] = u
	}

	return nil
}

func (m *memoryStorage) User(id int64) (User, error) {
	for _, u := range m.uu {
		if u.id == id {
			return u, nil
		}
	}

	return User{}, fmt.Errorf("...")
}

func (m *memoryStorage) Users() ([]User, error) {
	return m.uu, nil
}

func NewStorage() memoryStorage {
	return memoryStorage{
		uu: []User{},
	}
}
