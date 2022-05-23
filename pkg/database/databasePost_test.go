package database

import (
	"encoding/json"
	"fmt"
	"redditclone/pkg/database/mocks"
	"redditclone/pkg/post"
	"redditclone/pkg/user"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var (
	users = []*user.User{
		{
			Username:     "test1",
			PasswordHash: "test1",
			UserID:       1,
		},
		{
			Username:     "test2",
			PasswordHash: "test2",
			UserID:       2,
		},
		{
			Username:     "test3",
			PasswordHash: "test3",
			UserID:       3,
		},
		{
			Username:     "test4",
			PasswordHash: "test4",
			UserID:       4,
		},
	}
	posts = []*post.Post{
		{
			Author:           *users[0],
			Category:         "music",
			Comments:         nil,
			Time:             "2022-05-02T18:32:00+03:00",
			ID:               1,
			Score:            0,
			Text:             "test1",
			Title:            "test1",
			Type:             "text",
			UpvotePercentage: 100,
			Views:            0,
			Votes:            nil,
		},
		{
			Author:           *users[1],
			Category:         "funny",
			Comments:         nil,
			Time:             "2022-05-02T18:33:00+03:00",
			ID:               2,
			Score:            0,
			Text:             "test2",
			Title:            "test2",
			Type:             "text",
			UpvotePercentage: 0,
			Views:            0,
			Votes:            nil,
		},
		{
			Author:           *users[2],
			Category:         "music",
			Comments:         nil,
			Time:             "2022-05-02T18:34:00+03:00",
			ID:               3,
			Score:            0,
			Text:             "test3",
			Title:            "test3",
			Type:             "text",
			UpvotePercentage: 0,
			Views:            0,
			Votes:            nil,
		},
		{
			Author:           *users[3],
			Category:         "music",
			Comments:         nil,
			Time:             "2022-05-02T18:35:00+03:00",
			ID:               4,
			Score:            0,
			Text:             "test4",
			Title:            "test4",
			Type:             "text",
			UpvotePercentage: 0,
			Views:            0,
			Votes:            nil,
		},
	}

	dbUsers *DatabaseUser
)

func setupMongo() *PostRepoStruct {
	return &PostRepoStruct{
		users:      dbUsers,
		data:       &mocks.DatabasePost{},
		mx:         &sync.Mutex{},
		postsMuxes: make(map[uint64]*sync.Mutex),
		idGetter:   0,
	}
}

func TestPostLock(t *testing.T) {
	postRepo := setupMongo()
	postRepo.postsMuxes[0] = &sync.Mutex{}
	postRepo.postsMuxes[1] = &sync.Mutex{}

	postRepo.data.(*mocks.DatabasePost).
		On("Find", uint64(0), mock.AnythingOfType("*post.Post")).
		Return(nil)
	postRepo.data.(*mocks.DatabasePost).
		On("Find", uint64(1), mock.AnythingOfType("*post.Post")).
		Return(fmt.Errorf("error not found"))

	res := postRepo.Lock(0)
	require.True(t, res)
	res = postRepo.Lock(1)
	require.False(t, res)
	res = postRepo.Lock(2)
	require.False(t, res)
}

func TestPostUnlock(t *testing.T) {
	postRepo := setupMongo()
	postRepo.postsMuxes[0] = &sync.Mutex{}
	postRepo.postsMuxes[0].Lock()

	res := postRepo.Unlock(0)
	require.True(t, res)
	res = postRepo.Unlock(1)
	require.False(t, res)
}

func TestPostAdd(t *testing.T) {
	postRepo := setupMongo()

	postRepo.data.(*mocks.DatabasePost).
		On("Insert", mock.AnythingOfType("post.Post")).
		Return(nil)

	for i, pst := range posts {
		err := postRepo.Add(pst)
		require.NoError(t, err)
		require.Equal(t, uint64(i+1), pst.ID, "wrong post id")
	}
}

func TestPostAddErrors(t *testing.T) {
	postRepo := setupMongo()
	failPost := &post.Post{Title: "fail post"}

	postRepo.data.(*mocks.DatabasePost).
		On("Insert", mock.AnythingOfType("post.Post")).
		Return(fmt.Errorf("test error"))

	err := postRepo.Add(nil)
	require.Error(t, err, "add nil pointer")
	errFail := postRepo.Add(failPost)
	require.Errorf(t, errFail, "add fail post, has id %d", failPost.ID)
}

func TestPostFind(t *testing.T) {
	postRepo := setupMongo()

	postRepo.data.(*mocks.DatabasePost).
		On("Find", uint64(0), mock.AnythingOfType("*post.Post")).
		Return(nil)
	postRepo.data.(*mocks.DatabasePost).
		On("Find", uint64(1), mock.AnythingOfType("*post.Post")).
		Return(fmt.Errorf("not found error"))

	res := postRepo.Find(0)
	require.True(t, res)
	res = postRepo.Find(1)
	require.False(t, res)
}

func TestPostGet(t *testing.T) {
	postRepo := setupMongo()

	postRepo.data.(*mocks.DatabasePost).
		On("Find", uint64(0), mock.AnythingOfType("*post.Post")).
		Run(func(args mock.Arguments) {
			pst := args.Get(1).(*post.Post)
			*pst = *posts[0]
		}).
		Return(nil)
	postRepo.data.(*mocks.DatabasePost).
		On("Find", uint64(1), mock.AnythingOfType("*post.Post")).
		Return(fmt.Errorf("not found error"))

	res, err := postRepo.Get(0)
	require.NoError(t, err)
	require.Equal(t, res, *posts[0])

	res, err = postRepo.Get(1)
	require.Error(t, err)
	require.Equal(t, res, post.Post{})
}

func TestPostUpdate(t *testing.T) {
	postRepo := setupMongo()

	postRepo.data.(*mocks.DatabasePost).
		On("Replace", *posts[0]).
		Return(nil)
	postRepo.data.(*mocks.DatabasePost).
		On("Replace", *posts[1]).
		Return(fmt.Errorf("not found error"))

	err := postRepo.Update(*posts[0])
	require.NoError(t, err)

	err = postRepo.Update(*posts[1])
	require.Error(t, err)
}

func TestPostRemove(t *testing.T) {
	postRepo := setupMongo()
	postRepo.postsMuxes[0] = &sync.Mutex{}

	postRepo.data.(*mocks.DatabasePost).
		On("Delete", uint64(0)).
		Return(nil)
	postRepo.data.(*mocks.DatabasePost).
		On("Delete", uint64(1)).
		Return(fmt.Errorf("not found error"))

	res := postRepo.Remove(0)
	require.True(t, res)
	res = postRepo.Remove(1)
	require.False(t, res)
}

func TestPostToJson(t *testing.T) {
	type testCase struct {
		category   string
		author     user.User
		resultDb   []*post.Post
		errDb      error
		resultJson []byte
		errJson    error
	}

	dbUserSql, mock, err := sqlmock.New()
	require.NoError(t, err, "can`t connect to mysql")

	dbUser := &DatabaseUser{
		database: dbUserSql,
	}

	postRepo := setupMongo()
	postRepo.users = dbUser

	for _, usr := range users {
		rows := sqlmock.NewRows([]string{"username", "password", "user_id"})
		rows.AddRow(usr.Username, usr.PasswordHash, usr.UserID)
		mock.
			ExpectQuery("SELECT username, password, user_id FROM users WHERE").
			WithArgs(usr.Username).
			WillReturnRows(rows)
	}
	rows := sqlmock.NewRows([]string{"username", "password", "user_id"})
	rows.AddRow(1, 1, 1)
	mock.
		ExpectQuery("SELECT username, password, user_id FROM users WHERE").
		WithArgs("test username").
		WillReturnRows(rows)

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"score": -1})

	case1json, errMarshal := json.Marshal(posts[0])
	require.NoError(t, errMarshal)
	case1res := []byte("[" + string(case1json) + "]")

	testCases := []*testCase{
		{
			"music",
			user.User{
				Username: users[0].Username,
				UserID:   users[0].UserID,
			},
			[]*post.Post{posts[0]},
			nil,
			case1res,
			nil,
		},
		{
			"music",
			user.User{
				Username: "test username",
				UserID:   -1,
			},
			[]*post.Post{},
			nil,
			[]byte("[]"),
			nil,
		},
		{
			"music",
			user.User{
				Username: users[1].Username,
				UserID:   users[1].UserID,
			},
			nil,
			fmt.Errorf("strange error"),
			nil,
			fmt.Errorf("strange error"),
		},
	}

	for _, testCase := range testCases {
		filter := bson.M{
			"category": testCase.category,
			"author":   testCase.author,
		}
		postRepo.data.(*mocks.DatabasePost).
			On("GetAll", filter, findOptions).
			Return(testCase.resultDb, testCase.errDb)
		res, err := postRepo.ToJson("music", testCase.author.Username)

		require.Equal(t, err, testCase.errJson)
		require.Equal(t, res, testCase.resultJson)
	}
}

