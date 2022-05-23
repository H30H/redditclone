package errors

import (
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type ErrSignToken struct {
	Err error
}

type ErrMarshal struct {
	Err error
}

type ErrUnmarshalRequest struct {
	Err error
}

type ErrUnmarshalData struct {
	Err error
}

type ErrRequest struct {
	Err error
}

type ErrBadToken struct {
	Err error
}

type ErrBadContext struct {
	Err error
}

func SendHttpError(logger *zap.SugaredLogger, w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	logger.Errorf(err.Error())
	caseFunc := func(errs ...error) bool {
		for _, e := range errs {
			if errors.Is(err, e) {
				return true
			}
		}
		return false
	}
	switch {
	case caseFunc(ErrBadToken{}):
		w.WriteHeader(http.StatusUnauthorized)
	case caseFunc(ErrSignToken{},
		ErrMarshal{},
		ErrUnmarshalData{},
	):
		w.WriteHeader(http.StatusInternalServerError)
	case caseFunc(ErrRequest{}, ErrUnmarshalRequest{}):
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func (err ErrBadToken) Error() string {
	return fmt.Errorf("bad token: %w", err.Err).Error()
}

func (err ErrSignToken) Error() string {
	return fmt.Errorf("can't sign token: %w", err.Err).Error()
}

func (err ErrMarshal) Error() string {
	return fmt.Errorf("json marshal error: %w", err.Err).Error()
}

func (err ErrRequest) Error() string {
	return fmt.Errorf("bad request: %w", err.Err).Error()
}

func (err ErrUnmarshalRequest) Error() string {
	return fmt.Errorf("json unmarshal error: %w", err.Err).Error()
}

func (err ErrUnmarshalData) Error() string {
	return fmt.Errorf("json unmarshal error: %w", err.Err).Error()
}
