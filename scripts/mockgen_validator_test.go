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
	for key, val := range mockFiles {
		assert.Equal(t, key, "service_example_mock.go")
		assert.Equal(t, val.fileName, "service_example_mock.go")
		assert.Contains(t, val.path, "example/mock/service_example_mock.go")
	}
	assert.Len(t, projectGoFiles, 3)
}

func TestExtractFilesWithMocksSuccessfullyWithDefaultPath(t *testing.T) {
	projectGoFiles := map[string]fileData{}
	mockFiles := map[string]fileData{}
	os.Unsetenv("PROJECT_PATH")
	extractFilesWithMocks(mockFiles, projectGoFiles)
	assert.Len(t, mockFiles, 0)
	assert.Len(t, projectGoFiles, 1)
	assert.Equal(t, projectGoFiles["mockgen_validator.go"].fileName, "mockgen_validator.go")
	assert.Equal(t, projectGoFiles["mockgen_validator.go"].path, "mockgen_validator.go")
}
