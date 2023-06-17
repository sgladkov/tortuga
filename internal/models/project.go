package models

import "time"

type Project struct {
	Id          uint64
	Title       string
	Description string
	Tags        Tags
	Created     time.Time
	Status      uint8
	Owner       string
	Contractor  string
	Started     time.Time
	Deadline    time.Duration
	Price       uint64
}
