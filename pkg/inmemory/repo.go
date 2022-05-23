package inmemory

import (
	"redditclone/pkg/comment"
	"redditclone/pkg/post"
	"redditclone/pkg/user"
	"sync"
)

type UserRepo struct {
	data  map[string]user.User
	mux   *sync.Mutex
	count uint64
}

type PostRepo struct {
	data       map[uint64]post.Post
	mu         *sync.Mutex
	postsMuxes map[uint64]*sync.Mutex
	count      uint64
}

type CommentRepo struct {
	data  map[uint64]comment.Comment
	mux   *sync.Mutex
	count uint64
}
