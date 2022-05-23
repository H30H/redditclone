package frontendMessages

import (
	"encoding/json"
	"fmt"
	"net/http"
	"redditclone/pkg/errors"

	"go.uber.org/zap"
)

type ErrorMessage struct {
	Location string      `json:"location"`
	Param    string      `json:"param"`
	Value    interface{} `json:"value,omitempty"`
	Message  string      `json:"msg"`
}

type Error struct {
	Errors []ErrorMessage `json:"errors"`
}

type Message struct {
	Message string `json:"message"`
}

type Vote struct {
	UserID int64 `json:"user,string"`
	Vote   int   `json:"vote"`
}

func SendMessage(w http.ResponseWriter, message string, code int, logger *zap.SugaredLogger, from string) {
	res, errMarshal := json.Marshal(Message{
		Message: message,
	})
	if errMarshal != nil {
		errors.SendHttpError(
			logger, w, fmt.Errorf("%s: %w", from, errors.ErrMarshal{Err: errMarshal}),
		)
		return
	}
	http.Error(w, string(res), code)
}
