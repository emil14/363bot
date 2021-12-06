package main

import (
	"context"
	"fmt"
	"math"
	"os"

	pgx "github.com/jackc/pgx/v4"
)

var m = map[string][]int{
	"3d":  {2},
	"4d":  {3},       // 7
	"7d":  {4},       // 14
	"9d":  {5},       // 23
	"10d": {6},       // 33
	"14d": {7, 8, 9}, // 47 61 75
	"15d": {10},      // 90
}

type karma struct {
	days   uint8
	factor uint8
}

var karmaPolice = []karma{
	{3, 2},
	{7, 3},
	{14, 4},
	{23, 5},
	{33, 6},
	{47, 7},
	{61, 8},
	{75, 9},
	{90, 10},
}

func getFactor(days int64) uint8 {
	var f uint8 = 1

	for _, k := range karmaPolice {
		if int64(math.Abs(float64(days))) >= int64(k.days) {
			f = k.factor
		}
		break
	}

	return f
}

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
		SET days_without_weed = $1, karma = $2
		WHERE id=$3;
	`

	if smokedYesteday {
		if u.daysWithoutWeed > 0 {
			u.daysWithoutWeed = 0
			u.karma -= int64(10 * getFactor(u.daysWithoutWeed))
		} else {
			u.daysWithoutWeed -= 1
			u.karma -= int64(10 * getFactor(u.daysWithoutWeed))
		}
	} else {
		if u.daysWithoutWeed < 0 {
			u.daysWithoutWeed = 1
			u.karma += int64(10 * getFactor(u.daysWithoutWeed))
		} else {
			u.daysWithoutWeed += 1
			u.karma += int64(10 * getFactor(u.daysWithoutWeed))
		}
	}

	_, err = m.conn.Exec(context.TODO(), query, u.daysWithoutWeed, u.karma, id)

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
