// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/thehale/tooluse-screener/internal/lists"
	"github.com/thehale/tooluse-screener/policy"
)

func (o options) runTests(tail []string, out, complaints io.Writer) int {
	return o.withParsedFlags(tail, out, complaints, func(args []string) int {
		if *o.answering {
			return o.complainOfHookOutsideCheck(complaints)
		} else {
			return o.runSpecs(args, out, complaints)
		}
	})
}

func (o options) runSpecs(paths []string, out, complaints io.Writer) int {
	if len(paths) == 0 {
		o.printUsage(complaints)
		return usageError
	} else {
		return o.testSpecs(paths, out, complaints)
	}
}

func (o options) testSpecs(paths []string, out, complaints io.Writer) int {
	activePolicy, policyErr := policyAt(*o.path)
	specs, specsErr := policy.ReadSpecs(paths...)
	if err := errors.Join(policyErr, specsErr); err == nil {
		return runResults(specs.Test(activePolicy), out)
	} else {
		return complain(complaints, err)
	}
}

func runResults(results []policy.Result, out io.Writer) int {
	passed, failed := lists.Partition(results, func(one policy.Result) bool { return one.Outcome == policy.Pass })
	reportResults(out, results, len(passed), len(failed))
	return exitCodeFor(len(failed))
}

func reportResults(out io.Writer, results []policy.Result, passed, failed int) {
	for _, one := range results {
		reportResult(out, one)
	}
	_, _ = fmt.Fprintf(out, "%s. %d passed. %d failed\n", caseCount(len(results)), passed, failed)
}

func caseCount(count int) string {
	if count == 1 {
		return "1 case"
	} else {
		return fmt.Sprintf("%d cases", count)
	}
}

var labels = map[policy.Outcome]string{policy.Pass: "ok  ", policy.Fail: "FAIL"}

func reportResult(out io.Writer, one policy.Result) {
	_, _ = fmt.Fprintf(out, "%s %s:%d  %s %% %s\n", labels[one.Outcome], one.Case.Path, one.Case.Line, one.Case.Directory, one.Case.Prompt)
	if one.Outcome == policy.Fail {
		_, _ = fmt.Fprintf(out, "     got  %s\n", one.Verdict)
		_, _ = fmt.Fprintf(out, "     want %s\n", one.Case.Want)
	}
}

func exitCodeFor(failed int) int {
	if failed == 0 {
		return 0
	} else {
		return 1
	}
}
