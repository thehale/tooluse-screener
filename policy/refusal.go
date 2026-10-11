// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package policy

import "fmt"

func refusalAt(line int, format string, values ...any) error {
	return fmt.Errorf("line %d: %w", line, fmt.Errorf(format, values...))
}
