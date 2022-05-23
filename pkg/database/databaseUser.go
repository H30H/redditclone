package database

import (
	"database/sql"
	"fmt"
	"redditclone/pkg/user"

	_ "github.com/go-sql-driver/mysql"
)

type DatabaseUser struct {
	database *sql.DB
}

func InitDatabaseUser(path, databaseName string) (*DatabaseUser, error) {
	dsn := path + "/" + databaseName
	db, errOpen := sql.Open("mysql", dsn)
	if errOpen != nil {
		return nil, errOpen
	}
	db.SetMaxOpenConns(100)
	errConnect := db.Ping()
	if errConnect != nil {
		return nil, fmt.Errorf("ping: %w", errConnect)
	}
	return &DatabaseUser{database: db}, nil
}

func (d *DatabaseUser) Close() error {
	return d.database.Close()
}

func (d *DatabaseUser) Add(usr user.User) (id int64, err error) {
	result, err := d.database.Exec(
		"INSERT INTO users (`username`, `password`) VALUES (?, ?)",
		usr.Username,
		usr.PasswordHash,
	)
	if err == nil {
		id, err = result.LastInsertId()
	}
	return
}

func (d *DatabaseUser) Get(username string) (usr user.User, err error) {
	row := d.database.QueryRow("SELECT username, password, user_id FROM users WHERE username = ? LIMIT 1", username)
	err = row.Scan(&usr.Username, &usr.PasswordHash, &usr.UserID)
	return
}
