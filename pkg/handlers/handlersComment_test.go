package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"redditclone/pkg/comment"
	"redditclone/pkg/database/mocks"
	"redditclone/pkg/middleware"
	"redditclone/pkg/post"
	"redditclone/pkg/user"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCommentAdd(t *testing.T) {
	type testCase struct {
		valueVars    map[string]string
		contextKey   middleware.Key
		contextValue interface{}
		postID       uint64

		postRepoLockStatus  bool
		postRepoGetPost     post.Post
		postRepoGetError    error
		postRepoUpdateError error

		request         string
		statusCode      int
		responseIsPost  bool
		responsePost    post.Post
		responseMessage string
	}

	testCases := []testCase{
		{
			valueVars:           map[string]string{"post_id": "0"},
			contextKey:          middleware.UserContextKey,
			contextValue:        user.User{Username: "test", UserID: 0},
			postID:              0,
			postRepoLockStatus:  true,
			postRepoGetPost:     post.Post{},
			postRepoGetError:    nil,
			postRepoUpdateError: nil,
			request:             `{"comment":"test comment"}`,
			statusCode:          http.StatusOK,
			responseIsPost:      true,
			responsePost: post.Post{
				Comments: []comment.Comment{{
					Author: user.User{Username: "test", UserID: 0},
					Body:   "test comment",
					Time:   "",
					ID:     0,
				}},
			},
		},
		{
			valueVars:           map[string]string{"wrong vars": "0"},
			contextKey:          middleware.UserContextKey,
			contextValue:        user.User{Username: "test", UserID: 0},
			postID:              0,
			postRepoLockStatus:  true,
			postRepoGetPost:     post.Post{},
			postRepoGetError:    nil,
			postRepoUpdateError: nil,
			request:             `{"comment":"test comment"}`,
			statusCode:          http.StatusInternalServerError,
			responseIsPost:      false,
			responseMessage:     "",
		},
		{
			valueVars:           map[string]string{"post_id": "0"},
			contextKey:          middleware.AuthtorizationContextKey,
			contextValue:        user.User{Username: "test", UserID: 0},
			postID:              0,
			postRepoLockStatus:  true,
			postRepoGetPost:     post.Post{},
			postRepoGetError:    nil,
			postRepoUpdateError: nil,
			request:             `{"comment":"test comment"}`,
			statusCode:          http.StatusInternalServerError,
			responseIsPost:      false,
			responseMessage:     "Internal server error\n",
		},
		{
			valueVars:           map[string]string{"post_id": "0"},
			contextKey:          middleware.UserContextKey,
			contextValue:        user.User{Username: "test", UserID: 0},
			postID:              0,
			postRepoLockStatus:  true,
			postRepoGetPost:     post.Post{},
			postRepoGetError:    nil,
			postRepoUpdateError: nil,
			request:             `wrong request`,
			statusCode:          http.StatusInternalServerError,
			responseIsPost:      false,
			responseMessage:     "",
		},
		{
			valueVars:           map[string]string{"post_id": "0"},
			contextKey:          middleware.UserContextKey,
			contextValue:        user.User{Username: "test", UserID: 0},
			postID:              0,
			postRepoLockStatus:  true,
			postRepoGetPost:     post.Post{},
			postRepoGetError:    nil,
			postRepoUpdateError: nil,
			request:             `{"comment":""}`,
			statusCode:          http.StatusUnprocessableEntity,
			responseIsPost:      false,
			responseMessage:     "{\"errors\":[{\"location\":\"body\",\"param\":\"comment\",\"msg\":\"is required\"}]}\n",
		},
		{
			valueVars:           map[string]string{"post_id": "0"},
			contextKey:          middleware.UserContextKey,
			contextValue:        user.User{Username: "test", UserID: 0},
			postID:              0,
			postRepoLockStatus:  true,
			postRepoGetPost:     post.Post{},
			postRepoGetError:    nil,
			postRepoUpdateError: nil,
			request: func() (res string) {
				res = `{"comment":"`
				for i := 0; i < 2100; i++ {
					res += "a"
				}
				return res + `"}`
			}(),
			statusCode:      http.StatusUnprocessableEntity,
			responseIsPost:  false,
			responseMessage: "{\"errors\":[{\"location\":\"body\",\"param\":\"comment\",\"msg\":\"must be at most 2000 characters long\"}]}\n",
		},
		{
			valueVars:           map[string]string{"post_id": "0"},
			contextKey:          middleware.UserContextKey,
			contextValue:        user.User{Username: "test", UserID: 0},
			postID:              0,
			postRepoLockStatus:  false,
			postRepoGetPost:     post.Post{},
			postRepoGetError:    nil,
			postRepoUpdateError: nil,
			request:             `{"comment":"test comment"}`,
			statusCode:          http.StatusNotFound,
			responseIsPost:      false,
			responseMessage:     "{\"message\":\"post not found\"}\n",
		},
		{
			valueVars:           map[string]string{"post_id": "0"},
			contextKey:          middleware.UserContextKey,
			contextValue:        user.User{Username: "test", UserID: 0},
			postID:              0,
			postRepoLockStatus:  true,
			postRepoGetPost:     post.Post{},
			postRepoGetError:    fmt.Errorf("test error"),
			postRepoUpdateError: nil,
			request:             `{"comment":"test comment"}`,
			statusCode:          http.StatusInternalServerError,
			responseIsPost:      false,
			responseMessage:     "",
		},
		{
			valueVars:           map[string]string{"post_id": "0"},
			contextKey:          middleware.UserContextKey,
			contextValue:        user.User{Username: "test", UserID: 0},
			postID:              0,
			postRepoLockStatus:  true,
			postRepoGetPost:     post.Post{},
			postRepoGetError:    nil,
			postRepoUpdateError: fmt.Errorf("test error"),
			request:             `{"comment":"test comment"}`,
			statusCode:          http.StatusInternalServerError,
			responseIsPost:      false,
			responseMessage:     "",
		},
	}

	for _, testCase := range testCases {
		postHandler := setupPost()
		defer postHandler.Logger.Sync()

		r := httptest.NewRequest("GET", "/api/posts/", strings.NewReader(testCase.request))
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
			On("Update", mock.AnythingOfType("post.Post")).
			Return(testCase.postRepoUpdateError)

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Unlock", testCase.postID).
			Return(true)

		postHandler.CommentAdd(w, r.WithContext(ctx))

		resp := w.Result()

		body, errRead := ioutil.ReadAll(resp.Body)
		require.NoError(t, errRead)
		require.Equal(t, resp.StatusCode, testCase.statusCode)
		if testCase.responseIsPost {
			var post post.Post
			errUnmarshal := json.Unmarshal(body, &post)
			require.NoError(t, errUnmarshal)
			post.Time = ""
			post.Comments[0].Time = ""
			require.Equal(t, post, testCase.responsePost)
		} else {
			require.Equal(t, string(body), testCase.responseMessage)
		}
	}
}

