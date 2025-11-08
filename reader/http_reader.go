package reader

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	apperrors "abc/errors"
)

func NewHTTPReader(source string) (*HTTPReader, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	filename, totalSize, err := getUrlInfo(httpClient, source)
	if err != nil {
		return nil, err
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

		if resp.StatusCode != http.StatusOK {
			return 0, errors.New(apperrors.ERR_URL_NOT_EXISTS)
		}

		ctype := resp.Header.Get("Content-Type")
		fmt.Println("ctype", ctype)
		if strings.Contains(ctype, "application/gzip") || strings.Contains(ctype, "application/x-gzip") {
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

func getUrlInfo(client *http.Client, url string) (string, int64, error) {
	resp, err := client.Head(url)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", 0, errors.New(apperrors.ERR_URL_NOT_EXISTS)
	}

	filename := resp.Header.Get("Content-Disposition")
	filename = strings.TrimPrefix(filename, "attachment; filename=")
	if filename == "" {
		filename = filepath.Base(url)
	}

	fileSize := resp.ContentLength
	return filename, fileSize, nil
}
