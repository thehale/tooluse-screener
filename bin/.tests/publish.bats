#!/usr/bin/env bats
# Copyright (c) Joseph Hale, 2026
# SPDX-License-Identifier: MPL-2.0

setup() {
	REPO="$(cd "$BATS_TEST_DIRNAME/../.." && pwd)"
	RELEASE="$REPO/bin/${BATS_TEST_FILENAME##*/}"
	RELEASE="${RELEASE%.bats}"

	unset "${!GIT_@}"
	unset CI

	PATH="$BATS_TEST_TMPDIR/stub:$PATH"

	export GIT_CONFIG_GLOBAL=/dev/null
	export GIT_CONFIG_SYSTEM=/dev/null

	ORIGIN="${BATS_TEST_TMPDIR:?}/origin.git"
	VERSION="1.2.3"

	mkdir --parents "$BATS_TEST_TMPDIR/work"
	cd "$BATS_TEST_TMPDIR/work"
	unpublishable
	release_history
}

unpublishable() {
	mkdir --parents "$BATS_TEST_TMPDIR/stub"
	printf '#!/usr/bin/env bash\necho "the tests do not publish" >&2\nexit 1\n' >"$BATS_TEST_TMPDIR/stub/goreleaser"
	chmod +x "$BATS_TEST_TMPDIR/stub/goreleaser"
}

release_history() {
	git init --quiet --initial-branch main .
	git remote add origin "$ORIGIN"
	printf 'module %s\n\ngo 1.26.8\n' "${ORIGIN%.git}" >go.mod
	checks "exit 0"
	git add -A
	commit "first"
	commit "second"
}

checks() {
	mkdir --parents bin
	printf '#!/usr/bin/env bash\n%s\n' "$1" >bin/ci
	chmod +x bin/ci
}

commit() {
	git -c user.email=a@b -c user.name=a commit --quiet --allow-empty --message "$1"
}

serve() {
	git clone --quiet --bare . "$ORIGIN"
}

@test "refuses to publish from CI" {
	CI=1 run "$RELEASE"

	[ "$status" -eq 1 ]
	[[ "$output" == *"CI publishes nothing"* ]]
}

@test "refuses to publish a checkout whose checks fail" {
	checks "exit 1"

	run "$RELEASE"

	[ "$status" -eq 1 ]
	[[ "$output" == *"bin/ci failed"* ]]
}

@test "catches a file the checks themselves wrote" {
	checks "echo built >artifact.txt"
	git add -A
	commit "checks that write"
	git tag "v$VERSION"

	run "$RELEASE"

	[ "$status" -eq 1 ]
	[[ "$output" == *"working tree has changes"* ]]
}

@test "refuses to publish from a branch that is not main" {
	git checkout --quiet -b release

	run "$RELEASE"

	[ "$status" -eq 1 ]
	[[ "$output" == *"published from main"* ]]
}

@test "refuses to publish a working tree with changes" {
	echo "stray" >stray.txt

	run "$RELEASE"

	[ "$status" -eq 1 ]
	[[ "$output" == *"working tree has changes"* ]]
}

@test "refuses to publish a commit that carries no release tag" {
	run "$RELEASE"

	[ "$status" -eq 1 ]
	[[ "$output" == *"no vX.Y.Z tag"* ]]
}

@test "refuses to publish what origin has never heard of" {
	git tag "v$VERSION"
	git init --quiet --bare "$ORIGIN"

	run "$RELEASE"

	[ "$status" -eq 1 ]
	[[ "$output" == *"origin has no main"* ]]
}

@test "refuses to publish a commit that never left the workstation" {
	git tag "v$VERSION"
	serve
	commit "third"
	git tag --force "v$VERSION"

	run "$RELEASE"

	[ "$status" -eq 1 ]
	[[ "$output" == *"origin/main is a different commit"* ]]
}

@test "refuses to publish a tag that never left the workstation" {
	serve
	git tag "v$VERSION"

	run "$RELEASE"

	[ "$status" -eq 1 ]
	[[ "$output" == *"origin has no v$VERSION"* ]]
}

@test "refuses to publish when origin's tag names another commit" {
	git tag "v$VERSION"
	serve
	git --git-dir "$ORIGIN" update-ref "refs/tags/v$VERSION" "$(git rev-parse HEAD~1)"

	run "$RELEASE"

	[ "$status" -eq 1 ]
	[[ "$output" == *"names a different commit"* ]]
}

@test "refuses to publish when go.mod points away from origin" {
	git tag "v$VERSION"
	serve
	git remote set-url origin https://github.com/someone/else

	run "$RELEASE"

	[ "$status" -eq 1 ]
	[[ "$output" == *"is not origin"* ]]
}
