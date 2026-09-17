// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/thehale/tooluse-screener/internal/check"
	"github.com/thehale/tooluse-screener/internal/hook"
	"github.com/thehale/tooluse-screener/internal/policy"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(arguments []string, in io.Reader, out, complaints io.Writer) int {
	asked := reading(complaints)
	spelled := asFlags(arguments)

	switch wrong := misspelled(spelled); {
	case wrong != "":
		return misspelling(asked.flags, complaints, wrong)
	default:
		return asked.answer(spelled, in, out, complaints)
	}
}

type options struct {
	flags     *flag.FlagSet
	answering *bool
	reporting *bool
	path      *string
}

func reading(complaints io.Writer) options {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(complaints)
	flags.Usage = func() {}

	return options{
		flags:     flags,
		answering: flags.Bool("hook", false, "Answer a hook payload read on stdin."),
		reporting: flags.Bool("version", false, "Print the version and exit."),
		path:      flags.String("config-file", "", "Policy file to ask, before "+policy.Variable+" and this machine's own."),
	}
}

func (o options) answer(spelled []string, in io.Reader, out, complaints io.Writer) int {
	switch err := o.flags.Parse(spelled); {
	case errors.Is(err, flag.ErrHelp):
		usage(o.flags, out)
		return 0
	case err != nil:
		usage(o.flags, complaints)
		return usageError
	case *o.reporting:
		_, _ = fmt.Fprintln(out, name, version())
		return 0
	case *o.answering:
		return hook.Main(in, out, complaints, checking(*o.path))
	default:
		return checkOne(o.flags, out, complaints, *o.path)
	}
}

const name = "tooluse-screener"

const usageError = 64

func misspelled(arguments []string) string {
	for _, one := range arguments {
		switch {
		case one == "--" || !strings.HasPrefix(one, "-"):
			return ""
		case one == "-h" || strings.HasPrefix(one, "--"):
			continue
		default:
			return one
		}
	}
	return ""
}

func misspelling(asked *flag.FlagSet, complaints io.Writer, wrong string) int {
	_, _ = fmt.Fprintf(complaints, "%s: %s is not a flag. Write -%s.\n", name, wrong, wrong)
	usage(asked, complaints)
	return usageError
}

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
	flags(asked, writing)
}

func flags(asked *flag.FlagSet, writing io.Writer) {
	asked.VisitAll(func(one *flag.Flag) {
		placeholder, purpose := flag.UnquoteUsage(one)
		_, _ = fmt.Fprintf(writing, "  %s\n        %s\n", named(one, placeholder), purpose)
	})
}

func named(one *flag.Flag, placeholder string) string {
	return strings.TrimSpace("--" + one.Name + " " + placeholder)
}
