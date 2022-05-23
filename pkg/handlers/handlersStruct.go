package handlers

import (
	"redditclone/pkg/database"

	"go.uber.org/zap"
)

type UserHandler struct {
	Logger    *zap.SugaredLogger
	UserRepo  database.UserRepo
	SecretKey string
}

type PostHandler struct {
	Logger    *zap.SugaredLogger
	PostRepo  database.PostRepo
	SecretKey string
}
