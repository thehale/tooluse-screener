// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/thehale/tooluse-screener/internal/directories"
)

type Glob struct {
	wording
	names   []string
	written string
}

func NewGlob(written, reason, name string) (Glob, error) {
	expanded := directories.WithHomeExpanded(written)
	switch _, err := path.Match(filepath.ToSlash(expanded), ""); {
	case !filepath.IsAbs(expanded):
		return Glob{}, fmt.Errorf("a path is written from / or ~, and this is not: %s", written)
	case err != nil:
		return Glob{}, fmt.Errorf("%s: %w", written, err)
	default:
		return Glob{wording{reason, name}, names(expanded), written}, nil
	}
}

func (g Glob) IsMatchFor(file string) bool {
	return isMatch(g.names, names(file)) || isMatch(g.names, names(directories.RealPath(file)))
}

func names(file string) []string {
	return strings.Split(filepath.ToSlash(filepath.Clean(file)), "/")
}

func isMatch(glob, file []string) bool {
	switch {
	case len(glob) == 0:
		return len(file) == 0
	case glob[0] == "**":
		return isMatch(glob[1:], file) || len(file) > 0 && isMatch(glob, file[1:])
	default:
		return len(file) > 0 && isNameMatch(glob[0], file[0]) && isMatch(glob[1:], file[1:])
	}
}

func isNameMatch(glob, name string) bool {
	fitting, _ := path.Match(glob, name)
	return fitting
}

func (g Glob) Reason() string {
	return g.reason
}

func (g Glob) String() string {
	return g.nameOr(g.written)
}
