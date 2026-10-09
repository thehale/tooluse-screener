// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/thehale/tooluse-screener/internal/directories"
)

type Glob struct {
	Name    string
	Reason  string
	written string
}

func NewGlob(text string) (Glob, error) {
	expandedPath, isKnown := directories.WithHomeExpanded(text)
	switch _, err := path.Match(filepath.ToSlash(expandedPath), ""); {
	case !isKnown:
		message := fmt.Sprintf("path `%s` names another user's home. Write it from / instead.", text)
		return Glob{}, errors.New(message)
	case err != nil:
		return Glob{}, fmt.Errorf("invalid glob `%s`: %s", text, err)
	default:
		return Glob{written: text}, nil
	}
}

func (g Glob) IsMatchFor(file string) bool {
	glob := names(g.absolutePath())
	return isMatch(glob, names(file)) || isMatch(glob, names(directories.RealPath(file)))
}

func (g Glob) absolutePath() string {
	expandedPath, _ := directories.WithHomeExpanded(g.written)
	absolutePath, _ := filepath.Abs(expandedPath)
	return absolutePath
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
