package session

import (
	"fmt"
	"log"
	"net/http"
	"redditclone/pkg/token"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

const (
	AuthTime = time.Hour * 24 * 7
	tokenLen = 200
)

type SessionManager interface {
	CheckAllTimes(logger *zap.SugaredLogger) error
	AddAuth(w http.ResponseWriter) error
	CheckAuth(r *http.Request) error
	UpdateAuth(w http.ResponseWriter, r *http.Request) error
	Close() error
}

type SessionManagerStruct struct {
	database DatabaseSession
}

func InitSessionManager(database DatabaseSession) *SessionManagerStruct {
	return &SessionManagerStruct{database: database}
}

func (s *SessionManagerStruct) CheckAllTimes(logger *zap.SugaredLogger) error {
	rows, err := s.database.GetAll()
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		if logger != nil {
			logger.Debugf("demon: no auth users")
		}
		return nil
	}
	for i, row := range rows {
		if row.timeTo < time.Now().Unix() {
			err = s.database.RemoveToken(row.token)
			if err != nil {
				return fmt.Errorf("databaseSessionDeamon: %w", err)
			}
			if logger != nil {
				logger.Debugf("demon: remove token: %s", row.token)
			}
		} else {
			logger.Debugf("demon: %d: token: %s", i, row.token)
		}
	}
	return nil
}

func (s *SessionManagerStruct) AddAuth(w http.ResponseWriter) error {
	authToken := token.RandStringRunes(tokenLen)

	err := s.database.AddToken(authToken, time.Now().Add(AuthTime).Unix())
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   authToken,
		Expires: time.Now().Add(AuthTime),
		Path:    "/",
	}
	http.SetCookie(w, cookie)
	return err
}

func (s *SessionManagerStruct) CheckAuth(r *http.Request) error {
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		return ErrorTokenNotFound{}
	}
	token := sessionCookie.Value
	log.Printf("token: %s", token)
	timeToken, err := s.database.GetTime(token)
	log.Printf("err: %s", err)
	log.Printf("timeTo: %d", timeToken)
	if err != nil {
		return ErrorTokenNotFound{}
	}
	if timeToken < time.Now().Unix() {

		if err = s.database.RemoveToken(sessionCookie.Value); err != nil {
			return err
		}
		return ErrorTokenIsExpired{}
	}
	return nil
}

func (s *SessionManagerStruct) UpdateAuth(w http.ResponseWriter, r *http.Request) error {
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		return ErrorTokenNotFound{}
	}

	token := sessionCookie.Value

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   token,
		Expires: time.Now().Add(AuthTime),
		Path:    "/",
	}
	http.SetCookie(w, cookie)

	return s.database.UpdateAuth(token, time.Now().Add(AuthTime))
}

func (s *SessionManagerStruct) Close() error {
	return s.database.Close()
}
