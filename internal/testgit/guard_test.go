// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package testgit_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestTheGuardedPackagesIgnoreAnInheritedGitDir(t *testing.T) {
	scratch := scratchRepository(t)
	before := configOf(t, scratch)

	runGuardedTests(t, scratch)

	if after := configOf(t, scratch); after != before {
		t.Errorf("GIT_DIR=%s config changed from %q to %q", scratch, before, after)
	}
}

func scratchRepository(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run(t, "git", "-C", dir, "init", "--quiet", "--initial-branch", "main")
	run(t, "git", "-C", dir, "config", "user.name", "scratch")
	return dir
}

func runGuardedTests(t *testing.T, scratch string) {
	t.Helper()
	cmd := exec.Command("go", "test", "./internal/git/...", "./internal/rules/...")
	cmd.Dir = repositoryRoot(t)
	cmd.Env = append(os.Environ(), "GIT_DIR="+filepath.Join(scratch, ".git"))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go test under an inherited GIT_DIR: %v\n%s", err, output)
	}
}

func run(t *testing.T, name string, arguments ...string) {
	t.Helper()
	if output, err := exec.Command(name, arguments...).CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, arguments, err, output)
	}
}

func configOf(t *testing.T, repository string) string {
	t.Helper()
	config, err := os.ReadFile(filepath.Join(repository, ".git", "config"))
	if err != nil {
		t.Fatalf("reading scratch config: %v", err)
	}
	return string(config)
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine this file's path")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}
