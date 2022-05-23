package handlers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"redditclone/pkg/database/mocks"
	"redditclone/pkg/frontendMessages"
	"redditclone/pkg/middleware"
	"redditclone/pkg/post"
	"redditclone/pkg/user"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVoices(t *testing.T) {
	type testCase struct {
		voicetype int

		keyUser   middleware.Key
		valueUser interface{}
		keyVars   middleware.Key
		valueVars map[string]string

		postID uint64

		postRepoLockStatus   bool
		postRepoGet          post.Post
		postRepoGetError     error
		postRepoUpdateError  error
		postRepoUnlockStatus bool

		statusCode int
		response   string
	}

	testCases := []testCase{
		{
			voicetype:            1,
			keyUser:              middleware.UserContextKey,
			valueUser:            user.User{Username: "test", UserID: 0},
			keyVars:              middleware.GorrilaMuxVars,
			valueVars:            map[string]string{"post_id": "0"},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGet:          post.Post{},
			postRepoGetError:     nil,
			postRepoUpdateError:  nil,
			postRepoUnlockStatus: true,
			statusCode:           http.StatusOK,
			response:             "{\"title\":\"\",\"author\":{\"username\":\"\",\"id\":\"0\"},\"category\":\"\",\"id\":\"0\",\"created\":\"\",\"score\":1,\"views\":0,\"upvotePercentage\":100,\"type\":\"\",\"votes\":[{\"user\":\"0\",\"vote\":1}],\"comments\":null}",
		},
		{
			voicetype:            0,
			keyUser:              middleware.AuthtorizationContextKey,
			valueUser:            user.User{Username: "test", UserID: 0},
			keyVars:              middleware.GorrilaMuxVars,
			valueVars:            map[string]string{"post_id": "0"},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGet:          post.Post{},
			postRepoGetError:     nil,
			postRepoUpdateError:  nil,
			postRepoUnlockStatus: true,
			statusCode:           http.StatusInternalServerError,
			response:             "Internal server error\n",
		},
		{
			voicetype:            1,
			keyUser:              middleware.UserContextKey,
			valueUser:            user.User{Username: "test", UserID: 0},
			keyVars:              middleware.GorrilaMuxVars,
			valueVars:            map[string]string{"post_id": "0"},
			postID:               0,
			postRepoLockStatus:   false,
			postRepoGet:          post.Post{},
			postRepoGetError:     nil,
			postRepoUpdateError:  nil,
			postRepoUnlockStatus: true,
			statusCode:           http.StatusNotFound,
			response:             "{\"message\":\"post not found\"}\n",
		},
		{
			voicetype:            0,
			keyUser:              middleware.UserContextKey,
			valueUser:            user.User{Username: "test", UserID: 0},
			keyVars:              middleware.GorrilaMuxVars,
			valueVars:            map[string]string{"post_id": "0"},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGet:          post.Post{},
			postRepoGetError:     fmt.Errorf("test error"),
			postRepoUpdateError:  nil,
			postRepoUnlockStatus: true,
			statusCode:           http.StatusInternalServerError,
			response:             "",
		},
		{
			voicetype:            -1,
			keyUser:              middleware.UserContextKey,
			valueUser:            user.User{Username: "test", UserID: 0},
			keyVars:              middleware.GorrilaMuxVars,
			valueVars:            map[string]string{"post_id": "0"},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGet:          post.Post{},
			postRepoGetError:     nil,
			postRepoUpdateError:  fmt.Errorf("test error"),
			postRepoUnlockStatus: true,
			statusCode:           http.StatusInternalServerError,
			response:             "",
		},
		{
			voicetype:            1,
			keyUser:              middleware.UserContextKey,
			valueUser:            user.User{Username: "test", UserID: 0},
			keyVars:              middleware.GorrilaMuxVars,
			valueVars:            map[string]string{"post_id": "0"},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGet:          post.Post{},
			postRepoGetError:     nil,
			postRepoUpdateError:  nil,
			postRepoUnlockStatus: false,
			statusCode:           http.StatusInternalServerError,
			response:             "",
		},
		{
			voicetype:            0,
			keyUser:              middleware.UserContextKey,
			valueUser:            user.User{Username: "test", UserID: 0},
			keyVars:              middleware.GorrilaMuxVars,
			valueVars:            map[string]string{"post_id": "0"},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGet:          post.Post{Votes: []frontendMessages.Vote{{UserID: 0, Vote: 1}}},
			postRepoGetError:     nil,
			postRepoUpdateError:  nil,
			postRepoUnlockStatus: false,
			statusCode:           http.StatusInternalServerError,
			response:             "",
		},
		{
			voicetype:            -1,
			keyUser:              middleware.UserContextKey,
			valueUser:            user.User{Username: "test", UserID: 0},
			keyVars:              middleware.GorrilaMuxVars,
			valueVars:            map[string]string{"post_id": "0"},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGet:          post.Post{Votes: []frontendMessages.Vote{{UserID: 0, Vote: 1}}},
			postRepoGetError:     nil,
			postRepoUpdateError:  nil,
			postRepoUnlockStatus: false,
			statusCode:           http.StatusInternalServerError,
			response:             "",
		},
		{
			voicetype:            -1,
			keyUser:              middleware.UserContextKey,
			valueUser:            user.User{Username: "test", UserID: 0},
			keyVars:              middleware.GorrilaMuxVars,
			valueVars:            map[string]string{"test error": "0"},
			postID:               0,
			postRepoLockStatus:   true,
			postRepoGet:          post.Post{Votes: []frontendMessages.Vote{{UserID: 0, Vote: 1}}},
			postRepoGetError:     nil,
			postRepoUpdateError:  nil,
			postRepoUnlockStatus: false,
			statusCode:           http.StatusInternalServerError,
			response:             "",
		},
	}

	for _, testCase := range testCases {
		postHandler := setupPost()
		defer postHandler.Logger.Sync()

		var url string
		switch testCase.voicetype {
		case -1:
			url = "/api/post/" + fmt.Sprint(testCase.postID) + "/downvote"
		case 0:
			url = "/api/post/" + fmt.Sprint(testCase.postID) + "/unvote"
		case 1:
			url = "/api/post/" + fmt.Sprint(testCase.postID) + "/upvote"
		}

		r := httptest.NewRequest("GET", url, nil)
		r = mux.SetURLVars(r, testCase.valueVars)
		ctx := r.Context()
		ctx = context.WithValue(ctx, testCase.keyUser, testCase.valueUser)
		w := httptest.NewRecorder()

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Lock", testCase.postID).
			Return(testCase.postRepoLockStatus)

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Get", testCase.postID).
			Return(testCase.postRepoGet, testCase.postRepoGetError)

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Update", mock.AnythingOfType("post.Post")).
			Return(testCase.postRepoUpdateError)

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Unlock", testCase.postID).
			Return(testCase.postRepoUnlockStatus)

		switch testCase.voicetype {
		case -1:
			postHandler.PostRatingDown(w, r.WithContext(ctx))
		case 0:
			postHandler.PostRatingDefault(w, r.WithContext(ctx))
		case 1:
			postHandler.PostRatingUp(w, r.WithContext(ctx))
		}

		resp := w.Result()

		body, errRead := ioutil.ReadAll(resp.Body)

		require.NoError(t, errRead)
		require.Equal(t, resp.StatusCode, testCase.statusCode)
		require.Equal(t, string(body), testCase.response)
	}
}
