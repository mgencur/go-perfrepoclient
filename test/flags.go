package test

import (
	"flag"
)

// Flags holds the flags or defaults for PerfRepo client test suite
var Flags = initializeFlags()

// TestFlags holds the flags for PerfRepo client test suite
type TestFlags struct {
	URL  string // PerfRepo application URL
	User string // username for connecting to PerfRepo
	Pass string // password for connecting to PerfRepo
}

func initializeFlags() *TestFlags {
	var f TestFlags

	flag.StringVar(&f.URL, "url", "http://localhost:8080/testing-repo",
		"Provide the URL of the PerfRepo application")

	flag.StringVar(&f.User, "user", "perfrepouser",
		"Provide the username for connecting to PerfRepo")

	flag.StringVar(&f.Pass, "pass", "perfrepouser1.",
		"Provide the password for connecting to PerfRepo")

	flag.Parse()

	return &f
}
