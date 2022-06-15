package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type fileData struct {
	path     string
	fileName string
}

var errorsList = processErrors{
	errors: []string{},
}

type processErrors struct {
	errors []string
}

func (list *processErrors) append(msg string, err error) {
	if err != nil {
		msg = fmt.Sprintf("%s with error %s", msg, err.Error())
	}

	list.errors = append(list.errors, msg)
}

func (list *processErrors) print() {
	for _, err := range list.errors {
		println(err)
	}

	log.Fatal("Mocked files generation check failed. Please check previous lines for more details.")
}

func (list *processErrors) hasErrors() bool {
	return len(list.errors) > 0
}

/**
The script expects the original file to have the generation command in comment. For example:
// mockgen -source=./example/ORIGINAL_FILE.go -destination=./example/mock/MOCKED_FILE.go -package=mock
*/

func main() {
	projectGoFiles := map[string]fileData{}
	mockFiles := map[string]fileData{}

	extractFilesWithMocks(mockFiles, projectGoFiles)

	if !errorsList.hasErrors() {
		validateFiles(mockFiles, projectGoFiles)
	}

	if errorsList.hasErrors() {
		errorsList.print()
	}
}

func extractFilesWithMocks(mockFiles map[string]fileData, projectGoFiles map[string]fileData) {
	err := filepath.Walk("./", func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() && strings.Contains(file.Name(), ".go") && !strings.Contains(file.Name(), "_test") {
			projectGoFiles[file.Name()] = createFileData(path, file)
		}
		if strings.Contains(file.Name(), "_mock.go") {
			mockFiles[file.Name()] = createFileData(path, file)
		}

		return nil
	})

	if err != nil {
		errorsList.append("Failed to extract files to validate mocks generated properly.", err)
	}
}

func createFileData(path string, file os.FileInfo) fileData {
	return fileData{
		path:     path,
		fileName: file.Name(),
	}
}

func validateFiles(mockFiles map[string]fileData, projectFiles map[string]fileData) {
	regex := regexp.MustCompile("mockgen")

	for mockFileName, mockFileData := range mockFiles {
		originalFile, err := getOriginalFile(projectFiles, mockFileName)
		if err != nil {
			errorsList.append("Failed to get original file for mock "+mockFileName, err)

			continue
		}

		defer originalFile.Close()

		match := false

		cmdScanner := bufio.NewScanner(originalFile)
		for cmdScanner.Scan() {
			text := cmdScanner.Text()
			if regex.MatchString(text) {
				match = true
				newFilePath := "./scripts/temp/" + originalFile.Name()

				cmdRunner, err := buildAndCreateANewCommand(text, originalFile, newFilePath)
				if err != nil {
					errorsList.append("Failed to build command for mock validation", err)

					continue
				}

				_, err = cmdRunner.Output()
				if err != nil {
					errorsList.append("Failed to read mockgen command from original file "+originalFile.Name(), err)

					continue
				}

				compareMocks(mockFileData, newFilePath)
			}
		}

		if !match {
			errorsList.append("Failed to get mockgen generating command from original file "+originalFile.Name(), err)

			continue
		}
	}
}

func buildAndCreateANewCommand(text string, originalFile *os.File, newFilePath string) (*exec.Cmd, error) {
	var cmdNewArgs []string

	cmd := strings.Split(text, " ")
	if len(cmd) < 3 {
		err := errors.New("Invalid mockgen generating command " + text + " from original file " + originalFile.Name())

		return nil, err
	}

	for index, cmdArg := range cmd {
		switch {
		case index <= 1:
			continue
		case index != 3:
			cmdNewArgs = append(cmdNewArgs, cmdArg)
		case index == 3:
			cmdNewArgs = append(cmdNewArgs, "-destination="+newFilePath)
		}
	}

	cmdRunner := exec.Command("mockgen", cmdNewArgs...)

	return cmdRunner, nil
}

func getOriginalFile(projectFiles map[string]fileData, mockFileName string) (*os.File, error) {
	originalFileName := strings.Split(mockFileName, "_mock.go")[0]

	originalFileData, ok := projectFiles[originalFileName+".go"]
	if !ok {
		errorMsg := "Did not find original file for mock " + mockFileName + " this mock might be stale."

		return nil, errors.New(errorMsg)
	}

	originalFile, err := os.Open(originalFileData.path)
	if err != nil {
		return nil, err
	}

	return originalFile, nil
}

func compareMocks(mockFileData fileData, newFilePath string) {
	originalMockFile, err := os.Open(mockFileData.path)
	if err != nil {
		errorsList.append("Failed to open generated mock file "+newFilePath, err)

		return
	}
	defer originalMockFile.Close()

	generatedMockFile, err := os.Open(newFilePath)
	if err != nil {
		errorsList.append("Failed to open generated mock file "+newFilePath, err)

		return
	}

	defer os.Remove(newFilePath)
	defer generatedMockFile.Close()

	originalMockHash, err := hashFile(originalMockFile)
	if err != nil {
		return
	}

	generatedMockHash, err := hashFile(generatedMockFile)
	if err != nil {
		return
	}

	if !bytes.Equal(originalMockHash, generatedMockHash) {
		errorsList.append("Mock file "+mockFileData.fileName+" should be regenerated", nil)
	}
}

func hashFile(file *os.File) ([]byte, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		errorsList.append("Failed to calculate hash for file "+file.Name(), err)

		return nil, err
	}

	return hash.Sum(nil), nil
}
