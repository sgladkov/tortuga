package models

import (
	"database/sql/driver"
	"errors"
	"time"
)

type FixedTime time.Time

// Value implements driver.Valuer
func (ft FixedTime) Value() (driver.Value, error) {
	return driver.Value(time.Time(ft).UnixMilli()), nil
}

// Scan implements sql.Scanner
func (ft *FixedTime) Scan(value interface{}) error {
	if value == nil {
		*ft = FixedTime(time.Time{})
		return nil
	}

	v, ok := value.(int64)
	if !ok {
		return errors.New("cannot scan value. value is not int64")
	}
	*ft = FixedTime(time.UnixMilli(v))

	return nil
}
