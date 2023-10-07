package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)

type Tags []string

// Value implements driver.Valuer
func (tags Tags) Value() (driver.Value, error) {
	if len(tags) == 0 {
		return "", nil
	}
	return strings.Join(tags, "|"), nil
}

// Scan implements sql.Scanner
func (tags *Tags) Scan(value interface{}) error {
	if value == nil {
		*tags = Tags{}
		return nil
	}

	sv, err := driver.String.ConvertValue(value)
	if err != nil {
		return fmt.Errorf("cannot scan value. %w", err)
	}

	v, ok := sv.(string)
	if !ok {
		return errors.New("cannot scan value. cannot convert value to string")
	}
	*tags = strings.Split(v, "|")

	return nil
}

func (t Tags) Equal(t2 Tags) bool {
	if len(t) != len(t2) {
		return false
	}
	for i := range t {
		if t[i] != t2[i] {
			return false
		}
	}
	return true
}
