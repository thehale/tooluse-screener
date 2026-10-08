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
	Name    string
	Reason  string
	names   []string
	written string
}

func NewGlob(text string) (Glob, error) {
	expandedPath := directories.WithHomeExpanded(text)
	switch _, err := path.Match(filepath.ToSlash(expandedPath), ""); {
	case !filepath.IsAbs(expandedPath):
		return Glob{}, fmt.Errorf("a path is written from / or ~, and this is not: %s", text)
	case err != nil:
		return Glob{}, fmt.Errorf("%s: %w", text, err)
	default:
		return Glob{names: names(expandedPath), written: text}, nil
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
	isFit, _ := path.Match(glob, name)
	return isFit
}

func (g Glob) String() string {
	return nameOr(g.Name, g.written)
}
