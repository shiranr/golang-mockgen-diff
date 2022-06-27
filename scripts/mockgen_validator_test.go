package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestExtractFilesWithMocksSuccessfullyWithEnvVariable(t *testing.T) {
	projectGoFiles := map[string]fileData{}
	mockFiles := map[string]fileData{}
	os.Setenv("PROJECT_PATH", "../")
	extractFilesWithMocks(mockFiles, projectGoFiles)
	assert.Len(t, mockFiles, 1)
	assert.Len(t, projectGoFiles, 3)
}

func TestExtractFilesWithMocksSuccessfullyWithDefaultPath(t *testing.T) {
	projectGoFiles := map[string]fileData{}
	mockFiles := map[string]fileData{}
	os.Unsetenv("PROJECT_PATH")
	extractFilesWithMocks(mockFiles, projectGoFiles)
	assert.Len(t, mockFiles, 0)
	assert.Len(t, projectGoFiles, 1)
}
