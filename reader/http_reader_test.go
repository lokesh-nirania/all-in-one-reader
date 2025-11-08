package reader_test

import (
	"abc/reader"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type HTTPReaderTestSuite struct {
	suite.Suite
	server *httptest.Server
}

func (s *HTTPReaderTestSuite) SetupTest() {
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok.txt":
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "This is a simple text file for testing purposes.")
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
	url := s.server.URL + "/ok.txt"

	r, err := reader.NewHTTPReader(url)
	s.NoError(err)

	bytes, err := io.ReadAll(r)
	s.NoError(err)
	s.Equal("This is a simple text file for testing purposes.", string(bytes))
}

func (s *HTTPReaderTestSuite) TestNewHTTPReaderShouldReturnErrorIfURLDoesNotExist() {
	url := s.server.URL + "/does-not-exist"

	r, err := reader.NewHTTPReader(url)
	s.Error(err)
	s.Nil(r)
	s.Equal("url not exists", err.Error())
}
