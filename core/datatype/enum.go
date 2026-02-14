// Package datatype provides generic utilities for common data types.
package datatype

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
)

type StringEnum interface {
	~string
	IsValid() bool
}

func FromString[E StringEnum](value string) (E, error) {
	if e := E(value); e.IsValid() {
		return e, nil
	}
	return E(""), fmt.Errorf("invalid value '%v' for StringEnum %T", value, *new(E))
}

type DBStringEnum interface {
	IsValid() bool
	driver.Valuer
	sql.Scanner
}

func ValueStringEnum[E StringEnum](e E) (driver.Value, error) {
	if !e.IsValid() {
		return nil, fmt.Errorf("StringEnum %T with value '%v' is invalid", e, e)
	}
	return string(e), nil
}

func ScanStringEnum[E StringEnum](ptr *E, value any) error {
	switch v := value.(type) {
	case string:
		e, err := FromString[E](v)
		if err != nil {
			return err
		}
		*ptr = e
		return nil
	default:
		return fmt.Errorf("invalid value '%v' for StringEnum %T", value, *new(E))
	}
}

type NullStringEnum[E StringEnum] struct {
	Enum  E
	Valid bool
}

func (n NullStringEnum[E]) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return ValueStringEnum(n.Enum)
}

func (n *NullStringEnum[E]) Scan(value any) error {
	if value == nil {
		n.Valid = false
		return nil
	}

	if err := ScanStringEnum(&n.Enum, value); err != nil {
		n.Valid = false
		return err
	}

	n.Valid = true
	return nil
}
