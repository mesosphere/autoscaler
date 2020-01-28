package errors

import (
	"fmt"
)

// IsNotInTableError determines whether a given error is of the NotInTableError type
func IsNotInTableError(err error) bool {
	if _, ok := err.(NotInTableError); ok {
		return true
	}
	return false
}

// NotInTableError is an error which occurs when an addon searched for in an AddonTable is not found
type NotInTableError struct {
	Info string

	name    string
	version string
}

func (e NotInTableError) Error() string {
	if e.Info != "" {
		return e.Info
	}
	return fmt.Sprintf("addon %s version %s not found", e.name, e.version)
}
