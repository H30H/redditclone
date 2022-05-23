package handlers

import (
	"redditclone/pkg/database/mocks"

	"go.uber.org/zap"
)

func setupUser() *UserHandler {
	zapLogger := zap.NewNop()
	logger := zapLogger.Sugar()

	return &UserHandler{
		Logger:    logger,
		UserRepo:  &mocks.UserRepo{},
		SecretKey: "test key",
	}
}

func setupPost() *PostHandler {
	zapLogger := zap.NewNop()
	logger := zapLogger.Sugar()

	return &PostHandler{
		Logger:    logger,
		PostRepo:  &mocks.PostRepo{},
		SecretKey: "test key",
	}
}
