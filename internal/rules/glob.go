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
	described
	names   []string
	written string
}

func NewGlob(written, reason, description string) (Glob, error) {
	expanded := directories.WithHomeExpanded(written)
	switch _, err := path.Match(filepath.ToSlash(expanded), ""); {
	case !filepath.IsAbs(expanded):
		return Glob{}, fmt.Errorf("a path is written from / or ~, and this is not: %s", written)
	case err != nil:
		return Glob{}, fmt.Errorf("%s: %w", written, err)
	default:
		return Glob{described{reason, description}, names(expanded), written}, nil
	}
}

func (g Glob) Matches(file string) bool {
	return fits(g.names, names(file)) || fits(g.names, names(directories.Followed(file)))
}

func names(file string) []string {
	return strings.Split(filepath.ToSlash(filepath.Clean(file)), "/")
}

func fits(glob, file []string) bool {
	switch {
	case len(glob) == 0:
		return len(file) == 0
	case glob[0] == "**":
		return fits(glob[1:], file) || len(file) > 0 && fits(glob, file[1:])
	default:
		return len(file) > 0 && fitsName(glob[0], file[0]) && fits(glob[1:], file[1:])
	}
}

func fitsName(glob, name string) bool {
	fitting, _ := path.Match(glob, name)
	return fitting
}

func (g Glob) Reason() string {
	return g.reason
}

func (g Glob) String() string {
	return g.describes(g.written)
}
