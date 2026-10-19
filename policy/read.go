// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"errors"
	"fmt"
	"os"

	"github.com/thehale/tooluse-screener/internal/files"
)

func ReadSpecs(paths ...string) (Specs, error) {
	var specs Specs
	errs := make([]error, len(paths))
	for index, path := range paths {
		var found Specs
		found, errs[index] = specsAt(path)
		specs = append(specs, found...)
	}
	return specs, errors.Join(errs...)
}

func specsAt(path string) (Specs, error) {
	stat, err := os.Stat(path)
	switch {
	case err != nil:
		return nil, err
	case stat.IsDir():
		return specsUnder(path)
	default:
		spec, err := specAt(path)
		return Specs{spec}, err
	}
}

func specsUnder(root string) (Specs, error) {
	testFiles, err := files.WithSuffix(".test", root)
	switch {
	case err != nil:
		return nil, err
	case len(testFiles) == 0:
		return nil, fmt.Errorf("%s: no .test files", root)
	default:
		return specsFromFiles(testFiles)
	}
}

func specsFromFiles(paths []string) (Specs, error) {
	specs := make(Specs, len(paths))
	errs := make([]error, len(paths))
	for index, path := range paths {
		specs[index], errs[index] = specAt(path)
	}
	return specs, errors.Join(errs...)
}

func specAt(path string) (Spec, error) {
	contents, readErr := os.ReadFile(path)
	spec, parseErr := ParseSpec(string(contents))
	switch {
	case readErr != nil:
		return Spec{}, fmt.Errorf("%s: %w", path, readErr)
	case parseErr != nil:
		return Spec{}, fmt.Errorf("%s: %w", path, parseErr)
	default:
		return withPath(spec, path), nil
	}
}

func withPath(spec Spec, path string) Spec {
	for index := range spec.Cases {
		spec.Cases[index].Path = path
	}
	return spec
}
