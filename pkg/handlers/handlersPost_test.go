package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"redditclone/pkg/database/mocks"
	"redditclone/pkg/frontendMessages"
	"redditclone/pkg/middleware"
	"redditclone/pkg/post"
	"redditclone/pkg/user"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPostAdd(t *testing.T) {
	type testCase struct {
		keyUser   middleware.Key
		valueUser user.User

		postRepoAddError error

		request         string
		statusCode      int
		responseHasPost bool
		responsePost    post.Post
		responseMessage string
	}

	testCases := []testCase{
		{
			keyUser:          middleware.UserContextKey,
			valueUser:        user.User{Username: "test1", UserID: 0, PasswordHash: "test1"},
			postRepoAddError: nil,
			request:          "{}",
			statusCode:       http.StatusOK,
			responseHasPost:  true,
			responsePost: post.Post{
				Title:            "",
				Author:           user.User{Username: "test1", UserID: 0},
				Category:         "",
				ID:               0,
				Score:            1,
				Time:             "",
				Views:            0,
				UpvotePercentage: 100,
				Type:             "",
				Votes: []frontendMessages.Vote{
					{
						UserID: 0,
						Vote:   1,
					},
				},
				Comments: nil,
			},
		},
		{
			keyUser:          middleware.AuthtorizationContextKey,
			valueUser:        user.User{Username: "test1", UserID: 0, PasswordHash: "test1"},
			postRepoAddError: nil,
			request:          "{}",
			statusCode:       http.StatusInternalServerError,
			responseHasPost:  false,
			responseMessage:  "Internal server error\n",
		},
		{
			keyUser:          middleware.UserContextKey,
			valueUser:        user.User{Username: "test1", UserID: 0, PasswordHash: "test1"},
			postRepoAddError: nil,
			request:          "wrong json",
			statusCode:       http.StatusInternalServerError,
			responseHasPost:  false,
			responseMessage:  "",
		},
		{
			keyUser:          middleware.UserContextKey,
			valueUser:        user.User{Username: "test1", UserID: 0, PasswordHash: "test1"},
			postRepoAddError: fmt.Errorf("test error"),
			request:          "{}",
			statusCode:       http.StatusInternalServerError,
			responseHasPost:  false,
			responseMessage:  "Internal server error\n",
		},
	}

	for _, testCase := range testCases {
		postHandler := setupPost()
		defer postHandler.Logger.Sync()

		r := httptest.NewRequest("GET", "/api/posts/", strings.NewReader(testCase.request))
		ctx := r.Context()
		ctx = context.WithValue(ctx, testCase.keyUser, testCase.valueUser)
		w := httptest.NewRecorder()

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Add", mock.AnythingOfType("*post.Post")).
			Return(testCase.postRepoAddError)

		postHandler.PostAdd(w, r.WithContext(ctx))

		resp := w.Result()

		body, errRead := ioutil.ReadAll(resp.Body)
		require.NoError(t, errRead)
		require.Equal(t, resp.StatusCode, testCase.statusCode)

		if testCase.responseHasPost {
			var post post.Post
			errUnmarshal := json.Unmarshal(body, &post)
			require.NoError(t, errUnmarshal)
			post.Time = ""
			require.Equal(t, post, testCase.responsePost)
		} else {
			require.Equal(t, string(body), testCase.responseMessage)
		}
	}
}

