package main

import (
	"abc/reader"
	"fmt"
	"io"
	"log"
)

func main() {
	r, err := reader.NewReader("file://test2.txt")
	// r, err := reader.NewReader("https://sample-files.com/downloads/documents/txt/simple.txt")
	if err != nil {
		log.Fatalf("failed to create reader: %s", err)
	}

	bytes, err := io.ReadAll(r)
	if err != nil {
		log.Fatalf("failed to read file: %s", err)
	}

	fmt.Println("Read bytes: ", string(bytes))
}
