package ytnode

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed default-payloads/*.json
var payloadFiles embed.FS

var defaultPayloadsRaw = make(map[string][]byte)

func InitDefaultPayloads() error {
	if len(defaultPayloadsRaw) != 0 {
		return nil
	}

	files, err := payloadFiles.ReadDir("default-payloads")
	if err != nil {
		return fmt.Errorf("failed to read embedded directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := "default-payloads/" + file.Name()

		content, err := payloadFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %v: %w", file.Name(), err)
		}

		filename := strings.TrimSuffix(file.Name(), ".json")
		defaultPayloadsRaw[filename] = content
	}

	return nil
}

func GetDefaultPayload[T any](name string) (*T, error) {
	raw, ok := defaultPayloadsRaw[name]
	if !ok {
		return nil, fmt.Errorf("payload %s not found", name)
	}

	var payload T
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload %s: %w", name, err)
	}

	return &payload, nil
}
