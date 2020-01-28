package test

import (
	"fmt"
	"os"
	"time"
)

var (
	defaultSetupWaitDuration = time.Minute * 15
	defaultAddonWaitDuration = time.Minute / 4
)

func init() {
	if v := os.Getenv("ADDON_TESTS_SETUP_WAIT_DURATION"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s was not a valid duration: %s", v, err)
		}
		defaultSetupWaitDuration = d
	}

	if v := os.Getenv("ADDON_TESTS_PER_ADDON_WAIT_DURATION"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s was not a valid duration: %s", v, err)
		}
		defaultAddonWaitDuration = d
	}
}
