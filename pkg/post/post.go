package post

import (
	"math"
	"redditclone/pkg/comment"
	"redditclone/pkg/frontendMessages"
	"redditclone/pkg/user"
)

type Post struct {
	Title            string                  `json:"title"`
	Text             string                  `json:"text,omitempty"`
	URL              string                  `json:"url,omitempty"`
	Author           user.User               `json:"author"`
	Category         string                  `json:"category"`
	ID               uint64                  `json:"id,string"`
	Time             string                  `json:"created"`
	Score            int64                   `json:"score"`
	Views            uint                    `json:"views"`
	UpvotePercentage int64                   `json:"upvotePercentage"`
	Type             string                  `json:"type"`
	Votes            []frontendMessages.Vote `json:"votes"`
	Comments         []comment.Comment       `json:"comments"`
	CommentID        uint64                  `json:"-"`
}

/*
* author: {username: "user_pro4", id: "6257edff7ad43200093e7d31"}
* category: "music"
 * comments: []
 * created: "2022-04-15T11:40:57.429Z"
 * id: "625959c9359d030009a34c72"
 * score: 1
 * text: "kjhygtfdx "
 * title: "asdfgbhnjmk"
 * type: "text"
 * upvotePercentage: 100
 * views: 0
 * votes: [{user: "6257edff7ad43200093e7d31", vote: 1}]
*/

func (p *Post) GetVotes() {
	var score int64
	var upvoted uint
	defer func() {
		p.Score = score
		p.UpvotePercentage = int64(upvoted)
	}()
	if len(p.Votes) == 0 {
		return
	}
	for _, vote := range p.Votes {
		score += int64(vote.Vote)
		if vote.Vote == 1 {
			upvoted++
		}
	}
	upvoted = func(a1, a2 uint) uint {
		if a2 == 0 {
			return 0
		}
		res := float64(a1) / float64(a2)
		t := math.Trunc(res)
		if math.Abs(res-t) >= 0.5 {
			return uint(t+math.Copysign(1, res)) * 100
		}
		return uint(res) * 100
	}(upvoted, uint(len(p.Votes)))
}

func (p *Post) GetID() (res uint64) {
	res = p.CommentID
	p.CommentID++
	return
}
