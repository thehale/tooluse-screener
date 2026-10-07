// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"os"
	"path/filepath"
)

func Load() (Policy, error) {
	path, err := configuredPath()
	switch {
	case err != nil:
		return Policy{}, err
	case path == "":
		return Default(), nil
	default:
		return LoadFile(path)
	}
}

func LoadFile(path string) (Policy, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, err
	}
	return Parse(contents)
}

const Variable = "TOOLUSE_SCREENER_POLICY_FILE"
const fileName = "tooluse-screener/policy.yaml"

func configuredPath() (string, error) {
	if envPath := os.Getenv(Variable); envPath == "" {
		return pathInConfigDirectory()
	} else {
		return envPath, nil
	}
}

func pathInConfigDirectory() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", nil
	}
	path := filepath.Join(directory, fileName)
	switch _, err := os.Stat(path); {
	case err == nil:
		return path, nil
	case os.IsNotExist(err):
		return "", nil
	default:
		return "", err
	}
}
