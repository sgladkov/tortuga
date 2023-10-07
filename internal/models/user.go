package models

import (
	"time"
)

type User struct {
	Id          string
	Nickname    string
	Description string
	Nonce       uint64
	Registered  time.Time
	Status      uint8
	Tags        Tags
	Rating      float64
	Account     uint64
}

func (u User) Equal(u2 User) bool {
	if u.Id != u2.Id {
		return false
	}
	if u.Nickname != u2.Nickname {
		return false
	}
	if u.Description != u2.Description {
		return false
	}
	if u.Nonce != u2.Nonce {
		return false
	}
	if u.Registered.Round(time.Second) != u2.Registered.Round(time.Second) {
		return false
	}
	if u.Status != u2.Status {
		return false
	}
	if !u.Tags.Equal(u2.Tags) {
		return false
	}
	if u.Rating != u2.Rating {
		return false
	}
	if u.Account != u2.Account {
		return false
	}
	return true
}