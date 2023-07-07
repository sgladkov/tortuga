package models

import "time"

type Bid struct {
	Id       uint64
	Project  uint64
	User     string
	Price    uint64
	Deadline time.Duration
	Message  string
}
