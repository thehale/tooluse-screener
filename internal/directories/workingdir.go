// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package directories

import "github.com/thehale/tooluse-screener/internal/lists"

type WorkingDir string

func (w WorkingDir) Resolve(path string) string {
	if w == "" {
		return path
	} else {
		return Within(string(w), path)
	}
}

func (w WorkingDir) ResolveAll(paths []string) []string {
	return lists.Map(paths, w.Resolve)
}
