// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"io"

	"github.com/thehale/tooluse-screener/hook"
	"github.com/thehale/tooluse-screener/policy"
)

func (o options) runCheck(tail []string, in io.Reader, out, complaints io.Writer) int {
	return o.withParsedFlags(tail, out, complaints, func(args []string) int {
		if *o.answering {
			return o.checkHook(args, in, out, complaints)
		} else {
			return o.checkOne(args, out, complaints)
		}
	})
}

func (o options) checkHook(args []string, in io.Reader, out, complaints io.Writer) int {
	if len(args) == 0 {
		return hook.Main(in, out, complaints, lazyPolicy(*o.path))
	} else {
		o.printUsage(complaints)
		return usageError
	}
}

var exitCodes = map[policy.Decision]int{policy.Allow: 0, policy.Deny: 1, policy.Ask: 2}

func (o options) checkOne(args []string, out, complaints io.Writer) int {
	if len(args) == 1 {
		return o.checkLine(args[0], out, complaints)
	} else {
		o.printUsage(complaints)
		return usageError
	}
}

func (o options) checkLine(line string, out, complaints io.Writer) int {
	activePolicy, err := policy.Load(*o.path)
	if err == nil {
		verdict := activePolicy.CheckCommand(line)
		_, _ = fmt.Fprintln(out, verdict)
		return exitCodes[verdict.Decision]
	} else {
		return complain(complaints, err)
	}
}
