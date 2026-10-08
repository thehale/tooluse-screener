// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import "fmt"

func refusalAt(line int, format string, values ...any) error {
	return fmt.Errorf("line %d: %w", line, fmt.Errorf(format, values...))
}
