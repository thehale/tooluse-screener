// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

// Command tooluse-screener asks the policy about a Bash command.
//
//	> tooluse-screener "<command>"
//	decision: reason
//
// It exits 0 for allow, 1 for deny and 2 for ask.
//
// With --hook it is what an agent's settings point at: it reads a hook
// payload on stdin and answers in the shape that agent expects.
//
// `help` and `--version` say what it is and which build this is.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/thehale/tooluse-screener/internal/check"
	"github.com/thehale/tooluse-screener/internal/hook"
	"github.com/thehale/tooluse-screener/internal/policy"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(arguments []string, in io.Reader, out, complaints io.Writer) int {
	asked := flag.NewFlagSet(name, flag.ContinueOnError)
	asked.SetOutput(complaints)
	asked.Usage = func() {}
	answering := asked.Bool("hook", false, "Answer a hook payload read on stdin.")
	reporting := asked.Bool("version", false, "Print the version and exit.")
	path := asked.String("config-file", "", "Policy file to ask, before "+policy.Variable+" and this machine's own.")

	switch err := asked.Parse(asFlags(arguments)); {
	case errors.Is(err, flag.ErrHelp):
		usage(asked, out)
		return 0
	case err != nil:
		usage(asked, complaints)
		return usageError
	case *reporting:
		_, _ = fmt.Fprintln(out, name, version())
		return 0
	case *answering:
		return hook.Main(in, out, complaints, checking(*path))
	default:
		return checkOne(asked, out, complaints, *path)
	}
}

const name = "tooluse-screener"

const usageError = 64

func asFlags(arguments []string) []string {
	switch {
	case len(arguments) > 0 && arguments[0] == "help":
		return []string{"-h"}
	default:
		return arguments
	}
}

func version() string {
	build, known := debug.ReadBuildInfo()
	switch {
	case known && build.Main.Version != "":
		return build.Main.Version
	default:
		return "(unknown)"
	}
}

var exitCodes = map[check.Decision]int{check.Allow: 0, check.Deny: 1, check.Ask: 2}

func checkOne(asked *flag.FlagSet, out, complaints io.Writer, path string) int {
	if asked.NArg() != 1 {
		asked.Usage()
		return usageError
	}
	chosen, err := policy.Chosen(path)
	if err != nil {
		_, _ = fmt.Fprintf(complaints, "tooluse-screener: %v\n", err)
		return usageError
	}
	verdict := check.Evaluate(asked.Arg(0), chosen)
	_, _ = fmt.Fprintln(out, verdict)
	return exitCodes[verdict.Decision]
}

func checking(path string) hook.Checking {
	return func(line string) check.Verdict {
		chosen, err := policy.Chosen(path)
		if err != nil {
			panic(fmt.Sprintf("the policy will not read: %v", err))
		}
		return check.Evaluate(line, chosen)
	}
}

func usage(asked *flag.FlagSet, writing io.Writer) {
	_, _ = fmt.Fprintln(writing, "Check a Bash command against the shared agent permission policy.")
	_, _ = fmt.Fprintln(writing, "\nUsage: tooluse-screener [flags] <command>\n       tooluse-screener --hook [flags]\n       tooluse-screener help\n\nFlags:")
	asked.SetOutput(writing)
	asked.PrintDefaults()
}
