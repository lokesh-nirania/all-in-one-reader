package reader

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	apperrors "abc/errors"

	"github.com/google/uuid"
)

const (
	SCHEME_HTTP  = "http"
	SCHEME_HTTPS = "https"
	SCHEME_FILE  = "file"

	SCHEME_SUFFIX = "://"

	SCHEME_HTTP_PREFIX  = SCHEME_HTTP + SCHEME_SUFFIX
	SCHEME_HTTPS_PREFIX = SCHEME_HTTPS + SCHEME_SUFFIX
	SCHEME_FILE_PREFIX  = SCHEME_FILE + SCHEME_SUFFIX

	PART_FILE_SUFFIX = ".part"
)

func NewReader(source string) (*Reader, error) {
	if strings.HasPrefix(source, SCHEME_FILE_PREFIX) {
		path := strings.TrimPrefix(source, SCHEME_FILE_PREFIX)
		fileReader, err := NewFileReader(path)
		if err != nil {
			return nil, err
		}

		return &Reader{fileReader}, nil
	}

	if strings.HasPrefix(source, SCHEME_HTTP_PREFIX) || strings.HasPrefix(source, SCHEME_HTTPS_PREFIX) {
		httpReader, err := NewHTTPReader(source)
		if err != nil {
			return nil, err
		}

		return &Reader{httpReader}, nil
	}

	return nil, errors.New(apperrors.ERR_UNSUPPORTED_SCHEME)
}

type SourceReader interface {
	io.ReadCloser
	Filename() string
	TotalSize() int64
}

type Reader struct {
	src SourceReader
}

func (r *Reader) Read(p []byte) (int, error) {
	if r.src == nil {
		return 0, io.EOF
	}
	return r.src.Read(p)
}

func (r *Reader) Close() error {
	if r.src == nil {
		return nil
	}
	return r.src.Close()
}

func (r *Reader) StreamToFile(destinationFolder string) (string, int64, error) {
	if r.src == nil {
		return "", 0, errors.New(apperrors.ERR_READER_SOURCE_NIL)
	}

	tempPath := filepath.Join(destinationFolder, uuid.New().String()+PART_FILE_SUFFIX)
	out, err := os.Create(tempPath)
	if err != nil {
		return "", 0, err
	}
	defer out.Close()

	pr := &ProgressReader{
		Reader:    r.src,
		TotalSize: r.src.TotalSize(),
		Notify:    NotifyProgress,
	}

	n, err := io.Copy(out, pr)
	if err != nil && err != io.EOF {
		return tempPath, n, err
	}

	finalName := r.src.Filename()
	finalPath := filepath.Join(destinationFolder, finalName)

	finalPath, err = GetUniqueFilePath(finalPath)
	if err != nil {
		return tempPath, n, err
	}

	if renameErr := os.Rename(tempPath, finalPath); renameErr != nil {
		return tempPath, n, renameErr
	}

	return finalPath, n, nil
}
