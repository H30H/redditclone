package database

import (
	"redditclone/pkg/user"
	"sync"
)

type UserRepo interface {
	Add(user *user.User) (err error)
	Find(username string) (user.User, bool)
}

type UserRepoStruct struct {
	data *DatabaseUser
	mx   *sync.Mutex
}

func NewUserRepo(databaseUser *DatabaseUser) *UserRepoStruct {
	return &UserRepoStruct{
		data: databaseUser,
		mx:   &sync.Mutex{},
	}
}

func (d *UserRepoStruct) Add(user *user.User) (err error) {
	d.mx.Lock()
	defer d.mx.Unlock()
	userID, err := d.data.Add(*user)
	if err == nil {
		user.UserID = userID
	}
	return
}

func (d *UserRepoStruct) Find(username string) (user.User, bool) {
	d.mx.Lock()
	defer d.mx.Unlock()
	usr, err := d.data.Get(username)
	if err != nil {
		return user.User{}, false
	}
	return usr, true
}
