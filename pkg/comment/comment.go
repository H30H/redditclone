package comment

import (
	"redditclone/pkg/user"
)

type Comment struct {
	Author user.User `json:"author"`
	Body   string    `json:"body"`
	Time   string    `json:"created"`
	ID     uint64    `json:"id,string"`
}
