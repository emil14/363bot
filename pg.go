package main

import (
	"context"
	"fmt"
	"os"

	pgx "github.com/jackc/pgx/v4"
)

type postgresStorage struct {
	conn *pgx.Conn
}

func (m *postgresStorage) AddUser(id int64, name string) error {
	query := `
	INSERT INTO users (
		id,
		name,
		days_without_weed,
		karma
	)
	VALUES ($1,$2,$3,$4);`

	if _, err := m.conn.Exec(context.TODO(), query, id, name, 0, 0); err != nil {
		return err
	}

	return nil
}

func (m *postgresStorage) UpdateUser(id int64, smokedYesteday bool) error {
	u, err := m.User(id)
	if err != nil {
		return err
	}

	query := `
	UPDATE users
		SET days_without_weed = $1
		WHERE id=$2;
	`

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

	_, err = m.conn.Exec(context.TODO(), query, u.daysWithoutWeed, id)

	return err
}

func (m *postgresStorage) User(id int64) (User, error) {
	query := `
	SELECT id, name, days_without_weed, karma 
		FROM users WHERE id=$1;`

	var u User
	row := m.conn.QueryRow(context.TODO(), query, id)
	if err := row.Scan(&u.id, &u.name, &u.daysWithoutWeed, &u.karma); err != nil {
		return User{}, err
	}

	return u, nil
}

func (m *postgresStorage) Users() ([]User, error) {
	query := `
	SELECT id, name, days_without_weed, karma 
		FROM users;`

	rows, err := m.conn.Query(context.TODO(), query)
	if err != nil {
		return nil, err
	}

	uu := []User{}

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.id, &u.name, &u.daysWithoutWeed, &u.karma); err != nil {
			return nil, err
		}

		uu = append(uu, u)
	}

	return uu, nil
}

func NewPGStorage() postgresStorage {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	_, err = conn.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS users (
		id INT PRIMARY KEY,
		name VARCHAR(255),
		days_without_weed INT,
		karma INT
	);`)
	if err != nil {
		panic(err)
	}

	return postgresStorage{
		conn,
	}
}
