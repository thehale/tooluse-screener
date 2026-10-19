// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"io"
)

func (o options) dispatch(arguments []string, in io.Reader, out, complaints io.Writer) int {
	subcommand, tail := firstAndRest(arguments)
	switch {
	case subcommand == "" && *o.answering:
		return o.complainOfHookOutsideCheck(complaints)
	case subcommand == "":
		o.printUsage(complaints)
		return usageError
	case subcommand == "check":
		return o.runCheck(tail, in, out, complaints)
	default:
		return o.complainOfUnknownSubcommand(subcommand, complaints)
	}
}

func firstAndRest(arguments []string) (string, []string) {
	if len(arguments) == 0 {
		return "", nil
	} else {
		return arguments[0], arguments[1:]
	}
}

func (o options) complainOfUnknownSubcommand(subcommand string, complaints io.Writer) int {
	_, _ = fmt.Fprintf(complaints, "%s: %q is not a subcommand. Use: %s check '%s'\n", name, subcommand, name, subcommand)
	o.printUsage(complaints)
	return usageError
}

func (o options) complainOfHookOutsideCheck(complaints io.Writer) int {
	_, _ = fmt.Fprintf(complaints, "%s: use %s check --hook\n", name, name)
	o.printUsage(complaints)
	return usageError
}