func TestPostGet(t *testing.T) {
	type testCase struct {
		valueVars map[string]string
		postID    uint64

		postRepoLockStatus   bool
		postRepoGetPost      post.Post
		postRepoGetError     error
		postRepoUpdateError  error
		postRepoUnlockStatus bool

		statusCode      int
		responseHasPost bool
		responsePost    post.Post
		responseMessage string
	}

	testCases := []testCase{
		{
			valueVars:            map[string]string{"post_id": "0"},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGetPost:      post.Post{},
			postRepoGetError:     nil,
			postRepoUpdateError:  nil,
			postRepoUnlockStatus: true,

			statusCode:      http.StatusOK,
			responseHasPost: true,
			responsePost: post.Post{
				Views: 1,
			},
		},
		{
			valueVars:            map[string]string{"test error": "0"},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGetPost:      post.Post{},
			postRepoGetError:     nil,
			postRepoUpdateError:  nil,
			postRepoUnlockStatus: true,

			statusCode:      http.StatusInternalServerError,
			responseHasPost: false,
			responseMessage: "",
		},
		{
			valueVars:            map[string]string{"post_id": "0"},
			postID:               0,
			postRepoLockStatus:   false,
			postRepoGetPost:      post.Post{},
			postRepoGetError:     nil,
			postRepoUpdateError:  nil,
			postRepoUnlockStatus: true,

			statusCode:      http.StatusNotFound,
			responseHasPost: false,
			responseMessage: "{\"message\":\"post not found\"}\n",
		},
		{
			valueVars:            map[string]string{"post_id": "0"},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGetPost:      post.Post{},
			postRepoGetError:     fmt.Errorf("test error"),
			postRepoUpdateError:  nil,
			postRepoUnlockStatus: true,

			statusCode:      http.StatusInternalServerError,
			responseHasPost: false,
			responseMessage: "",
		},
		{
			valueVars:            map[string]string{"post_id": "0"},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGetPost:      post.Post{},
			postRepoGetError:     nil,
			postRepoUpdateError:  fmt.Errorf("test error"),
			postRepoUnlockStatus: true,

			statusCode:      http.StatusInternalServerError,
			responseHasPost: false,
			responseMessage: "",
		},
		{
			valueVars:            map[string]string{"post_id": "0"},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGetPost:      post.Post{},
			postRepoGetError:     nil,
			postRepoUpdateError:  nil,
			postRepoUnlockStatus: false,

			statusCode:      http.StatusInternalServerError,
			responseHasPost: false,
			responseMessage: "",
		},
	}

	for _, testCase := range testCases {
		postHandler := setupPost()
		defer postHandler.Logger.Sync()

		r := httptest.NewRequest("GET", "/api/post/"+fmt.Sprint(testCase.postID), nil)
		r = mux.SetURLVars(r, testCase.valueVars)
		w := httptest.NewRecorder()

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Lock", testCase.postID).
			Return(testCase.postRepoLockStatus)

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Get", testCase.postID).
			Return(testCase.postRepoGetPost, testCase.postRepoGetError)

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Update", mock.AnythingOfType("post.Post")).
			Return(testCase.postRepoUpdateError)

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Unlock", testCase.postID).
			Return(testCase.postRepoUnlockStatus)

		postHandler.PostGet(w, r)

		resp := w.Result()

		body, errRead := ioutil.ReadAll(resp.Body)
		require.NoError(t, errRead)
		require.Equal(t, resp.StatusCode, testCase.statusCode)

		if testCase.responseHasPost {
			var post post.Post
			errUnmarshal := json.Unmarshal(body, &post)
			require.NoError(t, errUnmarshal)
			post.Time = ""
			require.Equal(t, post, testCase.responsePost)
		} else {
			require.Equal(t, string(body), testCase.responseMessage)
		}
	}
}

func TestPosts(t *testing.T) {
	type testCase struct {
		postRepoToJsonRes   []byte
		postRepoToJsonError error

		statusCode int
		response   string
	}

	testCases := []testCase{
		{
			postRepoToJsonRes:   []byte("test result"),
			postRepoToJsonError: nil,
			statusCode:          http.StatusOK,
			response:            "test result",
		},
		{
			postRepoToJsonRes:   nil,
			postRepoToJsonError: fmt.Errorf("test error"),
			statusCode:          http.StatusInternalServerError,
			response:            "",
		},
	}

	for _, testCase := range testCases {
		postHandler := setupPost()
		defer postHandler.Logger.Sync()

		r := httptest.NewRequest("GET", "/api/posts/", nil)
		w := httptest.NewRecorder()

		postHandler.PostRepo.(*mocks.PostRepo).
			On("ToJson", "", "").
			Return(testCase.postRepoToJsonRes, testCase.postRepoToJsonError)

		postHandler.Posts(w, r)

		resp := w.Result()

		body, errRead := ioutil.ReadAll(resp.Body)
		require.NoError(t, errRead)
		require.Equal(t, resp.StatusCode, testCase.statusCode)
		require.Equal(t, string(body), testCase.response)

	}
}

