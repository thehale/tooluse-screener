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

	"github.com/thehale/tooluse-screener/hook"
	"github.com/thehale/tooluse-screener/policy"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(arguments []string, in io.Reader, out, complaints io.Writer) int {
	cliOptions := optionsFor(complaints)
	flagWords := asFlags(arguments)
	if badFlag := misspelling(flagWords); badFlag != "" {
		return cliOptions.complainOfMisspelling(complaints, badFlag)
	} else {
		return cliOptions.answer(flagWords, in, out, complaints)
	}
}

type options struct {
	flags     *flag.FlagSet
	answering *bool
	reporting *bool
	path      *string
}

func optionsFor(complaints io.Writer) options {
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

func (o options) answer(flagWords []string, in io.Reader, out, complaints io.Writer) int {
	switch err := o.flags.Parse(flagWords); {
	case errors.Is(err, flag.ErrHelp):
		o.printUsage(out)
		return 0
	case err != nil:
		o.printUsage(complaints)
		return usageError
	case *o.reporting:
		_, _ = fmt.Fprintln(out, name, version())
		return 0
	case *o.answering:
		return hook.Main(in, out, complaints, lazyPolicy(*o.path))
	default:
		return o.checkOne(out, complaints)
	}
}

const name = "tooluse-screener"

const usageError = 64

func misspelling(arguments []string) string {
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

func (o options) complainOfMisspelling(complaints io.Writer, badFlag string) int {
	_, _ = fmt.Fprintf(complaints, "%s: %s is not a flag. Write -%s.\n", name, badFlag, badFlag)
	o.printUsage(complaints)
	return usageError
}

func asFlags(arguments []string) []string {
	if len(arguments) > 0 && arguments[0] == "help" {
		return []string{"-h"}
	} else {
		return arguments
	}
}

func version() string {
	build, hasBuild := debug.ReadBuildInfo()
	if hasBuild && build.Main.Version != "" {
		return build.Main.Version
	} else {
		return "(unknown)"
	}
}

var exitCodes = map[policy.Decision]int{policy.Allow: 0, policy.Deny: 1, policy.Ask: 2}

func (o options) checkOne(out, complaints io.Writer) int {
	if o.flags.NArg() != 1 {
		o.flags.Usage()
		return usageError
	} else {
		return o.checkLine(o.flags.Arg(0), out, complaints)
	}
}

func (o options) checkLine(line string, out, complaints io.Writer) int {
	activePolicy, err := policyAt(*o.path)
	if err != nil {
		_, _ = fmt.Fprintf(complaints, "tooluse-screener: %v\n", err)
		return usageError
	}
	verdict := activePolicy.CheckCommand(line)
	_, _ = fmt.Fprintln(out, verdict)
	return exitCodes[verdict.Decision]
}

func (o options) printUsage(writer io.Writer) {
	_, _ = fmt.Fprintln(writer, "Check a Bash command against the shared agent permission policy.")
	_, _ = fmt.Fprintln(writer, "\nUsage: tooluse-screener [flags] <command>\n       tooluse-screener --hook [flags]\n       tooluse-screener help\n\nFlags:")
	o.printFlags(writer)
}

func (o options) printFlags(writer io.Writer) {
	o.flags.VisitAll(func(one *flag.Flag) {
		placeholder, purpose := flag.UnquoteUsage(one)
		_, _ = fmt.Fprintf(writer, "  %s\n        %s\n", flagName(one, placeholder), purpose)
	})
}

func flagName(one *flag.Flag, placeholder string) string {
	return strings.TrimSpace("--" + one.Name + " " + placeholder)
}
