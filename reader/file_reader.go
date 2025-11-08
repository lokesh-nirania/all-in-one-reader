package reader

import (
	"errors"
	"os"
	"path/filepath"

	apperrors "abc/errors"
)

func NewFileReader(source string) (*FileReader, error) {
	totalSize, exists := isFileExist(source)
	if !exists {
		return nil, errors.New(apperrors.ERR_FILE_NOT_FOUND)
	}
	filename := filepath.Base(source)
	return &FileReader{src: source, filename: filename, totalSize: totalSize}, nil
}

type FileReader struct {
	src       string
	filename  string
	file      *os.File
	totalSize int64
}

func (r *FileReader) Filename() string {
	return r.filename
}

func (r *FileReader) TotalSize() int64 {
	return r.totalSize
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

func isFileExist(path string) (int64, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, false
	}
	return info.Size(), true
}
