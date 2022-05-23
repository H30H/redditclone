package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"redditclone/pkg/errors"
	"redditclone/pkg/frontendMessages"
	"redditclone/pkg/middleware"
	"redditclone/pkg/post"
	"redditclone/pkg/token"
	"redditclone/pkg/user"
	"time"

	"github.com/gorilla/mux"
)

func (h *PostHandler) PostAdd(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	usr, ok := r.Context().Value(middleware.UserContextKey).(user.User)
	if !ok {
		h.Logger.Errorf("no context value: %s", middleware.UserContextKey)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	pst := post.Post{}
	errUnmarshal := json.Unmarshal([]byte(body), &pst)
	if errUnmarshal != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postAdd: %w", errors.ErrUnmarshalRequest{Err: errUnmarshal}),
		)
		return
	}
	pst.Author = usr
	pst.Time = time.Now().Format(time.RFC3339)
	pst.Votes = []frontendMessages.Vote{{UserID: usr.UserID, Vote: 1}}
	pst.GetVotes()
	errAdd := h.PostRepo.Add(&pst)
	if errAdd != nil {
		h.Logger.Errorf("can`t add post to database: %s", errAdd)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.Logger.Debugf("adding post with id: %d", pst.ID)
	pstStr, errConvJson := json.Marshal(pst)
	if errConvJson != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postAdd: %w", errors.ErrMarshal{Err: errConvJson}),
		)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(pstStr)
}

func (h *PostHandler) PostGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, errGet := token.GetMapItemUint64(vars, "post_id")
	h.Logger.Debugf("getting post with id: %d", id)
	if errGet != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postGet: %w", errors.ErrRequest{Err: errGet}),
		)
		return
	}
	if !h.PostRepo.Lock(id) {
		frontendMessages.SendMessage(w,
			"post not found",
			http.StatusNotFound,
			h.Logger, "postGet",
		)
		return
	}
	pst, errGet := h.PostRepo.Get(id)
	if errGet != nil {
		h.PostRepo.Unlock(id)
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postGet: %w", errGet),
		)
		return
	}
	pst.Views++
	errUpdate := h.PostRepo.Update(pst)
	if errUpdate != nil {
		h.PostRepo.Unlock(id)
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postGet: %w", errUpdate),
		)
		return
	}
	ok := h.PostRepo.Unlock(id)
	if !ok {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postGet: %w", fmt.Errorf("can`t unlock post")),
		)
		return
	}
	pstJson, err := json.Marshal(pst)
	if err != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postGet: %w", errors.ErrMarshal{Err: err}),
		)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(pstJson)
}

func (h *PostHandler) Posts(w http.ResponseWriter, r *http.Request) {
	postsStr, err := h.PostRepo.ToJson("", "")
	if err != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("login: %w", errors.ErrMarshal{Err: err}),
		)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(postsStr)
}

func (h *PostHandler) Categories(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category, errGet := token.GetMapItemString(vars, "category_name")
	if errGet != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("Categories: %w", errors.ErrRequest{Err: errGet}),
		)
		return
	}
	postsStr, err := h.PostRepo.ToJson(category, "")
	if err != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("Categories: %w", errors.ErrMarshal{Err: err}),
		)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(postsStr))
}

func (h *PostHandler) UserPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username, errGet := token.GetMapItemString(vars, "user_login")
	if errGet != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("UserPosts: %W", errGet),
		)
		return
	}
	postsStr, err := h.PostRepo.ToJson("", username)
	if err != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("UserPosts: %w", errors.ErrMarshal{Err: err}),
		)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(postsStr))
}

func (h *PostHandler) PostRemove(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idPost, errGet := token.GetMapItemUint64(vars, "post_id")
	if errGet != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("PostRemove: %W", errGet),
		)
		return
	}
	usr, ok := r.Context().Value(middleware.UserContextKey).(user.User)
	if !ok {
		h.Logger.Errorf("no context value: %s", middleware.UserContextKey)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !h.PostRepo.Lock(idPost) {
		frontendMessages.SendMessage(w,
			"post not found",
			http.StatusNotFound,
			h.Logger, "PostRemove",
		)
		return
	}
	pst, errGet := h.PostRepo.Get(idPost)
	if errGet != nil {
		h.PostRepo.Unlock(idPost)
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postRemove: %w", errGet),
		)
		return
	}
	if pst.Author.Username != usr.Username || pst.Author.UserID != usr.UserID {
		h.PostRepo.Unlock(idPost)
		frontendMessages.SendMessage(w,
			"this post doesn't belong to this user",
			http.StatusNotFound,
			h.Logger, "PostRemove",
		)
		return
	}
	ok = h.PostRepo.Remove(idPost)
	h.PostRepo.Unlock(idPost)
	if !ok {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postRemove: can`t remove post with id %d", idPost),
		)
		return
	}
	frontendMessages.SendMessage(w,
		"success",
		http.StatusOK,
		h.Logger, "PostRemove",
	)
}
