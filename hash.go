package main

import (
	"crypto/sha1"
	"io"
	"log"
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
	buf := make([]byte, 1*1024*1024)
	if _, err := io.CopyBuffer(h, f, buf); err != nil {
		return nil, err
	}

	sum := h.Sum(nil)
	if verbose {
		log.Printf("%x  %v\n", sum, path)
	}
	return sum, nil
}
