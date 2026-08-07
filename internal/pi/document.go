package pi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"piswitch/internal/system"
)

const maxDocumentMutationAttempts = 3

var errConcurrentDocumentChange = errors.New("配置文件在保存期间被其他程序修改")

func mutateJSONDocument(path string, mutate func(map[string]json.RawMessage) error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	for attempt := 0; attempt < maxDocumentMutationAttempts; attempt++ {
		payload, original, existed, err := readJSONDocument(path)
		if err != nil {
			return err
		}
		if err := mutate(payload); err != nil {
			return err
		}

		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		unchanged, err := documentUnchanged(path, original, existed)
		if err != nil {
			return err
		}
		if !unchanged {
			continue
		}
		if err := system.BackupFile(path); err != nil {
			return err
		}
		unchanged, err = documentUnchanged(path, original, existed)
		if err != nil {
			return err
		}
		if !unchanged {
			continue
		}
		return system.WriteFileAtomic(path, data, 0o644)
	}

	return fmt.Errorf("%w：%s", errConcurrentDocumentChange, path)
}

func readJSONDocument(path string) (map[string]json.RawMessage, []byte, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]json.RawMessage{}, nil, false, nil
	}
	if err != nil {
		return nil, nil, false, err
	}

	payload := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, nil, true, err
	}
	return payload, data, true, nil
}

func documentUnchanged(path string, original []byte, existed bool) (bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return !existed, nil
	}
	if err != nil {
		return false, err
	}
	return existed && bytes.Equal(data, original), nil
}

func setJSONField(target map[string]json.RawMessage, key string, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	target[key] = encoded
	return nil
}
