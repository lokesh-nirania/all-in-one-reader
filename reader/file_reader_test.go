package reader

import (
	"io"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	FR_FILE_LOCAL_NAME        = "test.txt"
	FR_FILE_LOCAL_REL_PATH    = "../test.txt"
	FR_FILE_LOCAL_CONTENT     = "sample text file"
	FR_FILE_LOCAL_SIZE        = 16
	FR_NONEXISTENT_PATH       = "nonexistent.txt"
	FR_NONEXISTENT_OPEN_ERROR = "open nonexistent.txt: no such file or directory"
)

type FileReaderTestSuite struct {
	suite.Suite
}

func TestFileReaderTestSuite(t *testing.T) {
	suite.Run(t, new(FileReaderTestSuite))
}

func (s *FileReaderTestSuite) TestNewFileReaderShouldSuccess() {
	r, err := NewFileReader(FR_FILE_LOCAL_REL_PATH)
	s.NoError(err)

	bytes, err := io.ReadAll(r)
	s.NoError(err)
	s.Equal(FR_FILE_LOCAL_CONTENT, string(bytes))
}

func (s *FileReaderTestSuite) TestNewFileReaderShouldReturnErrorIfFileDoesNotExist() {
	r, err := NewFileReader(FR_NONEXISTENT_PATH)
	s.Error(err)
	s.Nil(r)
	s.Equal("file not found", err.Error())
}

func (s *FileReaderTestSuite) TestFilenameShouldReturnCorrectFilename() {
	r, err := NewFileReader(FR_FILE_LOCAL_REL_PATH)
	s.NoError(err)

	s.Equal(FR_FILE_LOCAL_NAME, r.Filename())
}

func (s *FileReaderTestSuite) TestTotalSizeShouldReturnCorrectTotalSize() {
	r, err := NewFileReader(FR_FILE_LOCAL_REL_PATH)
	s.NoError(err)

	s.Equal(int64(FR_FILE_LOCAL_SIZE), r.TotalSize())
}

func (s *FileReaderTestSuite) TestReadShouldReturnCorrectData() {
	r := &FileReader{
		src: FR_FILE_LOCAL_REL_PATH,
	}

	bytes, err := io.ReadAll(r)
	s.NoError(err)
	s.Equal(FR_FILE_LOCAL_CONTENT, string(bytes))
}

func (s *FileReaderTestSuite) TestReadShouldReturnErrorIfFileDoesNotExist() {
	r := &FileReader{
		src: FR_NONEXISTENT_PATH,
	}

	bytes, err := io.ReadAll(r)
	s.Equal(0, len(bytes))
	s.Error(err)
	s.Equal(FR_NONEXISTENT_OPEN_ERROR, err.Error())
}
