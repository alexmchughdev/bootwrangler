package profile

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Parse parses one profile YAML document and rejects unknown fields.
func Parse(data []byte) (Profile, error) {
	return Load(bytes.NewReader(data))
}

// Load parses one profile YAML document and rejects unknown fields.
func Load(reader io.Reader) (Profile, error) {
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)

	var result Profile
	if err := decoder.Decode(&result); err != nil {
		if err == io.EOF {
			return Profile{}, fmt.Errorf("parse profile YAML: profile is empty")
		}
		return Profile{}, fmt.Errorf("parse profile YAML: %w", err)
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Profile{}, fmt.Errorf("parse profile YAML: multiple documents are not supported")
		}
		return Profile{}, fmt.Errorf("parse profile YAML: %w", err)
	}

	return result, nil
}

// LoadFile parses one profile YAML file.
func LoadFile(path string) (Profile, error) {
	file, err := os.Open(path)
	if err != nil {
		return Profile{}, fmt.Errorf("open profile %q: %w", path, err)
	}
	defer file.Close()

	result, err := Load(file)
	if err != nil {
		return Profile{}, fmt.Errorf("load profile %q: %w", path, err)
	}

	return result, nil
}

// LoadAndValidateBytes parses and validates a profile from a YAML byte slice.
func LoadAndValidateBytes(data []byte) (Profile, error) {
	result, err := Parse(data)
	if err != nil {
		return Profile{}, err
	}
	if err := Validate(result); err != nil {
		return Profile{}, err
	}
	return result, nil
}

// LoadAndValidateFile parses and validates one profile YAML file.
func LoadAndValidateFile(path string) (Profile, error) {
	result, err := LoadFile(path)
	if err != nil {
		return Profile{}, err
	}
	if err := Validate(result); err != nil {
		return Profile{}, err
	}

	return result, nil
}
