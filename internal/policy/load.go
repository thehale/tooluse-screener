// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"os"
	"path/filepath"
)

func Load() (Policy, error) {
	path, err := configured()
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
	written, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, err
	}
	return Parse(written)
}

const Variable = "TOOLUSE_SCREENER_POLICY_FILE"
const fileName = "tooluse-screener/policy.yaml"

func configured() (string, error) {
	if set := os.Getenv(Variable); set == "" {
		return inConfigDirectory()
	} else {
		return set, nil
	}
}

func inConfigDirectory() (string, error) {
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
