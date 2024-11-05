package helpers

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock struct to represent configuration for testing
type MockFile struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// Test ReadCfgJSON with a valid JSON file
func TestReadCfgJSON_ValidFile(t *testing.T) {
	// Create a temporary file with valid JSON content
	tmpFile, err := os.CreateTemp("", "file_*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name()) // Clean up the file after the test

	// Write valid JSON to the temporary file
	mockFile := MockFile{Name: "test", Value: 42}
	jsonData, err := json.Marshal(mockFile)
	assert.NoError(t, err)
	_, err = tmpFile.Write(jsonData)
	assert.NoError(t, err)
	tmpFile.Close()

	// Call ReadCfgJSON and check results
	var result MockFile
	data, err := ReadCfgJSON[MockFile](tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, data)

	result = *data
	assert.Equal(t, mockFile.Name, result.Name)
	assert.Equal(t, mockFile.Value, result.Value)
}

// Test ReadCfgJSON with an invalid JSON file
func TestReadCfgJSON_InvalidFile(t *testing.T) {
	// Create a temporary file with invalid JSON content
	tmpFile, err := os.CreateTemp("", "file_invalid_*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name()) // Clean up the file after the test

	// Write invalid JSON to the temporary file
	_, err = tmpFile.WriteString("{invalid_json}")
	assert.NoError(t, err)
	tmpFile.Close()

	// Call ReadCfgJSON and expect an error
	_, err = ReadCfgJSON[MockFile](tmpFile.Name())
	assert.Error(t, err)
}

// Test ReadCfgJSON with a non-existent file
func TestReadCfgJSON_NonExistentFile(t *testing.T) {
	// Call ReadCfgJSON with a path that does not exist
	_, err := ReadCfgJSON[MockFile]("non_existent_file.json")
	assert.Error(t, err)
}
