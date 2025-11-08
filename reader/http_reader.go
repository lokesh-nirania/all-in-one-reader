package reader

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

func NewHTTPReader(source string) (*HTTPReader, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	filename, totalSize, exists := isUrlExists(httpClient, source)
	if !exists {
		return nil, errors.New("url not exists")
	}

	httpClientForGET := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: 15 * time.Second,
			IdleConnTimeout:       90 * time.Second,
			DisableKeepAlives:     false,
		},
	}

	return &HTTPReader{
		src:       source,
		filename:  filename,
		client:    httpClientForGET,
		totalSize: totalSize,
	}, nil
}

type HTTPReader struct {
	src       string
	client    *http.Client
	body      io.ReadCloser
	filename  string
	totalSize int64
}

func (r *HTTPReader) Filename() string {
	return r.filename
}

func (r *HTTPReader) TotalSize() int64 {
	return r.totalSize
}

func (r *HTTPReader) Read(p []byte) (int, error) {
	if r.body == nil {
		resp, err := r.client.Get(r.src)
		if err != nil {
			return 0, err
		}
		ctype := resp.Header.Get("Content-Type")
		if strings.Contains(ctype, "application/gzip") {
			gz, err := gzip.NewReader(resp.Body)
			if err != nil {
				resp.Body.Close()
				return 0, err
			}
			r.filename = gz.Name
			r.body = gz
		} else {
			r.body = resp.Body
		}
	}
	return r.body.Read(p)
}

func isUrlExists(client *http.Client, url string) (string, int64, bool) {
	resp, err := client.Head(url)
	if err != nil {
		return "", 0, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", 0, false
	}

	filename := resp.Header.Get("Content-Disposition")
	filename = strings.TrimPrefix(filename, "attachment; filename=")
	if filename == "" {
		filename = filepath.Base(url)
	}

	fileSize := resp.ContentLength
	return filename, fileSize, true
}
