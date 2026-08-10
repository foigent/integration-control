package campaign

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

const maxContractBytes = 4 << 20

func decodeStrict[T any](body []byte) (T, error) {
	var zero T
	if len(body) == 0 || len(body) > maxContractBytes || !utf8.Valid(body) {
		return zero, errors.New("contract JSON size or UTF-8 encoding is invalid")
	}
	if err := rejectDuplicateKeys(body); err != nil {
		return zero, err
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	var value T
	if err := decoder.Decode(&value); err != nil {
		return zero, fmt.Errorf("decode strict contract JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return zero, errors.New("contract JSON contains trailing content")
	}
	return value, nil
}

func rejectDuplicateKeys(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return errors.New("contract JSON contains trailing content")
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("scan contract JSON: %w", err)
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]struct{}{}
		for decoder.More() {
			keyToken, keyErr := decoder.Token()
			if keyErr != nil {
				return fmt.Errorf("scan contract object key: %w", keyErr)
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("contract object key is not a string")
			}
			if _, duplicate := seen[key]; duplicate {
				return fmt.Errorf("contract JSON contains duplicate key %q", key)
			}
			seen[key] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		end, endErr := decoder.Token()
		if endErr != nil || end != json.Delim('}') {
			return errors.New("contract JSON object is not closed")
		}
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		end, endErr := decoder.Token()
		if endErr != nil || end != json.Delim(']') {
			return errors.New("contract JSON array is not closed")
		}
	default:
		return errors.New("contract JSON contains an unexpected delimiter")
	}
	return nil
}

func documentSchemaVersion(body []byte) (string, error) {
	if len(body) == 0 || len(body) > maxContractBytes || !utf8.Valid(body) {
		return "", errors.New("referenced document size or UTF-8 encoding is invalid")
	}
	if err := rejectDuplicateKeys(body); err != nil {
		return "", err
	}
	var envelope struct {
		SchemaVersion string `json:"schema_version"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&envelope); err != nil {
		return "", fmt.Errorf("decode referenced document identity: %w", err)
	}
	if envelope.SchemaVersion == "" {
		return "", errors.New("referenced document does not declare schema_version")
	}
	return envelope.SchemaVersion, nil
}
