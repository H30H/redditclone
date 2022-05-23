package handlers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"redditclone/pkg/database/mocks"
	"redditclone/pkg/middleware"
	"redditclone/pkg/session"
	sessionMocks "redditclone/pkg/session/mocks"
	"redditclone/pkg/user"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	type testCase struct {
		key   middleware.Key
		value interface{}

		repoFindStatus bool
		repoAddError   error

		sessionError error

		url     string
		request string

		response   string
		statusCode int
	}

	testCases := []testCase{
		{
			key:            middleware.AuthtorizationContextKey,
			value:          &session.SessionManagerStruct{},
			repoFindStatus: false,
			repoAddError:   nil,
			sessionError:   nil,
			url:            "/api/register",
			request:        `{"username":"test1","password":"testtest"}`,
			response:       "{\"token\":\"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjoiMCIsInVzZXJuYW1lIjoidGVzdDEifX0.oZ2FJZSdaRufCM1kjU41381gocjjzLcVDJqGu8m8Xyc\"}",
			statusCode:     http.StatusCreated,
		},
		{
			key:            middleware.UserContextKey,
			value:          &session.SessionManagerStruct{},
			repoFindStatus: false,
			repoAddError:   nil,
			sessionError:   nil,
			url:            "/api/register",
			request:        `{"username":"test1","password":"testtest"}`,
			response:       "Internal server error\n",
			statusCode:     http.StatusInternalServerError,
		},
		{
			key:            middleware.AuthtorizationContextKey,
			value:          &session.SessionManagerStruct{},
			repoFindStatus: false,
			repoAddError:   nil,
			sessionError:   nil,
			url:            "/api/register",
			request:        ``,
			response:       "",
			statusCode:     http.StatusInternalServerError,
		},
		{
			key:            middleware.AuthtorizationContextKey,
			value:          &session.SessionManagerStruct{},
			repoFindStatus: true,
			repoAddError:   nil,
			sessionError:   nil,
			url:            "/api/register",
			request:        `{"username":"test1","password":"testtest"}`,
			response:       "{\"errors\":[{\"location\":\"body\",\"param\":\"username\",\"value\":\"test1\",\"msg\":\"already exists\"}]}\n",
			statusCode:     http.StatusUnprocessableEntity,
		},
		{
			key:            middleware.AuthtorizationContextKey,
			value:          &session.SessionManagerStruct{},
			repoFindStatus: false,
			repoAddError:   fmt.Errorf("test error"),
			sessionError:   nil,
			url:            "/api/register",
			request:        `{"username":"test1","password":"testtest"}`,
			response:       "",
			statusCode:     http.StatusInternalServerError,
		},
		{
			key:            middleware.AuthtorizationContextKey,
			value:          &session.SessionManagerStruct{},
			repoFindStatus: false,
			repoAddError:   nil,
			sessionError:   fmt.Errorf("test error"),
			url:            "/api/register",
			request:        `{"username":"test1","password":"testtest"}`,
			response:       "",
			statusCode:     http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		userHandler := setupUser()
		defer userHandler.Logger.Sync()

		var authorization session.SessionManager = &sessionMocks.SessionManager{}
		r := httptest.NewRequest("GET", testCase.url, strings.NewReader(testCase.request))
		ctx := r.Context()
		ctx = context.WithValue(ctx, testCase.key, authorization)

		w := httptest.NewRecorder()

		userHandler.UserRepo.(*mocks.UserRepo).
			On("Find", mock.AnythingOfType("string")).
			Return(user.User{}, testCase.repoFindStatus)

		userHandler.UserRepo.(*mocks.UserRepo).
			On("Add", mock.AnythingOfType("*user.User")).
			Return(testCase.repoAddError)

		authorization.(*sessionMocks.SessionManager).
			On("AddAuth", w).
			Return(testCase.sessionError)

		userHandler.Register(w, r.WithContext(ctx))

		resp := w.Result()

		body, errRead := ioutil.ReadAll(resp.Body)

		require.NoError(t, errRead)
		require.Equal(t, resp.StatusCode, testCase.statusCode)
		require.Equal(t, string(body), testCase.response)
	}
}

