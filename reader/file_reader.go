package reader

import (
	"errors"
	"os"
)

func NewFileReader(source string) (*FileReader, error) {
	if !isFileExist(source) {
		return nil, errors.New("file not found")
	}
	return &FileReader{src: source}, nil
}

type FileReader struct {
	src  string
	file *os.File
}

func (r *FileReader) Read(p []byte) (int, error) {
	if r.file == nil {
		file, err := os.Open(r.src)
		if err != nil {
			return 0, err
		}
		r.file = file
	}
	return r.file.Read(p)
}

func isFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}
