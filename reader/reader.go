package reader

import (
	"errors"
	"io"
	"strings"
)

const (
	SCHEME_HTTP  = "http"
	SCHEME_HTTPS = "https"
	SCHEME_FILE  = "file"

	SCHEME_SUFFIX = "://"

	SCHEME_HTTP_PREFIX  = SCHEME_HTTP + SCHEME_SUFFIX
	SCHEME_HTTPS_PREFIX = SCHEME_HTTPS + SCHEME_SUFFIX
	SCHEME_FILE_PREFIX  = SCHEME_FILE + SCHEME_SUFFIX
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

	return nil, errors.New("unsupported scheme")
}

type Reader struct {
	src io.Reader
}

func (r *Reader) Read(p []byte) (int, error) {
	if r.src == nil {
		return 0, io.EOF
	}
	return r.src.Read(p)
}
