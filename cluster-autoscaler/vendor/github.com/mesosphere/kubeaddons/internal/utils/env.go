package utils

import (
	"os"
	"time"
)

// EnvDuration is a helper function to override a string with an ENV override
// that is a time.Duration.
func EnvDuration(env string, dest *time.Duration) error {
	if v := os.Getenv(env); v != "" {
		duration, err := time.ParseDuration(v)
		if err != nil {
			return err
		}
		*dest = duration
	}
	return nil
}
