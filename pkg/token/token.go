package token

import (
	"fmt"
	"math/rand"
	"redditclone/pkg/errors"
	"redditclone/pkg/user"
	"strconv"

	"github.com/dgrijalva/jwt-go"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

const runeLen = 20

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GetToken(usr user.User, secretKey string) (res string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": map[string]interface{}{
			"username": usr.Username,
			"id":       fmt.Sprint(usr.UserID),
		},
	})
	res, err = token.SignedString([]byte(secretKey))
	if err != nil {
		err = errors.ErrSignToken{Err: err}
	}
	return
}

func GetTokenString(mp map[string]interface{}, key string) (str string, err error) {
	if username, ok := mp[key]; !ok {
		err = errors.ErrBadToken{Err: fmt.Errorf(`no key "%s"`, key)}
		return
	} else {
		if str, ok = username.(string); !ok {
			err = errors.ErrBadToken{Err: fmt.Errorf(`invalid type of "%s", should be "string"`, key)}
			return
		}
	}
	return
}

func GetTokenUint64(mp map[string]interface{}, key string) (num uint64, err error) {
	numStr, err := GetTokenString(mp, key)
	if err != nil {
		return
	}
	num, err = strconv.ParseUint(numStr, 10, 64)
	if err != nil {
		err = errors.ErrBadToken{Err: fmt.Errorf("can't parse num: %w", err)}
	}
	return
}

func GetTokenInt64(mp map[string]interface{}, key string) (num int64, err error) {
	numStr, err := GetTokenString(mp, key)
	if err != nil {
		return
	}
	num, err = strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		err = errors.ErrBadToken{Err: fmt.Errorf("can't parse num: %w", err)}
	}
	return
}

func GetMapItemString(mp map[string]string, key string) (string, error) {
	if strMap, ok := mp[key]; !ok {
		return "", fmt.Errorf(`no keyworld "%s"`, key)
	} else {
		return strMap, nil
	}
}

func GetMapItemUint64(mp map[string]string, key string) (num uint64, err error) {
	if mp == nil {
		err = fmt.Errorf("empty map")
		return
	}
	numStr, errGet := GetMapItemString(mp, key)
	if errGet != nil {
		err = errGet
		return
	}
	idParse, errConv := strconv.ParseUint(numStr, 10, 64)
	if errConv != nil {
		err = fmt.Errorf("can`t convert str to uint64: %w", errConv)
		return
	}
	num = idParse
	return
}

func CheckToken(tokenStr, secretKey string) (usr user.User, err error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			return nil, errors.ErrBadToken{Err: fmt.Errorf("bad sign method")}
		}
		return []byte(secretKey), nil
	})
	if err != nil || !token.Valid {
		if err != nil {
			err = errors.ErrBadToken{Err: fmt.Errorf("invalid token: %w", err)}
		} else {
			err = errors.ErrBadToken{Err: fmt.Errorf("invalid token")}
		}
		return
	}
	payload, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = errors.ErrBadToken{Err: fmt.Errorf("no payload")}
		return
	}

	// if runeStr, ok := payload["rune"].(string); !ok || len(runeStr) != runeLen {
	// 	err = errors.ErrBadToken{Err: fmt.Errorf("wrong token")}
	// 	return
	// } else {
	// 	for _, c := range runeStr {
	// 		flag := true
	// 		for _, k := range letterRunes {
	// 			if c == k {
	// 				flag = false
	// 				break
	// 			}
	// 		}
	// 		if flag {
	// 			err = errors.ErrBadToken{Err: fmt.Errorf("wrong token")}
	// 			return
	// 		}
	// 	}
	// }

	userItems, ok := payload["user"]
	if !ok {
		err = errors.ErrBadToken{Err: fmt.Errorf("wrong token key user")}
		return
	}
	userMap, ok := userItems.(map[string]interface{})
	if !ok {
		err = errors.ErrBadToken{Err: fmt.Errorf("wrong token key user: should be map[string]interface{}")}
		return
	}
	usr.Username, err = GetTokenString(userMap, "username")
	if err != nil {
		return
	}
	usr.UserID, err = GetTokenInt64(userMap, "id")
	return
}

func RemoveInArr[T any](arr []T, index uint) []T {
	if len(arr) == 0 || index >= uint(len(arr)) {
		return arr
	}
	switch index {
	case 0:
		arr = arr[1:]
	case uint(len(arr)) - 1:
		arr = arr[:len(arr)-1]
	default:
		arr = append(arr[:index], arr[index+1:]...)
	}
	return arr
}
