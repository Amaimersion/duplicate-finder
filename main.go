package main

import (
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
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

	err := walkFolder(cfg.folder1, func(filePath string) {
		hash, err := fileToMD5(filePath)

		if err != nil {
			logger.Printf("unable to hash %v: %v", filePath, err)
			return
		}

		logger.Printf("%v: %v", hash, filePath)
	})

	if err != nil {
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

func walkFolder(path string, f func(filePath string)) error {
	fsys := os.DirFS(path)
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			logger.Printf("%v will be skipped: %v", path, err)
			return fs.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		fullPath := filepath.Join(path, p)

		f(fullPath)

		return nil
	})

	return err
}

func fileToMD5(path string) (string, error) {
	f, err := os.Open(path)

	if err != nil {
		return "", err
	}

	defer f.Close()

	h := md5.New()

	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	b16 := fmt.Sprintf("%x", h.Sum(nil))

	return b16, nil
}
