package reader_test

import (
	"abc/reader"
	"io"
	"testing"

	"github.com/stretchr/testify/suite"
)

type FileReaderTestSuite struct {
	suite.Suite
}

func TestFileReaderTestSuite(t *testing.T) {
	suite.Run(t, new(FileReaderTestSuite))
}

func (s *FileReaderTestSuite) TestNewFileReaderShouldSuccess() {
	r, err := reader.NewFileReader("../test.txt")
	s.NoError(err)

	bytes, err := io.ReadAll(r)
	s.NoError(err)
	s.Equal("sample text file", string(bytes))
}

func (s *FileReaderTestSuite) TestNewFileReaderShouldReturnErrorIfFileDoesNotExist() {
	r, err := reader.NewFileReader("nonexistent.txt")
	s.Error(err)
	s.Nil(r)
	s.Equal("file not found", err.Error())
}
