package reader

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	// Local file constants
	FILE_LOCAL_NAME       = "test.txt"
	FILE_LOCAL_REL_PATH   = "../test.txt"
	FILE_LOCAL_SCHEME     = "file://" + FILE_LOCAL_REL_PATH
	FILE_LOCAL_CONTENT    = "sample text file"
	FILE_LOCAL_SIZE       = 16
	FILE_NOT_EXIST_SCHEME = "file://not-ok.txt"
	FILE_LOCAL_UNIQUE_1   = "test_1.txt"

	// HTTP constants
	HTTP_OK_FILE_NAME        = "ok.txt"
	HTTP_OK_FILE_PATH        = "/" + HTTP_OK_FILE_NAME
	HTTP_OK_FILE_CONTENT     = "This is a simple text file for testing purposes."
	HTTP_NOT_EXIST_FILE_PATH = "/" + "does-not-exist"

	// Unsupported scheme
	UNSUPPORTED_SCHEME_URL = "ftp://test.txt"
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
		case HTTP_OK_FILE_PATH:
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, HTTP_OK_FILE_CONTENT)
		default:
			http.NotFound(w, r)
		}
	}))
}

func (s *ReaderTestSuite) TearDownTest() {
	s.server.Close()
}

func (s *ReaderTestSuite) TestNewReaderShouldSuccessForFile() {
	r, err := NewReader(FILE_LOCAL_SCHEME)

	s.NoError(err)
	s.NotNil(r)
}

func (s *ReaderTestSuite) TestNewReaderShouldReturnErrorIfFileDoesNotExist() {
	r, err := NewReader(FILE_NOT_EXIST_SCHEME)

	s.Error(err)
	s.Nil(r)
	s.Equal("file not found", err.Error())
}

func (s *ReaderTestSuite) TestNewReaderShouldSuccessForHTTP() {
	url := s.server.URL + HTTP_OK_FILE_PATH
	r, err := NewReader(url)

	s.NoError(err)
	s.NotNil(r)
}

func (s *ReaderTestSuite) TestNewReaderShouldReturnErrorIfHTTPDoesNotExist() {
	url := s.server.URL + HTTP_NOT_EXIST_FILE_PATH
	r, err := NewReader(url)

	s.Error(err)
	s.Nil(r)
	s.Equal("url not exists", err.Error())
}
func (s *ReaderTestSuite) TestNewReaderShouldReturnErrorIfUnsupportedScheme() {
	r, err := NewReader(UNSUPPORTED_SCHEME_URL)

	s.Error(err)
	s.Nil(r)
	s.Equal("unsupported scheme", err.Error())
}

func (s *ReaderTestSuite) TestReadShouldReturnEOFIfReaderIsNil() {
	r := &Reader{}

	_, err := r.Read(make([]byte, 1024))
	s.Error(err)
	s.Equal(io.EOF, err)
}

func (s *ReaderTestSuite) TestReadShouldReturnDataIfReaderIsNotNil() {
	r, err := NewReader(FILE_LOCAL_SCHEME)

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	s.NoError(err)
	s.Equal(FILE_LOCAL_SIZE, n)
	s.Equal(FILE_LOCAL_CONTENT, string(buf[:n]))
}

func (s *ReaderTestSuite) TestStreamToFileShouldReturnErrorIfReaderIsNil() {
	r := &Reader{}
	dest := s.T().TempDir()

	path, n, err := r.StreamToFile(dest)
	s.Error(err)
	s.Equal("", path)
	s.Equal(int64(0), n)
	s.Equal("reader source is nil", err.Error())
}

func (s *ReaderTestSuite) TestStreamToFileShouldWriteFileForLocalFile() {
	r, err := NewReader(FILE_LOCAL_SCHEME)
	s.NoError(err)
	s.NotNil(r)

	dest := s.T().TempDir()
	path, n, err := r.StreamToFile(dest)
	s.NoError(err)
	s.NotEmpty(path)
	s.Equal(int64(FILE_LOCAL_SIZE), n)
	s.Equal(FILE_LOCAL_NAME, filepath.Base(path))

	data, readErr := os.ReadFile(path)
	s.NoError(readErr)
	s.Equal(FILE_LOCAL_CONTENT, string(data))
}

func (s *ReaderTestSuite) TestStreamToFileShouldCreateUniqueNameIfExists() {
	r, err := NewReader(FILE_LOCAL_SCHEME)
	s.NoError(err)
	s.NotNil(r)

	dest := s.T().TempDir()

	// Pre-create a file with the expected final name to force unique naming
	preExisting := filepath.Join(dest, FILE_LOCAL_NAME)
	writeErr := os.WriteFile(preExisting, []byte("existing"), 0o644)
	s.NoError(writeErr)

	path, n, err := r.StreamToFile(dest)
	s.NoError(err)
	s.NotEmpty(path)
	s.Equal(int64(FILE_LOCAL_SIZE), n)
	s.Equal(FILE_LOCAL_UNIQUE_1, filepath.Base(path))

	data, readErr := os.ReadFile(path)
	s.NoError(readErr)
	s.Equal(FILE_LOCAL_CONTENT, string(data))
}

func (s *ReaderTestSuite) TestStreamToFileShouldWriteFileForHTTP() {
	url := s.server.URL + HTTP_OK_FILE_PATH
	r, err := NewReader(url)
	s.NoError(err)
	s.NotNil(r)

	dest := s.T().TempDir()
	path, n, err := r.StreamToFile(dest)
	s.NoError(err)
	s.NotEmpty(path)
	s.Equal(HTTP_OK_FILE_NAME, filepath.Base(path))

	expected := HTTP_OK_FILE_CONTENT
	s.Equal(int64(len(expected)), n)

	data, readErr := os.ReadFile(path)
	s.NoError(readErr)
	s.Equal(expected, string(data))
}
