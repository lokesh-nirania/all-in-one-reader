package reader_test

import (
	"abc/reader"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ReaderTestSuite struct {
	suite.Suite
	server *httptest.Server
}

func TestReaderTestSuite(t *testing.T) {
	suite.Run(t, new(ReaderTestSuite))

}

func (s *ReaderTestSuite) SetupTest() {
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

func (s *ReaderTestSuite) TearDownTest() {
	s.server.Close()
}

func (s *ReaderTestSuite) TestNewReaderShouldSuccessForFile() {
	r, err := reader.NewReader("file://../test.txt")

	s.NoError(err)
	s.NotNil(r)
}

func (s *ReaderTestSuite) TestNewReaderShouldReturnErrorIfFileDoesNotExist() {
	r, err := reader.NewReader("file://not-ok.txt")

	s.Error(err)
	s.Nil(r)
	s.Equal("file not found", err.Error())
}

func (s *ReaderTestSuite) TestNewReaderShouldSuccessForHTTP() {
	url := s.server.URL + "/ok.txt"
	r, err := reader.NewReader(url)

	s.NoError(err)
	s.NotNil(r)
}

func (s *ReaderTestSuite) TestNewReaderShouldReturnErrorIfHTTPDoesNotExist() {
	url := s.server.URL + "/does-not-exist"
	r, err := reader.NewReader(url)

	s.Error(err)
	s.Nil(r)
	s.Equal("url not exists", err.Error())
}
func (s *ReaderTestSuite) TestNewReaderShouldReturnErrorIfUnsupportedScheme() {
	r, err := reader.NewReader("ftp://test.txt")

	s.Error(err)
	s.Nil(r)
	s.Equal("unsupported scheme", err.Error())
}

func (s *ReaderTestSuite) TestReadShouldReturnEOFIfReaderIsNil() {
	r := &reader.Reader{}

	_, err := r.Read(make([]byte, 1024))
	s.Error(err)
	s.Equal(io.EOF, err)
}

func (s *ReaderTestSuite) TestReadShouldReturnDataIfReaderIsNotNil() {
	r, err := reader.NewReader("file://../test.txt")

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	s.NoError(err)
	s.Equal(16, n)
	s.Equal("sample text file", string(buf[:n]))
}
