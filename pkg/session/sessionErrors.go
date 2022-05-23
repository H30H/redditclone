package session

type ErrorTokenNotFound struct {
}

func (t ErrorTokenNotFound) Error() string {
	return "token not found in database"
}

type ErrorTokenIsExpired struct {
}

func (t ErrorTokenIsExpired) Error() string {
	return "token is expired"
}
