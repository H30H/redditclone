package database

import (
	"fmt"
	"redditclone/pkg/user"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestAddUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	// rows := sqlmock.NewRows([]string{"username", "password", "user_id"})
	users := []*user.User{
		{
			Username:     "test1",
			PasswordHash: "test1",
			UserID:       0,
		},
		{
			Username:     "test2",
			PasswordHash: "test2",
			UserID:       0,
		},
		{
			Username:     "test3",
			PasswordHash: "test3",
			UserID:       0,
		},
		{
			Username:     "test4",
			PasswordHash: "test4",
			UserID:       0,
		},
	}

	repo := UserRepoStruct{
		data: &DatabaseUser{
			database: db,
		},
		mx: &sync.Mutex{},
	}

	for i, usr := range users {
		id := int64(i + 1)
		mock.
			ExpectExec("INSERT INTO users").
			WithArgs(usr.Username, usr.PasswordHash).
			WillReturnResult(sqlmock.NewResult(id, 1))
		err := repo.Add(usr)
		if err != nil {
			t.Fatalf("has unexpected error: %s", err)
		}
		if usr.UserID != id {
			t.Fatalf("id didn`t change: has `userID` %d, but expected %d", usr.UserID, id)
		}
	}
}

func TestGetUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	users := []*user.User{
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

	repo := UserRepoStruct{
		data: &DatabaseUser{
			database: db,
		},
		mx: &sync.Mutex{},
	}

	for _, usr := range users {
		rows := sqlmock.NewRows([]string{"username", "password", "user_id"})
		rows.AddRow(usr.Username, usr.PasswordHash, usr.UserID)
		mock.
			ExpectQuery("SELECT username, password, user_id FROM users WHERE").
			WithArgs(usr.Username).
			WillReturnRows(rows)
		user, ok := repo.Find(usr.Username)
		if !ok {
			t.Fatalf("has unexpected error: user not found")
		}
		if !reflect.DeepEqual(user, *usr) {
			t.Fatalf("get wrong user: has %+v, but expected %+v", user, *usr)
		}
	}

	mock.
		ExpectQuery("SELECT username, password, user_id FROM users WHERE").
		WithArgs("new user").
		WillReturnError(fmt.Errorf("user not found"))
	_, ok := repo.Find("new user")
	if ok {
		t.Fatalf("has unexpected error: user not found")
	}
}

func TestInitUserRepo(t *testing.T) {
	res := NewUserRepo(nil)
	require.NotNil(t, res)
	require.Nil(t, res.data)
	require.NotNil(t, res.mx)
}
