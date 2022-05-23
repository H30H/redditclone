package session

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type DatabaseRow struct {
	token  string
	timeTo int64
}

type DatabaseSession interface {
	AddToken(token string, time int64) error
	GetTime(token string) (time int64, err error)
	GetAll() (res []*DatabaseRow, err error)
	RemoveToken(token string) error
	UpdateAuth(token string, timeTo time.Time) error
	Close() error
}

type DatabaseSessionStruct struct {
	database *sql.DB
}

func (d *DatabaseSessionStruct) GetTime(token string) (time int64, err error) {
	row := d.database.QueryRow("SELECT time_to FROM authorization WHERE token = ? LIMIT 1", token)
	err = row.Scan(&time)
	return
}

func (d *DatabaseSessionStruct) GetAll() (res []*DatabaseRow, err error) {
	rows, err := d.database.Query("select * from authorization")
	if err != nil {
		return nil, fmt.Errorf("databaseSessionStruct: GetAll: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var token string
		var timeTo int64
		err := rows.Scan(&token, &timeTo)
		if err != nil {
			return nil, fmt.Errorf("databaseSessionStruct: GetAll: %w", err)
		}
		res = append(res, &DatabaseRow{token: token, timeTo: timeTo})
	}
	return
}

func (d *DatabaseSessionStruct) AddToken(token string, time int64) error {
	_, err := d.database.Exec("INSERT INTO authorization (`token`, `time_to`) VALUES (?, ?)",
		token, time)
	return err
}

func (d *DatabaseSessionStruct) RemoveToken(token string) error {
	_, err := d.database.Exec("DELETE FROM authorization WHERE token = ?",
		token)
	return err
}

func (d *DatabaseSessionStruct) UpdateAuth(token string, timeTo time.Time) error {
	_, err := d.database.Exec("UPDATE authorization SET time_to = ? WHERE token = ?",
		timeTo.Unix(), token)
	return err
}

func (d *DatabaseSessionStruct) Close() error {
	return d.database.Close()
}

func InitDatabaseSession(path, databaseName string) (*DatabaseSessionStruct, error) {
	dsn := path + "/" + databaseName
	log.Printf("database session: %s", dsn)
	db, errOpen := sql.Open("mysql", dsn)
	if errOpen != nil {
		return nil, errOpen
	}
	db.SetMaxOpenConns(100)
	errConnect := db.Ping()
	if errConnect != nil {
		return nil, fmt.Errorf("ping: %w", errConnect)
	}
	return &DatabaseSessionStruct{database: db}, nil
}
