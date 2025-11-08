package reader

import (
	"errors"
	"io"
	"net/http"
	"time"
)

func NewHTTPReader(source string) (*HTTPReader, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	if !isUrlExists(httpClient, source) {
		return nil, errors.New("url not exists")
	}

	return &HTTPReader{
		src:    source,
		client: httpClient,
	}, nil
}

type HTTPReader struct {
	src    string
	client *http.Client
	body   io.ReadCloser
}

func (r *HTTPReader) Read(p []byte) (int, error) {
	if r.body == nil {
		resp, err := r.client.Get(r.src)
		if err != nil {
			return 0, err
		}
		r.body = resp.Body
	}

	return r.body.Read(p)
}

func isUrlExists(client *http.Client, url string) bool {
	resp, err := client.Head(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}
