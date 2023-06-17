package models

import "time"

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
