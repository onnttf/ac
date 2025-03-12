package util

import (
	"fmt"
	"os"

	jsoniter "github.com/json-iterator/go"
)

// ReadJSON reads a JSON file and unmarshals it into the provided interface.
func ReadJSON(filePath string, v any) error {
	byteValue, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s, err: %w", filePath, err)
	}
	if err := jsoniter.Unmarshal(byteValue, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON, err: %w", err)
	}
	return nil
}
