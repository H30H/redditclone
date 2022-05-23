package main

import (
	"log"
	"net/http"
	"net/url"
	"redditclone/pkg/database"
	"redditclone/pkg/handlers"
	"redditclone/pkg/middleware"
	"redditclone/pkg/session"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const (
	secretKey       = "kekw"
	staticDirectory = "./web/"
)

func main() {
	//root:testpass12345@tcp(localhost:3306)
	panicOnErr := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	databaseUser, errUser := database.InitDatabaseUser("root:testpass12345@(localhost:3306)", "redditclone")
	panicOnErr(errUser)
	defer databaseUser.Close()
	userBase := database.NewUserRepo(databaseUser)

	databasePost, errPost := database.InitDatabasePost("mongodb://localhost:27017", "redditclone", "posts")
	panicOnErr(errPost)
	defer databasePost.Close()
	postBase, errPost := database.NewPostRepo(databasePost, databaseUser)
	panicOnErr(errPost)

	databaseSession, err := session.InitDatabaseSession("root:testpass12345@(localhost:3306)", "redditclone")
	panicOnErr(err)
	sessionManager := session.InitSessionManager(databaseSession)
	defer sessionManager.Close()

	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

	go func() {
		ticker := time.NewTicker(time.Second * 5)
		for {
			<-ticker.C
			errDemon := sessionManager.CheckAllTimes(logger)
			if errDemon != nil {
				logger.Debugf("demonErr: %s", errDemon)
			}
		}
	}()

	userHandler := &handlers.UserHandler{
		UserRepo:  userBase,
		Logger:    logger,
		SecretKey: secretKey,
	}

	handler := &handlers.PostHandler{
		PostRepo:  postBase,
		Logger:    logger,
		SecretKey: secretKey,
	}

	middleware := &middleware.Middleware{
		Authorization: sessionManager,
		Users:         userBase,
		Logger:        logger,
		Secretkey:     secretKey,
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/register", middleware.AddAuth(userHandler.Register)).Methods("POST")
	r.HandleFunc("/api/login", middleware.AddAuth(userHandler.Login)).Methods("POST")
	r.HandleFunc("/api/posts/", handler.Posts).Methods("GET")
	r.HandleFunc("/api/posts", middleware.CheckAuth(handler.PostAdd)).Methods("POST")
	r.HandleFunc("/api/posts/{category_name}", handler.Categories).Methods("GET")
	r.HandleFunc("/api/post/{post_id:[0-9]+}", handler.PostGet).Methods("GET")
	r.HandleFunc("/api/post/{post_id:[0-9]+}", middleware.CheckAuth(handler.CommentAdd)).Methods("POST")
	r.HandleFunc("/api/post/{post_id:[0-9]+}/{comment_id:[0-9]+}", middleware.CheckAuth(handler.CommentRemove)).Methods("DELETE")
	r.HandleFunc("/api/post/{post_id:[0-9]+}/upvote", middleware.CheckAuth(handler.PostRatingUp)).Methods("GET")
	r.HandleFunc("/api/post/{post_id:[0-9]+}/unvote", middleware.CheckAuth(handler.PostRatingDefault)).Methods("GET")
	r.HandleFunc("/api/post/{post_id:[0-9]+}/downvote", middleware.CheckAuth(handler.PostRatingDown)).Methods("GET")
	r.HandleFunc("/api/post/{post_id:[0-9]+}", middleware.CheckAuth(handler.PostRemove)).Methods("DELETE")
	r.HandleFunc("/api/user/{user_login}", handler.UserPosts).Methods("GET")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDirectory))))
	r.PathPrefix("/").Handler(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = "/"
			h.ServeHTTP(w, r2)
		})
	}(http.FileServer(http.Dir(staticDirectory + "html/"))))

	log.Printf("start at \"localhost:8080\"")
	http.ListenAndServe("localhost:8080", r)
}
