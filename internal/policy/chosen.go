// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"os"
	"path/filepath"
)

func Chosen(named string) (Policy, error) {
	path, err := preferred(named)
	switch {
	case err != nil:
		return Policy{}, err
	case path == "":
		return Default(), nil
	default:
		return Read(path)
	}
}

const Variable = "TOOLUSE_SCREENER_POLICY_FILE"

const fileName = "tooluse-screener/policy.yaml"

func preferred(named string) (string, error) {
	if named != "" {
		return named, nil
	}
	if set := os.Getenv(Variable); set != "" {
		return set, nil
	}
	return configured()
}

func configured() (string, error) {
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
