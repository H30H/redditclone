package middleware

import (
	"context"
	"net/http"
	"redditclone/pkg/database"
	"redditclone/pkg/frontendMessages"
	"redditclone/pkg/session"
	"redditclone/pkg/token"

	"strings"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Key int

const (
	UserContextKey Key = iota
	AuthtorizationContextKey
	GorrilaMuxVars
)

type Middleware struct {
	Users         *database.UserRepoStruct
	Authorization *session.SessionManagerStruct
	Logger        *zap.SugaredLogger
	Secretkey     string
}

func (m Middleware) AddAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, AuthtorizationContextKey, m.Authorization)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (m Middleware) CheckAuth(next http.HandlerFunc) http.HandlerFunc {
	badAuthorization := func(message string, w http.ResponseWriter) {
		frontendMessages.SendMessage(w,
			message, http.StatusUnauthorized,
			m.Logger, "middleware",
		)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errAuth := m.Authorization.CheckAuth(r)
		if errAuth != nil {
			switch errAuth.(type) {
			case session.ErrorTokenIsExpired:
				badAuthorization("expired authorization", w)

			case session.ErrorTokenNotFound:
				badAuthorization("leave and login again", w)
			default:
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}
		tokenArr := strings.Split(r.Header["Authorization"][0], " ")
		if len(tokenArr) != 2 || tokenArr[0] != "Bearer" {
			badAuthorization("bad authorization", w)
			m.Logger.Infof("middleware: bad handler: %s", r.Header["Authorization"][0])
			return
		}
		tokenStr := tokenArr[1]
		usr, err := token.CheckToken(tokenStr, m.Secretkey)
		if err != nil {
			badAuthorization("bad access token", w)
			m.Logger.Debugf("middleware: bad access token: %s", tokenStr)
			return
		}
		err = m.Authorization.UpdateAuth(w, r)
		if err != nil {
			m.Logger.Errorf("middleware: can`t update auth token: %w", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := recover(); err != nil {
				m.Logger.Errorf("middleware: have panic: %s, url: %s", err, r.URL.Path)
				http.Error(w, "Internal server error", 500)
			}
		}()
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserContextKey, usr)
		ctx = context.WithValue(ctx, GorrilaMuxVars, mux.Vars(r))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