func TestCategories(t *testing.T) {
	type testCase struct {
		valueVars           map[string]string
		category            string
		postRepoToJsonRes   []byte
		postRepoToJsonError error

		statusCode int
		response   string
	}

	testCases := []testCase{
		{
			valueVars:           map[string]string{"category_name": "test category"},
			category:            "test category",
			postRepoToJsonRes:   []byte("some result"),
			postRepoToJsonError: nil,
			statusCode:          http.StatusOK,
			response:            "some result",
		},
		{
			valueVars:           map[string]string{"wrong_category": "test category"},
			category:            "",
			postRepoToJsonRes:   []byte("some result"),
			postRepoToJsonError: nil,
			statusCode:          http.StatusInternalServerError,
			response:            "",
		},
		{
			valueVars:           map[string]string{"category_name": "test category"},
			category:            "test category",
			postRepoToJsonRes:   nil,
			postRepoToJsonError: fmt.Errorf("test error"),
			statusCode:          http.StatusInternalServerError,
			response:            "",
		},
	}

	for _, testCase := range testCases {
		postHandler := setupPost()
		defer postHandler.Logger.Sync()

		r := httptest.NewRequest("GET", "/api/posts/", nil)
		r = mux.SetURLVars(r, testCase.valueVars)
		w := httptest.NewRecorder()

		postHandler.PostRepo.(*mocks.PostRepo).
			On("ToJson", testCase.category, "").
			Return(testCase.postRepoToJsonRes, testCase.postRepoToJsonError)

		postHandler.Categories(w, r)

		resp := w.Result()

		body, errRead := ioutil.ReadAll(resp.Body)
		require.NoError(t, errRead)
		require.Equal(t, resp.StatusCode, testCase.statusCode)
		require.Equal(t, string(body), testCase.response)
	}
}

func TestUserPosts(t *testing.T) {
	type testCase struct {
		valueVars map[string]string
		username  string

		postRepoToJsonRes   []byte
		postRepoToJsonError error

		statusCode int
		response   string
	}

	testCases := []testCase{
		{
			valueVars:           map[string]string{"user_login": "testuser"},
			username:            "testuser",
			postRepoToJsonRes:   []byte("test result"),
			postRepoToJsonError: nil,
			statusCode:          http.StatusOK,
			response:            "test result",
		},
		{
			valueVars:           map[string]string{"wrong vars": "testuser"},
			username:            "testuser",
			postRepoToJsonRes:   []byte("test result"),
			postRepoToJsonError: nil,
			statusCode:          http.StatusInternalServerError,
			response:            "",
		},
		{
			valueVars:           map[string]string{"user_login": "testuser"},
			username:            "testuser",
			postRepoToJsonRes:   nil,
			postRepoToJsonError: fmt.Errorf("test error"),
			statusCode:          http.StatusInternalServerError,
			response:            "",
		},
	}

	for _, testCase := range testCases {
		postHandler := setupPost()
		defer postHandler.Logger.Sync()

		r := httptest.NewRequest("GET", "/api/posts/", nil)
		r = mux.SetURLVars(r, testCase.valueVars)
		w := httptest.NewRecorder()

		postHandler.PostRepo.(*mocks.PostRepo).
			On("ToJson", "", testCase.username).
			Return(testCase.postRepoToJsonRes, testCase.postRepoToJsonError)

		postHandler.UserPosts(w, r)

		resp := w.Result()

		body, errRead := ioutil.ReadAll(resp.Body)
		require.NoError(t, errRead)
		require.Equal(t, resp.StatusCode, testCase.statusCode)
		require.Equal(t, string(body), testCase.response)
	}
}

