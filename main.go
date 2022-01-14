package main

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
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

	err = walkFolder(cfg.folder1, func(filePath, fileName string) {
		hash, err := fileToMD5(filePath)

		if err != nil {
			logger.Printf("unable to hash %v: %v", filePath, err)
			return
		}

		i := info{
			path: filePath,
			name: fileName,
			hash: hash,
		}
		s := i.decode() + "\n"
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
		candidate := info{
			hash: hash,
			path: filePath,
			name: fileName,
		}

		for scanner.Scan() {
			t := scanner.Text()
			original := info{}

			if err := original.encode(t); err != nil {
				logger.Println(err)
				continue
			}

			if original.path == candidate.path {
				continue
			}

			if original.hash != candidate.hash {
				continue
			}

			print(original, candidate)

			if len(cfg.move) > 0 {
				if err := move(candidate.path, cfg.move, candidate.name); err == nil {
					logger.Printf("moved: %v", candidate.name)
				} else {
					logger.Printf("unable to move %v: %v", candidate.name, err)
				}
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

func print(original, duplicate info) {
	logger.Printf(
		"%v (f2) is duplicate of %v (f1)",
		duplicate.name,
		original.name,
	)
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
	name string
}

func (i info) decode() string {
	// We will use Base64 in order to guarantee that there will be
	// only N spaces because we will use it in encode().
	return fmt.Sprintf(
		"%v %v %v",
		i.hash,
		toBase64(i.path),
		toBase64(i.name),
	)
}

func (i *info) encode(s string) error {
	n := 3
	parts := strings.SplitN(s, " ", n)

	if len(parts) != n {
		return errors.New("invalid info string")
	}

	path, err := fromBase64(parts[1])

	if err != nil {
		return err
	}

	name, err := fromBase64(parts[2])

	if err != nil {
		return err
	}

	i.hash = parts[0]
	i.path = path
	i.name = name

	return nil
}

func toBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func fromBase64(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)

	if err != nil {
		return "", err
	}

	return string(b), err
}
