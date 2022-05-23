package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"redditclone/pkg/errors"
	"redditclone/pkg/frontendMessages"
	"redditclone/pkg/middleware"
	"redditclone/pkg/token"
	"redditclone/pkg/user"

	"github.com/gorilla/mux"
)

func (h *PostHandler) setVoice(w http.ResponseWriter, r *http.Request, value int) error {
	usr, ok := r.Context().Value(middleware.UserContextKey).(user.User)
	if !ok {
		h.Logger.Errorf("no context value: UserContextKey")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil
	}
	vars := mux.Vars(r)
	postID, errGet := token.GetMapItemUint64(vars, "post_id")
	if errGet != nil {
		return errGet
	}
	if !h.PostRepo.Lock(postID) {
		frontendMessages.SendMessage(w,
			"post not found",
			http.StatusNotFound,
			h.Logger, "setVoice",
		)
		return nil
	}
	flag := value != 0
	pst, errGet := h.PostRepo.Get(postID)
	if errGet != nil {
		h.PostRepo.Unlock(postID)
		return errGet
	}
	for i, vt := range pst.Votes {
		if vt.UserID == usr.UserID {
			if flag {
				flag = false
				pst.Votes[i].Vote = value
				break
			}
			pst.Votes = token.RemoveInArr(pst.Votes, uint(i))
		}
	}
	if flag {
		pst.Votes = append(pst.Votes, frontendMessages.Vote{UserID: usr.UserID, Vote: value})
	}
	pst.GetVotes()
	errUpdate := h.PostRepo.Update(pst)
	if errUpdate != nil {
		h.PostRepo.Unlock(postID)
		return errUpdate
	}
	ok = h.PostRepo.Unlock(postID)
	if !ok {
		return fmt.Errorf("can`t unlock post")
	}
	res, err := json.Marshal(pst)
	if err != nil {
		return errors.ErrMarshal{Err: err}
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
	return nil
}

func (h *PostHandler) PostRatingUp(w http.ResponseWriter, r *http.Request) {
	err := h.setVoice(w, r, 1)
	if err != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postRatingUp: %w", err),
		)
	}
}

func (h *PostHandler) PostRatingDown(w http.ResponseWriter, r *http.Request) {
	err := h.setVoice(w, r, -1)
	if err != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postRatingDown: %w", err),
		)
	}
}

func (h *PostHandler) PostRatingDefault(w http.ResponseWriter, r *http.Request) {
	err := h.setVoice(w, r, 0)
	if err != nil {
		errors.SendHttpError(
			h.Logger, w,
			fmt.Errorf("postRatingDefault: %w", err),
		)
	}
}