func TestPostRemove(t *testing.T) {
	type testCase struct {
		valueVars    map[string]string
		contextKey   middleware.Key
		contextValue interface{}

		postID uint64

		postRepoLockStatus   bool
		postRepoGetPost      post.Post
		postRepoGetError     error
		postRepoRemoveStatus bool

		statusCode int
		response   string
	}

	testCases := []testCase{
		{
			valueVars:            map[string]string{"post_id": "0"},
			contextKey:           middleware.UserContextKey,
			contextValue:         user.User{Username: "test", UserID: 0},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGetPost:      post.Post{Author: user.User{Username: "test", UserID: 0}},
			postRepoGetError:     nil,
			postRepoRemoveStatus: true,
			statusCode:           http.StatusOK,
			response:             "{\"message\":\"success\"}\n",
		},
		{
			valueVars:            map[string]string{"wrong vars": "0"},
			contextKey:           middleware.UserContextKey,
			contextValue:         user.User{Username: "test", UserID: 0},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGetPost:      post.Post{Author: user.User{Username: "test", UserID: 0}},
			postRepoGetError:     nil,
			postRepoRemoveStatus: true,
			statusCode:           http.StatusInternalServerError,
			response:             "",
		},
		{
			valueVars:            map[string]string{"post_id": "0"},
			contextKey:           middleware.AuthtorizationContextKey,
			contextValue:         user.User{Username: "test", UserID: 0},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGetPost:      post.Post{Author: user.User{Username: "test", UserID: 0}},
			postRepoGetError:     nil,
			postRepoRemoveStatus: true,
			statusCode:           http.StatusInternalServerError,
			response:             "Internal server error\n",
		},
		{
			valueVars:            map[string]string{"post_id": "0"},
			contextKey:           middleware.UserContextKey,
			contextValue:         user.User{Username: "test", UserID: 0},
			postID:               0,
			postRepoLockStatus:   false,
			postRepoGetPost:      post.Post{Author: user.User{Username: "test", UserID: 0}},
			postRepoGetError:     nil,
			postRepoRemoveStatus: true,
			statusCode:           http.StatusNotFound,
			response:             "{\"message\":\"post not found\"}\n",
		},
		{
			valueVars:            map[string]string{"post_id": "0"},
			contextKey:           middleware.UserContextKey,
			contextValue:         user.User{Username: "test", UserID: 0},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGetPost:      post.Post{},
			postRepoGetError:     fmt.Errorf("test error"),
			postRepoRemoveStatus: true,
			statusCode:           http.StatusInternalServerError,
			response:             "",
		},
		{
			valueVars:            map[string]string{"post_id": "0"},
			contextKey:           middleware.UserContextKey,
			contextValue:         user.User{Username: "test2", UserID: 0},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGetPost:      post.Post{Author: user.User{Username: "test", UserID: 0}},
			postRepoGetError:     nil,
			postRepoRemoveStatus: true,
			statusCode:           http.StatusNotFound,
			response:             "{\"message\":\"this post doesn't belong to this user\"}\n",
		},
		{
			valueVars:            map[string]string{"post_id": "0"},
			contextKey:           middleware.UserContextKey,
			contextValue:         user.User{Username: "test", UserID: 0},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGetPost:      post.Post{Author: user.User{Username: "test", UserID: 0}},
			postRepoGetError:     nil,
			postRepoRemoveStatus: false,
			statusCode:           http.StatusInternalServerError,
			response:             "",
		},
	}

	for _, testCase := range testCases {
		postHandler := setupPost()
		defer postHandler.Logger.Sync()

		r := httptest.NewRequest("GET", "/api/posts/", nil)
		r = mux.SetURLVars(r, testCase.valueVars)
		ctx := r.Context()
		ctx = context.WithValue(ctx, testCase.contextKey, testCase.contextValue)

		w := httptest.NewRecorder()

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Lock", testCase.postID).
			Return(testCase.postRepoLockStatus)

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Get", testCase.postID).
			Return(testCase.postRepoGetPost, testCase.postRepoGetError)

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Remove", testCase.postID).
			Return(testCase.postRepoRemoveStatus)

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Unlock", testCase.postID).
			Return(true)

		postHandler.PostRemove(w, r.WithContext(ctx))

		resp := w.Result()

		body, errRead := ioutil.ReadAll(resp.Body)
		require.NoError(t, errRead)
		require.Equal(t, resp.StatusCode, testCase.statusCode)
		require.Equal(t, string(body), testCase.response)

	}
}