func TestLogin(t *testing.T) {
	type testCase struct {
		key   middleware.Key
		value interface{}

		repoFindUser   user.User
		repoFindStatus bool

		sessionError error

		url     string
		request string

		response   string
		statusCode int
	}

	testCases := []testCase{
		{
			key:            middleware.AuthtorizationContextKey,
			value:          &session.SessionManagerStruct{},
			repoFindUser:   user.User{Username: "test1", PasswordHash: user.GetPasswordHash("test1"), UserID: 0},
			repoFindStatus: true,
			sessionError:   nil,
			url:            "/api/login",
			request:        `{"username":"test1","password":"test1"}`,
			response:       "{\"token\":\"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjoiMCIsInVzZXJuYW1lIjoidGVzdDEifX0.oZ2FJZSdaRufCM1kjU41381gocjjzLcVDJqGu8m8Xyc\"}",
			statusCode:     http.StatusCreated,
		},
		{
			key:            middleware.UserContextKey,
			value:          &session.SessionManagerStruct{},
			repoFindUser:   user.User{Username: "test1", PasswordHash: user.GetPasswordHash("test1"), UserID: 0},
			repoFindStatus: true,
			sessionError:   nil,
			url:            "/api/login",
			request:        `{"username":"test1","password":"test1"}`,
			response:       "Internal server error\n",
			statusCode:     http.StatusInternalServerError,
		},
		{
			key:            middleware.AuthtorizationContextKey,
			value:          &session.SessionManagerStruct{},
			repoFindUser:   user.User{Username: "test1", PasswordHash: user.GetPasswordHash("test1"), UserID: 0},
			repoFindStatus: true,
			sessionError:   nil,
			url:            "/api/login",
			request:        ``,
			response:       "",
			statusCode:     http.StatusInternalServerError,
		},
		{
			key:            middleware.AuthtorizationContextKey,
			value:          &session.SessionManagerStruct{},
			repoFindUser:   user.User{Username: "test1", PasswordHash: user.GetPasswordHash("test3"), UserID: 0},
			repoFindStatus: true,
			sessionError:   nil,
			url:            "/api/login",
			request:        `{"username":"test1","password":"test1"}`,
			response:       "{\"message\":\"invalid password\"}\n",
			statusCode:     http.StatusUnauthorized,
		},
		{
			key:            middleware.AuthtorizationContextKey,
			value:          &session.SessionManagerStruct{},
			repoFindUser:   user.User{},
			repoFindStatus: false,
			sessionError:   nil,
			url:            "/api/login",
			request:        `{"username":"test1","password":"test1"}`,
			response:       "{\"message\":\"user not found\"}\n",
			statusCode:     http.StatusUnauthorized,
		},
		{
			key:            middleware.AuthtorizationContextKey,
			value:          &session.SessionManagerStruct{},
			repoFindUser:   user.User{Username: "test1", PasswordHash: user.GetPasswordHash("test1"), UserID: 0},
			repoFindStatus: true,
			sessionError:   fmt.Errorf("test error"),
			url:            "/api/login",
			request:        `{"username":"test1","password":"test1"}`,
			response:       "",
			statusCode:     http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		userHandler := setupUser()
		defer userHandler.Logger.Sync()

		var authorization session.SessionManager = &sessionMocks.SessionManager{}
		r := httptest.NewRequest("GET", testCase.url, strings.NewReader(testCase.request))

		ctx := r.Context()
		ctx = context.WithValue(ctx, testCase.key, authorization)

		w := httptest.NewRecorder()

		userHandler.UserRepo.(*mocks.UserRepo).
			On("Find", mock.AnythingOfType("string")).
			Return(testCase.repoFindUser, testCase.repoFindStatus)

		authorization.(*sessionMocks.SessionManager).
			On("AddAuth", w).
			Return(testCase.sessionError)

		userHandler.Login(w, r.WithContext(ctx))

		resp := w.Result()

		body, errRead := ioutil.ReadAll(resp.Body)

		require.NoError(t, errRead)
		require.Equal(t, resp.StatusCode, testCase.statusCode, string(body))
		require.Equal(t, string(body), testCase.response)
	}
}
