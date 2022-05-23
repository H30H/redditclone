package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"redditclone/pkg/comment"
	"redditclone/pkg/errors"
	"redditclone/pkg/frontendMessages"
	"redditclone/pkg/middleware"
	"redditclone/pkg/token"
	"redditclone/pkg/user"
	"time"

	"github.com/gorilla/mux"
)

func (h *PostHandler) CommentAdd(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	vars := mux.Vars(r)
	id, errGet := token.GetMapItemUint64(vars, "post_id")
	if errGet != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postAddComment: %w", errors.ErrRequest{Err: errGet}),
		)
		return
	}
	usr, ok := r.Context().Value(middleware.UserContextKey).(user.User)
	if !ok {
		h.Logger.Errorf("no context value: %s", middleware.UserContextKey)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	type readComment struct {
		Body string `json:"comment"`
	}
	readCmt := readComment{}
	errUnmarchal := json.Unmarshal(body, &readCmt)
	if errUnmarchal != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postAddComment: %w", errors.ErrUnmarshalRequest{Err: errUnmarchal}),
		)
		return
	}
	if readCmt.Body == "" || len(readCmt.Body) >= 2000 {
		var err frontendMessages.ErrorMessage
		if readCmt.Body == "" {
			err = frontendMessages.ErrorMessage{
				Location: "body",
				Param:    "comment",
				Message:  "is required",
			}
		} else {
			err = frontendMessages.ErrorMessage{
				Location: "body",
				Param:    "comment",
				Message:  "must be at most 2000 characters long",
			}
		}
		res, errMarshal := json.Marshal(frontendMessages.Error{Errors: []frontendMessages.ErrorMessage{err}})
		if errMarshal != nil {
			errors.SendHttpError(
				h.Logger, w,
				fmt.Errorf("postAddComment: %w", errors.ErrMarshal{Err: errMarshal}),
			)
			return
		}
		http.Error(w, string(res), http.StatusUnprocessableEntity)
		return
	}
	if !h.PostRepo.Lock(id) {
		frontendMessages.SendMessage(w,
			"post not found",
			http.StatusNotFound,
			h.Logger, "postAddComment",
		)
		return
	}
	pst, errGet := h.PostRepo.Get(id)
	if errGet != nil {
		h.PostRepo.Unlock(id)
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postAddComment: %w", errGet),
		)
		return
	}
	cmt := comment.Comment{
		Author: usr,
		Body:   readCmt.Body,
		ID:     pst.GetID(),
		Time:   time.Now().Format(time.RFC3339),
	}
	pst.Comments = append(pst.Comments, cmt)
	errUpdate := h.PostRepo.Update(pst)
	h.PostRepo.Unlock(id)
	if errUpdate != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postAddComment: %w", errGet),
		)
		return
	}
	pstJson, err := json.Marshal(pst)
	if err != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postAddComment: %w", errors.ErrMarshal{Err: err}),
		)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(pstJson)
}

func (h *PostHandler) CommentRemove(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idPost, errIdPost := token.GetMapItemUint64(vars, "post_id")
	if errIdPost != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("PostRemoveComment: %W", errIdPost),
		)
		return
	}
	idComment, errIdComment := token.GetMapItemUint64(vars, "comment_id")
	if errIdComment != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("PostRemoveComment: %W", errIdComment),
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
			h.Logger, "PostRemoveComment",
		)
		return
	}
	pst, _ := h.PostRepo.Get(idPost)
	flag := true
	for i, cmt := range pst.Comments {
		if cmt.ID == idComment {
			if cmt.Author.UserID != usr.UserID || cmt.Author.Username != usr.Username {
				h.Logger.Errorf("PostRemoveComment: can`t remove comment: Author: {Username: %s, UserID: %d}, User: {Username: %s, UserID: %d}, commeniID: %d",
					cmt.Author.Username, cmt.Author.UserID, usr.Username, usr.UserID, idComment,
				)
				frontendMessages.SendMessage(w,
					"this comment doesn't belong to this user",
					http.StatusNotFound,
					h.Logger, "PostRemoveComment",
				)
				h.PostRepo.Unlock(idPost)
				return
			}
			pst.Comments = token.RemoveInArr(pst.Comments, uint(i))
			flag = false
			errUpdate := h.PostRepo.Update(pst)
			if errUpdate != nil {
				h.PostRepo.Unlock(idPost)
				errors.SendHttpError(
					h.Logger, w,
					fmt.Errorf("postRemoveComment: %w", errUpdate),
				)
				return
			}
			break
		}
	}
	h.PostRepo.Unlock(idPost)
	if flag {
		frontendMessages.SendMessage(w,
			"comment not found",
			http.StatusNotFound,
			h.Logger, "PostRemoveComment",
		)
		return
	}
	res, errMarshal := json.Marshal(pst)
	if errMarshal != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postRemoveComment: %w", errors.ErrMarshal{Err: errMarshal}),
		)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}
