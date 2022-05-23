package inmemory

import (
	"redditclone/pkg/comment"
	"sync"
)

func NewDatabaseComment() *CommentRepo {
	return &CommentRepo{
		data: make(map[uint64]comment.Comment),
		mux:  &sync.Mutex{},
	}
}

func (d *CommentRepo) Add(cmt comment.Comment) bool {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.data[cmt.ID] = cmt
	return true
}

func (d *CommentRepo) Find(id uint64) (comment.Comment, bool) {
	d.mux.Lock()
	defer d.mux.Unlock()
	res, ok := d.data[id]
	if !ok {
		return comment.Comment{}, false
	}
	return res, true
}

func (d *CommentRepo) Remove(id uint64) bool {
	d.mux.Lock()
	defer d.mux.Unlock()
	_, ok := d.data[id]
	if !ok {
		return false
	}
	delete(d.data, id)
	return true
}

func (d *CommentRepo) GetID() (res uint64) {
	res = d.count
	d.count++
	return
}
