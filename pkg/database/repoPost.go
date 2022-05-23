package database

import (
	"encoding/json"
	"fmt"
	"redditclone/pkg/post"
	"redditclone/pkg/user"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostRepo interface {
	Lock(id uint64) bool
	Unlock(id uint64) bool
	Add(pst *post.Post) (err error)
	Find(id uint64) (ok bool)
	Get(id uint64) (pst post.Post, err error)
	Update(pst post.Post) (err error)
	Remove(id uint64) bool
	ToJson(category, username string) ([]byte, error)
}

type PostRepoStruct struct {
	users      *DatabaseUser
	data       DatabasePost
	mx         *sync.Mutex
	postsMuxes map[uint64]*sync.Mutex
	idGetter   uint64
}

func NewPostRepo(databasePost DatabasePost, databaseUser *DatabaseUser) (*PostRepoStruct, error) {
	mx := map[uint64]*sync.Mutex{}
	posts, err := databasePost.GetAll(bson.M{}, options.Find())
	if err != nil {
		return nil, err
	}
	var maxID uint64
	for _, pst := range posts {
		mx[pst.ID] = &sync.Mutex{}
		if maxID < pst.ID {
			maxID = pst.ID
		}
	}
	return &PostRepoStruct{
		users:      databaseUser,
		data:       databasePost,
		mx:         &sync.Mutex{},
		postsMuxes: mx,
		idGetter:   maxID,
	}, nil
}

func (d *PostRepoStruct) Lock(id uint64) bool {
	mu, ok := d.postsMuxes[id]
	if !ok {
		return false
	}
	mu.Lock()
	if !d.Find(id) {
		mu.Unlock()
		return false
	}
	return true
}

func (d *PostRepoStruct) Unlock(id uint64) bool {
	mu, ok := d.postsMuxes[id]
	if !ok {
		return false
	}
	mu.Unlock()
	return true
}

func (d *PostRepoStruct) Add(pst *post.Post) (err error) {
	if pst == nil {
		return fmt.Errorf("nil pointer post")
	}
	d.mx.Lock()
	defer d.mx.Unlock()
	d.idGetter++
	pst.ID = d.idGetter
	err = d.data.Insert(*pst)
	if err != nil {
		d.idGetter--
		return err
	}
	d.postsMuxes[pst.ID] = &sync.Mutex{}
	return
}

func (d *PostRepoStruct) Find(id uint64) (ok bool) {
	d.mx.Lock()
	defer d.mx.Unlock()
	var pst post.Post
	err := d.data.Find(id, &pst)
	return err == nil
}

func (d *PostRepoStruct) Get(id uint64) (pst post.Post, err error) {
	d.mx.Lock()
	defer d.mx.Unlock()
	err = d.data.Find(id, &pst)
	return
}

func (d *PostRepoStruct) Update(pst post.Post) (err error) {
	d.mx.Lock()
	defer d.mx.Unlock()
	err = d.data.Replace(pst)
	return
}

func (d *PostRepoStruct) Remove(id uint64) bool {
	d.mx.Lock()
	defer d.mx.Unlock()
	err := d.data.Delete(id)
	if err != nil {
		return false
	}
	delete(d.postsMuxes, id)
	return true
}

func (d *PostRepoStruct) getUserID(username string) int64 {
	usr, err := d.users.Get(username)
	if err != nil {
		return -1
	}
	return usr.UserID
}

func (d *PostRepoStruct) ToJson(category, username string) ([]byte, error) {
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"score": -1})
	filter := bson.M{}
	if category != "" {
		filter["category"] = category
	}
	if username != "" {
		filter["author"] = user.User{Username: username, UserID: d.getUserID(username)}
	}
	resArr, err := d.data.GetAll(filter, findOptions)
	if err != nil {
		return nil, err
	}
	res, errMarshal := json.Marshal(resArr)
	return res, errMarshal
}
