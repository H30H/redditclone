package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"redditclone/pkg/errors"
	"redditclone/pkg/frontendMessages"
	"redditclone/pkg/middleware"
	"redditclone/pkg/session"
	"redditclone/pkg/token"
	"redditclone/pkg/user"
)

type userJson struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type sendToken struct {
	Token string `json:"token"`
}

func (u userJson) ToUser() user.User {
	return user.User{
		Username:     u.Username,
		PasswordHash: user.GetPasswordHash(u.Password),
		UserID:       0,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	auth, ok := r.Context().Value(middleware.AuthtorizationContextKey).(session.SessionManager)
	if !ok {
		h.Logger.Errorf("no context value: %s", middleware.AuthtorizationContextKey)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	body, _ := ioutil.ReadAll(r.Body)
	usrJson := userJson{}
	errUnmarshal := json.Unmarshal(body, &usrJson)
	if errUnmarshal != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("registration: %w", errors.ErrUnmarshalRequest{Err: errUnmarshal}),
		)
		return
	}
	usr := usrJson.ToUser()
	if _, ok := h.UserRepo.Find(usr.Username); ok {
		res, errMarshal := json.Marshal(frontendMessages.Error{Errors: []frontendMessages.ErrorMessage{{
			Location: "body",
			Param:    "username",
			Value:    usr.Username,
			Message:  "already exists",
		}}})
		if errMarshal != nil {
			errors.SendHttpError(
				h.Logger, w,
				fmt.Errorf("registration: %w", errors.ErrMarshal{Err: errMarshal}),
			)
			return
		}
		http.Error(w, string(res), http.StatusUnprocessableEntity)
		return
	}
	errAdd := h.UserRepo.Add(&usr)
	if errAdd != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("registration: can`t add user to database: %w", errAdd),
		)
		return
	}
	tokenStr, errToken := token.GetToken(usr, h.SecretKey)
	if errToken != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("registration: %w", errToken),
		)
		return
	}
	errAuth := auth.AddAuth(w)
	if errAuth != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("registration: can`t add authorization to database: %w, token: %s", errAuth, tokenStr),
		)
		return
	}
	send, errMarshal := json.Marshal(sendToken{Token: tokenStr})
	if errMarshal != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("registration: %w", errors.ErrMarshal{Err: errMarshal}),
		)
		return
	}
	h.Logger.Infof(`registered user: "%s", userID: "%d"`, usr.Username, usr.UserID)
	w.WriteHeader(http.StatusCreated)
	w.Write(send)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	auth, ok := r.Context().Value(middleware.AuthtorizationContextKey).(session.SessionManager)
	if !ok {
		h.Logger.Errorf("no context value: %s", middleware.AuthtorizationContextKey)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	body, _ := ioutil.ReadAll(r.Body)
	usrJson := userJson{}
	errUnmarshal := json.Unmarshal(body, &usrJson)
	if errUnmarshal != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("login: %w", errors.ErrUnmarshalRequest{Err: errUnmarshal}),
		)
		return
	}
	userGet, ok := h.UserRepo.Find(usrJson.Username)
	if !ok || userGet.PasswordHash != usrJson.ToUser().PasswordHash {
		var mesg string
		if !ok {
			mesg = "user not found"
		} else {
			mesg = "invalid password"
			h.Logger.Debugf("bad login: has password %s, but needed: %s", userGet.PasswordHash, usrJson.ToUser().PasswordHash)
		}
		res, errMarshal := json.Marshal(frontendMessages.Message{Message: mesg})
		if errMarshal != nil {
			errors.SendHttpError(
				h.Logger, w,
				fmt.Errorf("login: %w", errors.ErrMarshal{Err: errMarshal}),
			)
			return
		}
		http.Error(w, string(res), http.StatusUnauthorized)
		return
	}
	tokenStr, errToken := token.GetToken(userGet, h.SecretKey)
	if errToken != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("login: %w", errToken),
		)
		return
	}
	errAuth := auth.AddAuth(w)
	if errAuth != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("login: can`t add authorization to database: %w", errAuth),
		)
		return
	}
	send, errMarshal := json.Marshal(sendToken{Token: tokenStr})
	if errMarshal != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("login: %w", errors.ErrMarshal{Err: errMarshal}),
		)
		return
	}
	h.Logger.Infof(`login user: "%s", userID: "%d"`, userGet.Username, userGet.UserID)
	w.WriteHeader(http.StatusCreated)
	w.Write(send)
}
