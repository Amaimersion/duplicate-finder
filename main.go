package main

import (
	"bufio"
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var logger = log.New(os.Stdout, "", 0)

func main() {
	cfg := parseFlags()

	if err := cfg.check(); err != nil {
		logger.Fatal(err)
	}

	// We will store data in file, not in memory, because
	// amount of data can be very large.
	tempFile, err := os.CreateTemp("", "duplicate-finder-")

	if err != nil {
		logger.Fatalf("unable to create temp file: %v", err)
	}

	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	err = walkFolder(cfg.folder1, func(filePath, _ string) {
		hash, err := fileToMD5(filePath)

		if err != nil {
			logger.Printf("unable to hash %v: %v", filePath, err)
			return
		}

		i := info{
			path: filePath,
			hash: hash,
		}
		s := i.string() + "\n"
		_, err = tempFile.Write([]byte(s))

		if err != nil {
			logger.Printf("unable to write info %v: %v", filePath, err)
			return
		}
	})

	if err != nil {
		logger.Fatal(err)
	}

	err = walkFolder(cfg.folder2, func(filePath, fileName string) {
		hash, err := fileToMD5(filePath)

		if err != nil {
			logger.Printf("unable to hash %v: %v", filePath, err)
			return
		}

		_, err = tempFile.Seek(0, 0)

		if err != nil {
			logger.Printf("unable to seek: %v", err)
			return
		}

		scanner := bufio.NewScanner(tempFile)

		for scanner.Scan() {
			original := info{}

			if err = original.fromString(scanner.Text()); err != nil {
				logger.Println(err)
				continue
			}

			if original.path == filePath {
				continue
			}

			if original.hash != hash {
				continue
			}

			if len(cfg.move) > 0 {
				if err := move(filePath, cfg.move, fileName); err == nil {
					logger.Printf("moved: %v", fileName)
				} else {
					logger.Printf("unable to move %v: %v", fileName, err)
				}
			} else {
				logger.Printf(
					"duplicate - %v, original - %v",
					filePath,
					original.path,
				)
			}
		}

		if err := scanner.Err(); err != nil {
			logger.Printf("unable to scan: %v", err)
			return
		}
	})

	if err != nil {
		logger.Fatal(err)
	}
}

type config struct {
	folder1 string
	folder2 string
	move    string
}

func (c config) check() error {
	if len(c.folder1) == 0 || len(c.folder2) == 0 {
		return errors.New("f1 and f2 must be specified")
	}

	if c.folder1 == c.folder2 {
		return errors.New("same folder not supported, instead copy f1 to new folder")
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
	flag.StringVar(
		&cfg.move,
		"move",
		"",
		"Move duplicates to this folder.",
	)

	flag.Parse()

	return cfg
}

func walkFolder(path string, f func(filePath, fileName string)) error {
	fsys := os.DirFS(path)
	err := fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			logger.Printf("%v will be skipped: %v", name, err)
			return fs.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		fullPath := filepath.Join(path, name)

		f(fullPath, name)

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

func move(path, to, name string) error {
	newPath := filepath.Join(to, name)
	dir := strings.TrimSuffix(newPath, filepath.Base(newPath))

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	if err := os.Rename(path, newPath); err != nil {
		return err
	}

	return nil
}

type info struct {
	path string
	hash string
}

func (i info) string() string {
	return fmt.Sprintf("%v %v", i.hash, i.path)
}

func (i *info) fromString(s string) error {
	parts := strings.SplitN(s, " ", 2)

	if len(parts) != 2 {
		return errors.New("invalid info string")
	}

	i.hash = parts[0]
	i.path = parts[1]

	return nil
}
