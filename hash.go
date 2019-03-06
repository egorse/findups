package main

import (
	"crypto/sha1"
	"io"
	"os"
)

// Hash is function seen at https://godoc.org/crypto/sha1#example-New--File,
func Hash(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
