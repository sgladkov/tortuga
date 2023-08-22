package models

type Rate struct {
	Id        uint64
	Evaluator string
	Project   uint64
	Evaluated string
	Rate      uint8
	Message   string
}
