// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/thehale/tooluse-screener/policy"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(arguments []string, in io.Reader, out, complaints io.Writer) int {
	cliOptions := optionsFor(complaints)
	return cliOptions.withParsedFlags(asFlags(arguments), out, complaints, func(args []string) int {
		return cliOptions.answer(args, in, out, complaints)
	})
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

func (o options) answer(args []string, in io.Reader, out, complaints io.Writer) int {
	if *o.reporting {
		_, _ = fmt.Fprintln(out, name, version())
		return 0
	} else {
		return o.dispatch(args, in, out, complaints)
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

func complain(complaints io.Writer, err error) int {
	_, _ = fmt.Fprintf(complaints, "%s: %v\n", name, err)
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

func (o options) printUsage(writer io.Writer) {
	_, _ = fmt.Fprintln(writer, "Check a Bash command against the shared agent permission policy.")
	_, _ = fmt.Fprintln(writer, "\nUsage: tooluse-screener [flags] check <command>\n       tooluse-screener [flags] check --hook\n       tooluse-screener help\n\nFlags:")
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
