package main

import (
	"time"
)

type User struct {
	ID           int
	Login        string
	PasswordHash string
	Salt         string

	LastLogin *LastLogin
}

type LastLogin struct {
	Login     string
	IP        string
	CreatedAt time.Time
}

var LastLoginHistory = map[int][2]LastLogin{}

func (u *User) getLastLogin() *LastLogin {
	userID := u.ID
	if LastLoginHistory[userID] == nil {
		return nil
	} else if LastLoginHistory[userID][1] != nil {
		return LastLoginHistory[userID][1]
	} else {
		return LastLoginHistory[userID][0]
	}
}
