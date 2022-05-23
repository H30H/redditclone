package user

import (
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"strconv"
)

type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"-" bson:"-"`
	UserID       int64  `json:"id,string"`
}

func Md5Hash(data string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

func Crc32Hash(data string) string {
	crcH := crc32.ChecksumIEEE([]byte(data))
	dataHash := strconv.FormatUint(uint64(crcH), 10)
	return dataHash
}

func GetPasswordHash(password string) string {
	return Crc32Hash(Md5Hash(password+"heh")) + "b" + Md5Hash(password)[:len(password)]
}