func TestPostRepoInit(t *testing.T) {
	type testCase struct {
		getAll []*post.Post
		getErr error
		result *PostRepoStruct
		resErr error
	}

	testCase1map := map[uint64]*sync.Mutex{}
	for _, pst := range posts {
		testCase1map[pst.ID] = &sync.Mutex{}
	}

	testCases := []testCase{
		{
			posts,
			nil,
			&PostRepoStruct{
				users:      nil,
				data:       nil,
				mx:         &sync.Mutex{},
				postsMuxes: testCase1map,
				idGetter:   uint64(len(posts)),
			},
			nil,
		},
		{
			nil,
			fmt.Errorf("test error"),
			nil,
			fmt.Errorf("test error"),
		},
	}

	for _, testCase := range testCases {
		postRepo := setupMongo()
		postRepo.data.(*mocks.DatabasePost).
			On("GetAll", bson.M{}, options.Find()).
			Return(testCase.getAll, testCase.getErr)
		res, err := NewPostRepo(postRepo.data, nil)

		require.Equal(t, err, testCase.resErr)
		if res == nil || testCase.result == nil {
			require.Equal(t, res, testCase.result)
		} else {
			require.Equal(t, res.data, postRepo.data)
			require.Equal(t, res.users, testCase.result.users)
			require.Equal(t, res.idGetter, testCase.result.idGetter)
		}
	}
}
