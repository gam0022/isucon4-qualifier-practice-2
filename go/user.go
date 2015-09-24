package main

import (
)

type User struct {
	ID           int
	Login        string
	PasswordHash string
	Salt         string
}

type LastLogin struct {
	Login     string
	IP        string
	CreatedAt string
}

func GetLastLogin(userID int) *LastLogin {
	pair := LastLoginHistory[userID]
	if pair[1].IP != "" {
		return &pair[1]
	} else {
		return &pair[0]
	}
}