func TestCommentRemove(t *testing.T) {
	type testCase struct {
		valueVars    map[string]string
		contextKey   middleware.Key
		contextValue interface{}
		postID       uint64

		postRepoLockStatus  bool
		postRepoGetPost     post.Post
		postRepoGetError    error
		postRepoUpdateError error

		statusCode      int
		responseIsPost  bool
		responsePost    post.Post
		responseMessage string
	}

	testCases := []testCase{
		{
			valueVars: map[string]string{
				"post_id": "0", "comment_id": "0",
			},
			contextKey:         middleware.UserContextKey,
			contextValue:       user.User{Username: "test", UserID: 0},
			postID:             0,
			postRepoLockStatus: true,
			postRepoGetPost: post.Post{
				Comments: []comment.Comment{
					{
						Author: user.User{Username: "test", UserID: 0},
						Body:   "test comment",
						ID:     0,
					},
				},
			},
			postRepoGetError:    nil,
			postRepoUpdateError: nil,
			statusCode:          http.StatusOK,
			responseIsPost:      true,
			responsePost:        post.Post{Comments: []comment.Comment{}},
		},
		{
			valueVars: map[string]string{
				"wrong post_id": "0", "comment_id": "0",
			},
			contextKey:          middleware.UserContextKey,
			contextValue:        user.User{Username: "test", UserID: 0},
			postID:              0,
			postRepoLockStatus:  true,
			postRepoGetPost:     post.Post{},
			postRepoGetError:    nil,
			postRepoUpdateError: nil,
			statusCode:          http.StatusInternalServerError,
			responseIsPost:      false,
			responseMessage:     "",
		},
		{
			valueVars: map[string]string{
				"post_id": "0", "wrong comment_id": "0",
			},
			contextKey:          middleware.UserContextKey,
			contextValue:        user.User{Username: "test", UserID: 0},
			postID:              0,
			postRepoLockStatus:  true,
			postRepoGetPost:     post.Post{},
			postRepoGetError:    nil,
			postRepoUpdateError: nil,
			statusCode:          http.StatusInternalServerError,
			responseIsPost:      false,
			responseMessage:     "",
		},
		{
			valueVars: map[string]string{
				"post_id": "0", "comment_id": "0",
			},
			contextKey:          middleware.AuthtorizationContextKey,
			contextValue:        user.User{Username: "test", UserID: 0},
			postID:              0,
			postRepoLockStatus:  true,
			postRepoGetPost:     post.Post{},
			postRepoGetError:    nil,
			postRepoUpdateError: nil,
			statusCode:          http.StatusInternalServerError,
			responseIsPost:      false,
			responseMessage:     "Internal server error\n",
		},
		{
			valueVars: map[string]string{
				"post_id": "0", "comment_id": "0",
			},
			contextKey:          middleware.UserContextKey,
			contextValue:        user.User{Username: "test", UserID: 0},
			postID:              0,
			postRepoLockStatus:  false,
			postRepoGetPost:     post.Post{},
			postRepoGetError:    nil,
			postRepoUpdateError: nil,
			statusCode:          http.StatusNotFound,
			responseIsPost:      false,
			responseMessage:     "{\"message\":\"post not found\"}\n",
		},
		{
			valueVars: map[string]string{
				"post_id": "0", "comment_id": "0",
			},
			contextKey:         middleware.UserContextKey,
			contextValue:       user.User{Username: "test2", UserID: 0},
			postID:             0,
			postRepoLockStatus: true,
			postRepoGetPost: post.Post{
				Comments: []comment.Comment{
					{
						Author: user.User{Username: "test", UserID: 0},
						Body:   "test comment",
						ID:     0,
					},
				},
			},
			postRepoGetError:    nil,
			postRepoUpdateError: nil,
			statusCode:          http.StatusNotFound,
			responseIsPost:      false,
			responseMessage:     "{\"message\":\"this comment doesn't belong to this user\"}\n",
		},
		{
			valueVars: map[string]string{
				"post_id": "0", "comment_id": "0",
			},
			contextKey:         middleware.UserContextKey,
			contextValue:       user.User{Username: "test", UserID: 0},
			postID:             0,
			postRepoLockStatus: true,
			postRepoGetPost: post.Post{
				Comments: []comment.Comment{
					{
						Author: user.User{Username: "test", UserID: 0},
						Body:   "test comment",
						ID:     0,
					},
				},
			},
			postRepoGetError:    nil,
			postRepoUpdateError: fmt.Errorf("test error"),
			statusCode:          http.StatusInternalServerError,
			responseIsPost:      false,
			responseMessage:     "",
		},
		{
			valueVars: map[string]string{
				"post_id": "0", "comment_id": "0",
			},
			contextKey:          middleware.UserContextKey,
			contextValue:        user.User{Username: "test", UserID: 0},
			postID:              0,
			postRepoLockStatus:  true,
			postRepoGetPost:     post.Post{},
			postRepoGetError:    nil,
			postRepoUpdateError: nil,
			statusCode:          http.StatusNotFound,
			responseIsPost:      false,
			responseMessage:     "{\"message\":\"comment not found\"}\n",
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
			On("Update", mock.AnythingOfType("post.Post")).
			Return(testCase.postRepoUpdateError)

		postHandler.PostRepo.(*mocks.PostRepo).
			On("Unlock", testCase.postID).
			Return(true)

		postHandler.CommentRemove(w, r.WithContext(ctx))

		resp := w.Result()

		body, errRead := ioutil.ReadAll(resp.Body)
		require.NoError(t, errRead)
		require.Equal(t, resp.StatusCode, testCase.statusCode)
		if testCase.responseIsPost {
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
