package main

import (
	"errors"
	"flag"
	"log"
	"os"
)

var logger = log.New(
	os.Stdout,
	"",
	0,
)

func main() {
	cfg := parseFlags()

	if err := cfg.check(); err != nil {
		logger.Fatal(err)
	}
}

type config struct {
	folder1 string
	folder2 string
}

func (c config) check() error {
	if len(c.folder1) == 0 || len(c.folder2) == 0 {
		return errors.New("f1 and f2 must be specified")
	}

	return nil
}

func parseFlags() config {
	cfg := config{}

	flag.StringVar(
		&cfg.folder1,
		"f1",
		"",
		"Path to the folder № 1.",
	)
	flag.StringVar(
		&cfg.folder2,
		"f2",
		"",
		"Path to the folder № 2.",
	)

	flag.Parse()

	return cfg
}
