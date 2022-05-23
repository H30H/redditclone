package inmemory

import (
	"encoding/json"
	"fmt"
	"redditclone/pkg/post"
	"sort"
	"sync"
)

func NewDatabasePost() *PostRepo {
	return &PostRepo{
		data:       make(map[uint64]post.Post),
		postsMuxes: make(map[uint64]*sync.Mutex),
		mu:         &sync.Mutex{},
	}
}

func (d *PostRepo) Add(pst post.Post) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.data[pst.ID] = pst
	_, ok := d.postsMuxes[pst.ID]
	if !ok {
		d.postsMuxes[pst.ID] = &sync.Mutex{}
	}
	return true
}

func (d *PostRepo) Find(id uint64) (ok bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok = d.data[id]
	return
}

func (d *PostRepo) Get(id uint64) (pst post.Post, ok bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	pst, ok = d.data[id]
	return
}

func (d *PostRepo) Lock(id uint64) bool {
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

func (d *PostRepo) Unlock(id uint64) bool {
	mu, ok := d.postsMuxes[id]
	if !ok {
		return false
	}
	mu.Unlock()
	return true
}

func (d *PostRepo) Remove(id uint64) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.data[id]
	if !ok {
		return false
	}
	delete(d.data, id)
	delete(d.postsMuxes, id)
	return true
}

func (d *PostRepo) ToJson(category, username string) ([]byte, error) {
	if len(d.data) == 0 {
		return []byte("[]"), nil
	}
	var compareCategory func(s1, s2 string) bool
	if category == "" {
		compareCategory = func(s1, s2 string) bool { return true }
	} else {
		compareCategory = func(s1, s2 string) bool { return s1 == s2 }
	}
	var compareUsername func(s1, s2 string) bool
	if username == "" {
		compareUsername = func(s1, s2 string) bool { return true }
	} else {
		compareUsername = func(s1, s2 string) bool { return s1 == s2 }
	}
	res := []post.Post{}
	for _, pst := range d.data {
		if !compareCategory(pst.Category, category) || !compareUsername(pst.Author.Username, username) {
			continue
		}
		res = append(res, pst)
	}

	sort.Slice(res, func(i1, i2 int) bool {
		return res[i1].Score > res[i2].Score
	})
	resJson, errMershal := json.Marshal(res)
	if errMershal != nil {
		return nil, fmt.Errorf("err in encoding posts to json: %w", errMershal)
	}
	return resJson, nil
}

func (d *PostRepo) GetID() (res uint64) {
	res = d.count
	d.count++
	return
}
