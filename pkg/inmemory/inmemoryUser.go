package inmemory

import (
	"redditclone/pkg/user"
	"sync"
)

func NewDatabaseUser() *UserRepo {
	return &UserRepo{
		data: make(map[string]user.User),
		mux:  &sync.Mutex{},
	}
}

func (d *UserRepo) Add(user user.User) bool {
	d.mux.Lock()
	defer d.mux.Unlock()
	_, ok := d.data[user.Username]
	if ok {
		return false
	}
	d.data[user.Username] = user
	return true
}

func (d *UserRepo) Remove(username string) bool {
	d.mux.Lock()
	defer d.mux.Unlock()
	_, ok := d.data[username]
	if !ok {
		return false
	}
	delete(d.data, username)
	return true
}

func (d *UserRepo) Find(username string) (user.User, bool) {
	d.mux.Lock()
	defer d.mux.Unlock()
	res, ok := d.data[username]
	if !ok {
		return user.User{}, false
	}
	return res, true
}

func (d *UserRepo) GetID() (res uint64) {
	res = d.count
	d.count++
	return
}
