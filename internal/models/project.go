package models

import "time"

type ProjectStatus uint8

const (
	Open ProjectStatus = iota
	InWork
	InReview
	Completed
	Canceled
)

type Project struct {
	Id          uint64
	Title       string
	Description string
	Tags        Tags
	Created     time.Time
	Status      ProjectStatus
	Owner       string
	Contractor  string
	Started     time.Time
	Deadline    time.Duration
	Price       uint64
}
