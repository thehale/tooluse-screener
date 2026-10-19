// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"errors"
	"flag"
	"io"
)

func (o options) withParsedFlags(tail []string, out, complaints io.Writer, then func([]string) int) int {
	if badFlag := misspelling(tail); badFlag == "" {
		return o.parseFlags(tail, out, complaints, then)
	} else {
		return o.complainOfMisspelling(complaints, badFlag)
	}
}

func (o options) parseFlags(tail []string, out, complaints io.Writer, then func([]string) int) int {
	switch err := o.flags.Parse(tail); {
	case errors.Is(err, flag.ErrHelp):
		o.printUsage(out)
		return 0
	case err != nil:
		o.printUsage(complaints)
		return usageError
	default:
		return then(o.flags.Args())
	}
}
