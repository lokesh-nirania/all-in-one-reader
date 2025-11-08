package main

import (
	"abc/reader"
	"fmt"
	"log"
	"os"
)

func main() {
	downloadFolder := "downloads"
	os.MkdirAll(downloadFolder, 0755)

	// r, err := reader.NewReader("file://test.txt")
	// r, err := reader.NewReader("https://sample-files.com/downloads/documents/txt/simple.txt")
	// r, err := reader.NewReader("https://getsamplefiles.com/download/gzip/sample-1.gz")

	// r, err := reader.NewReader("file:///Users/lokesh.nirania/Downloads/ubuntu-25.04-desktop-arm64.iso")
	r, err := reader.NewReader("https://mirror.bharatdatacenter.com/ubuntu-releases/24.04.3/ubuntu-24.04.3-desktop-amd64.iso")
	if err != nil {
		log.Fatalf("failed to create reader: %s", err)
	}

	filePath, n, err := r.StreamToFile(downloadFolder)
	if err != nil {
		log.Fatalf("failed to stream to file: %s", err)
	}
	fmt.Printf("streamed to file: %s with size %d bytes\n", filePath, n)
}
