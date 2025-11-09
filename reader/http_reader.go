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

	"abc/cache"
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

	cacheMgr, err := cache.NewManager(".cache")
	if err != nil {
		return nil, err
	}

	return &HTTPReader{
		src:       source,
		filename:  filename,
		client:    httpClientForGET,
		totalSize: totalSize,
		cacheMgr:  cacheMgr,
	}, nil
}

type HTTPReader struct {
	src       string
	client    *http.Client
	body      io.ReadCloser
	filename  string
	totalSize int64
	cacheMgr  *cache.Manager
	fromCache bool
}

func (r *HTTPReader) Filename() string {
	return r.filename
}

func (r *HTTPReader) TotalSize() int64 {
	return r.totalSize
}

func (r *HTTPReader) Read(p []byte) (int, error) {
	if r.body == nil {
		var cached *cache.Entry
		if r.cacheMgr != nil {
			if e, ok := r.cacheMgr.Get(r.src); ok && e.Completed {
				cached = e
			}
		}

		if cached != nil {
			req, err := http.NewRequest(http.MethodGet, r.src, nil)
			if err != nil {
				return 0, err
			}
			if cached.ETag != "" {
				req.Header.Set("If-None-Match", cached.ETag)
			}
			if cached.LastModified != "" {
				req.Header.Set("If-Modified-Since", cached.LastModified)
			}
			resp, err := r.client.Do(req)
			if err != nil {
				return 0, err
			}
			if resp.StatusCode == http.StatusNotModified {
				f, err := r.cacheMgr.OpenFile(cached)
				if err != nil {
					resp.Body.Close()
					return 0, err
				}
				r.filename = cached.Filename
				if cached.Size > 0 {
					r.totalSize = cached.Size
				}
				r.body = f
				r.fromCache = true
				resp.Body.Close()
			} else {
				if resp.StatusCode != http.StatusOK {
					resp.Body.Close()
					return 0, errors.New(apperrors.ERR_URL_NOT_EXISTS)
				}
				var tempPath string
				var cacheWriter io.WriteCloser
				if r.cacheMgr != nil {
					p, w, err := r.cacheMgr.CreateTempWriter(r.src)
					if err == nil {
						tempPath = p
						cacheWriter = w
					}
				}
				ctype := resp.Header.Get("Content-Type")
				fmt.Println("ctype", ctype)
				var readerBody io.ReadCloser = resp.Body
				if strings.Contains(ctype, "application/gzip") || strings.Contains(ctype, "application/x-gzip") {
					gz, err := gzip.NewReader(resp.Body)
					if err != nil {
						resp.Body.Close()
						if cacheWriter != nil {
							cacheWriter.Close()
						}
						return 0, err
					}
					if gz.Name != "" {
						r.filename = gz.Name
					}
					readerBody = gz
				}
				if cacheWriter != nil {
					etag := resp.Header.Get("ETag")
					lastModified := resp.Header.Get("Last-Modified")
					r.body = newTeeReadCloser(readerBody, cacheWriter, func(written int64) {
						_, _ = r.cacheMgr.Commit(r.src, tempPath, r.filename, etag, lastModified, written)
					})
				} else {
					r.body = readerBody
				}
			}
		}

		if r.body == nil {
			resp, err := r.client.Get(r.src)
			if err != nil {
				return 0, err
			}
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				return 0, errors.New(apperrors.ERR_URL_NOT_EXISTS)
			}
			var tempPath string
			var cacheWriter io.WriteCloser
			if r.cacheMgr != nil {
				p, w, err := r.cacheMgr.CreateTempWriter(r.src)
				if err == nil {
					tempPath = p
					cacheWriter = w
				}
			}
			ctype := resp.Header.Get("Content-Type")
			fmt.Println("ctype", ctype)
			var readerBody io.ReadCloser = resp.Body
			if strings.Contains(ctype, "application/gzip") || strings.Contains(ctype, "application/x-gzip") {
				gz, err := gzip.NewReader(resp.Body)
				if err != nil {
					resp.Body.Close()
					if cacheWriter != nil {
						cacheWriter.Close()
					}
					return 0, err
				}
				if gz.Name != "" {
					r.filename = gz.Name
				}
				readerBody = gz
			}
			if cacheWriter != nil {
				etag := resp.Header.Get("ETag")
				lastModified := resp.Header.Get("Last-Modified")
				r.body = newTeeReadCloser(readerBody, cacheWriter, func(written int64) {
					_, _ = r.cacheMgr.Commit(r.src, tempPath, r.filename, etag, lastModified, written)
				})
			} else {
				r.body = readerBody
			}
		}
	}
	return r.body.Read(p)
}

func (r *HTTPReader) Close() error {
	if r.body != nil {
		err := r.body.Close()
		r.body = nil
		return err
	}
	return nil
}

func (r *HTTPReader) IsServingFromCache() bool {
	return r.fromCache
}

func getUrlInfo(client *http.Client, url string) (string, int64, error) {
	resp, err := client.Head(url)
	if err != nil {
		resp, err = client.Get(url)
		if err != nil {
			return "", 0, err
		}
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

type teeReadCloser struct {
	src     io.ReadCloser
	tee     io.WriteCloser
	written int64
	onClose func(written int64)
}

func newTeeReadCloser(src io.ReadCloser, tee io.WriteCloser, onClose func(written int64)) io.ReadCloser {
	return &teeReadCloser{
		src:     src,
		tee:     tee,
		onClose: onClose,
	}
}

func (t *teeReadCloser) Read(p []byte) (int, error) {
	n, err := t.src.Read(p)
	if n > 0 && t.tee != nil {
		if wn, werr := t.tee.Write(p[:n]); werr == nil {
			t.written += int64(wn)
		}
	}
	return n, err
}

func (t *teeReadCloser) Close() error {
	var err error
	if t.src != nil {
		err = t.src.Close()
	}
	if t.tee != nil {
		_ = t.tee.Close()
	}
	if t.onClose != nil {
		t.onClose(t.written)
	}
	return err
}
