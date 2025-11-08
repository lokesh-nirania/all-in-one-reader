package reader

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	SLASH = "/"

	NETWORK_RESET_ON_HEAD_PATH = SLASH + "network-reset-on-head"
	NETWORK_RESET_ON_GET_PATH  = SLASH + "network-reset-on-get"

	HEAD_OK_FILE_NAME = "head-ok.txt"
	HEAD_OK_FILE_PATH = SLASH + HEAD_OK_FILE_NAME // HEAD will pass, GET will fail

	GET_OK_FILE_NAME = "get-ok.txt"
	GET_OK_FILE_PATH = SLASH + GET_OK_FILE_NAME // HEAD and GET will pass
	GET_FILE_CONTENT = "This is a simple text file for testing purposes."

	GET_OK_GZIP_FILE_NAME = "get-ok.gz"
	GET_OK_GZIP_FILE_PATH = SLASH + GET_OK_GZIP_FILE_NAME
	GZIP_FILE_CONTENT     = "This is a simple text file from gzip for testing purposes.."

	DOES_NOT_EXIST_FILE_PATH = SLASH + "does-not-exist.txt"
)

type HTTPReaderTestSuite struct {
	suite.Suite
	server *httptest.Server
}

func (s *HTTPReaderTestSuite) SetupTest() {
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headOKConditions := (r.Method == http.MethodHead && r.URL.Path == HEAD_OK_FILE_PATH) ||
			(r.Method == http.MethodHead && r.URL.Path == GET_OK_FILE_PATH) ||
			(r.Method == http.MethodHead && r.URL.Path == GET_OK_GZIP_FILE_PATH) ||
			(r.Method == http.MethodHead && r.URL.Path == NETWORK_RESET_ON_GET_PATH)

		hiJackConditions := (r.Method == http.MethodHead && r.URL.Path == NETWORK_RESET_ON_HEAD_PATH) ||
			(r.Method == http.MethodGet && r.URL.Path == NETWORK_RESET_ON_GET_PATH)

		getOKConditions := (r.Method == http.MethodGet && r.URL.Path == GET_OK_FILE_PATH)
		getOKGZIPConditions := (r.Method == http.MethodGet && r.URL.Path == GET_OK_GZIP_FILE_PATH)

		getNotAllowedConditions := (r.Method == http.MethodGet && r.URL.Path == HEAD_OK_FILE_PATH)

		switch {
		case headOKConditions:
			w.WriteHeader(http.StatusOK)

		case getNotAllowedConditions:
			http.Error(w, "GET not allowed", http.StatusForbidden)

		case getOKConditions:
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, GET_FILE_CONTENT)

		case getOKGZIPConditions:
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			io.WriteString(gz, GZIP_FILE_CONTENT)

		case hiJackConditions:
			hj, ok := w.(http.Hijacker)
			if !ok {
				http.Error(w, "cannot hijack connection", http.StatusInternalServerError)
				return
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				return
			}
			conn.Close()

		default:
			http.NotFound(w, r)
		}
	}))

}

func (s *HTTPReaderTestSuite) TearDownTest() {
	s.server.Close()
}

func TestHTTPReaderTestSuite(t *testing.T) {
	suite.Run(t, new(HTTPReaderTestSuite))
}

func (s *HTTPReaderTestSuite) TestNewHTTPReaderShouldSuccess() {
	url := s.server.URL + GET_OK_FILE_PATH

	r, err := NewHTTPReader(url)
	s.NotNil(r)
	s.NoError(err)
}

func (s *HTTPReaderTestSuite) TestNewHTTPReaderShouldReturnErrorIfURLDoesNotExist() {
	url := s.server.URL + DOES_NOT_EXIST_FILE_PATH

	r, err := NewHTTPReader(url)
	s.Error(err)
	s.Nil(r)
	s.Equal("url not exists", err.Error())
}

func (s *HTTPReaderTestSuite) TestNewHTTPReaderShouldReturnErrorIfNetworkReset() {
	url := s.server.URL + NETWORK_RESET_ON_HEAD_PATH
	expectedError := "Head \"" + url + "\": EOF"

	_, err := NewHTTPReader(url)
	s.Error(err)
	s.Equal(expectedError, err.Error())
}

func (s *HTTPReaderTestSuite) TestFilenameShouldReturnCorrectFilename() {
	r := &HTTPReader{
		filename: GET_OK_FILE_NAME,
	}

	s.Equal(GET_OK_FILE_NAME, r.Filename())
}

func (s *HTTPReaderTestSuite) TestTotalSizeShouldReturnCorrectTotalSize() {
	r := &HTTPReader{
		totalSize: 100,
	}

	s.Equal(int64(100), r.TotalSize())
}

func (s *HTTPReaderTestSuite) TestReadShouldReturnCorrectData() {
	url := s.server.URL + GET_OK_FILE_PATH
	r, err := NewHTTPReader(url)
	s.NoError(err)
	s.NotNil(r)

	bytes, err := io.ReadAll(r)
	s.NoError(err)
	s.Equal(GET_FILE_CONTENT, string(bytes))
}

func (s *HTTPReaderTestSuite) TestReadShouldReturnCorrectGZIPData() {
	url := s.server.URL + GET_OK_GZIP_FILE_PATH
	r, err := NewHTTPReader(url)
	s.NoError(err)
	s.NotNil(r)

	bytes, err := io.ReadAll(r)
	s.NoError(err)
	s.Equal(GZIP_FILE_CONTENT, string(bytes))
}

func (s *HTTPReaderTestSuite) TestReadShouldReturnErrorIfGetRequestFails() {
	url := s.server.URL + NETWORK_RESET_ON_GET_PATH
	r, err := NewHTTPReader(url)
	s.NoError(err)
	s.NotNil(r)

	expectedError := "Get \"" + url + "\": EOF"

	bytes, err := io.ReadAll(r)
	s.Error(err)
	s.Equal(0, len(bytes))
	s.Equal(expectedError, err.Error())
}

func (s *HTTPReaderTestSuite) TestReadShouldReturnErrorIfGetReturnsNotOK() {
	url := s.server.URL + HEAD_OK_FILE_PATH
	r, err := NewHTTPReader(url)
	s.NoError(err)
	s.NotNil(r)

	bytes, err := io.ReadAll(r)
	s.Error(err)
	s.Equal(0, len(bytes))
	s.Equal("url not exists", err.Error())
}
