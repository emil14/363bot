package main

import (
	"context"
	"math"

	pgx "github.com/jackc/pgx/v4"
)

type timeFactor struct {
	days   uint8
	factor uint8
}

var tf = [...]timeFactor{
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

	for _, k := range tf {
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

func (pg *postgresStorage) AddUser(ctx context.Context, id int64, name string) error {
	query := `
	INSERT INTO users (
		id,
		name,
		days_without_weed,
		karma
	)
	VALUES ($1,$2,$3,$4);`

	if _, err := pg.conn.Exec(ctx, query, id, name, 0, 0); err != nil {
		return err
	}

	return nil
}

func (pg *postgresStorage) UpdateUser(ctx context.Context, id int64, smokedYesteday bool) error {
	u, err := pg.User(ctx, id)
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
			u.karma /= int64(getFactor(u.daysWithoutWeed))
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

	_, err = pg.conn.Exec(ctx, query, u.daysWithoutWeed, u.karma, id)

	return err
}

func (pg *postgresStorage) User(ctx context.Context, id int64) (User, error) {
	query := `
	SELECT id, name, days_without_weed, karma 
		FROM users WHERE id=$1;`

	var u User
	row := pg.conn.QueryRow(ctx, query, id)
	if err := row.Scan(&u.id, &u.name, &u.daysWithoutWeed, &u.karma); err != nil {
		return User{}, err
	}

	return u, nil
}

func (pg *postgresStorage) Users(ctx context.Context) ([]User, error) {
	rows, err := pg.conn.Query(ctx, `SELECT id, name, days_without_weed, karma FROM users;`)
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

func (pg *postgresStorage) Close(ctx context.Context) error {
	return pg.conn.Close(ctx)
}

func NewPostgres(pgConnStr string) (Storage, error) {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, pgConnStr)
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS users (
		id INT PRIMARY KEY,
		name VARCHAR(32),
		days_without_weed INT,
		karma INT
	);`)
	if err != nil {
		return nil, err
	}

	return &postgresStorage{conn}, nil
}

func MustNewPostgres(pgConnStr string) Storage {
	pg, err := NewPostgres(pgConnStr)
	if err != nil {
		panic(err)
	}
	return pg
}
