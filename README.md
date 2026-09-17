<div align="center">

# tooluse-screener

One Bash command policy, shared by every coding agent.

<!-- BADGES -->
[![License: MPL-2.0](https://badgen.net/github/license/thehale/tooluse-screener)](https://github.com/thehale/tooluse-screener/blob/main/LICENSE)
[![Sponsor thehale on GitHub](https://badgen.net/badge/icon/Sponsor/pink?icon=github&label)](https://github.com/sponsors/thehale)
[![Joseph Hale's software engineering blog](https://jhale.dev/badges/website.svg)](https://jhale.dev)
[![Follow Joseph Hale on LinkedIn](https://jhale.dev/badges/follow.svg)](https://www.linkedin.com/comm/mynetwork/discovery-see-all?usecase=PEOPLE_FOLLOWS&followMember=thehale)
</div>

## Quickstart

```console
$ tooluse-screener "git status"
allow: Every command is allowed: git status

$ tooluse-screener "git commit -m 'a message' && git status"
allow: Every command is allowed: git commit, git status

$ tooluse-screener "rm -rf /"
deny: Command matches a denied rule: Emptying a whole tree

$ tooluse-screener "nmap localhost"
ask: Command is not in the shared allow list
```

One refused command denies the whole line. Every command on it has to be
vouched for before any of it is allowed, and anything left over is the
agent's own question to ask. The exit code carries the same answer: 0
allowed, 1 denied, 2 to ask.

## Installation

```bash
mise use github:thehale/tooluse-screener
```

Point each agent's Bash hook at `tooluse-screener --hook`, which reads
the payload on stdin and answers in the shape that agent expects —
`PreToolUse` for Claude Code, and both `PreToolUse` and
`PermissionRequest` for Codex, which only enforces a denial on the
first.

## Configuration

What ships covers git, and a few things nobody wants run by accident.
Write the rest in a file of your own, which **replaces** the shipped one
rather than adding to it — the first of these that answers is the only
one read:

| Policy                          | Where                    |
| ------------------------------- | ------------------------ |
| `--config-file PATH`            | The command line         |
| `$TOOLUSE_SCREENER_POLICY_FILE` | The environment          |
| `tooluse-screener/policy.yaml`  | Your platform's config¹  |
| The built-in one                | Compiled into the binary |

¹ `$XDG_CONFIG_HOME` or `~/.config` on Linux,
`~/Library/Application Support` on macOS, `%AppData%` on Windows.

```yaml
trusted_git_directories:
  - ~/src

denied:
  - git clone
  - description: Cutting or moving a GitHub release
    reason: Cut one with the publish script instead.
    patterns:
      - '(?:^|[\s/])gh(?:\s+\S+)*?\s+release\s+(?:create|delete)\b'

allowed:
  - git status
  - ls
```

An entry is a command, or a group of `commands` and `patterns` sharing a
`reason`, shown with the verdict, and a `description`, which is what the
rules call themselves there. `commands` are literal text; `patterns` are
[RE2](https://github.com/google/re2/wiki/Syntax), which has no
lookarounds and reads `\b` as ASCII.

A denied entry is looked for anywhere in a command, so a deploy is
caught behind the environment variable preceding it. An allowed one has
to open the command and end at a word boundary, so `ls` allows `ls -la`
and says nothing about `lsblk`. An entry starting with `git` covers
every spelling of that operation, and an allowed one holds only where it
points inside `trusted_git_directories`.

A policy that will not parse enforces nothing and says why on stderr, so
check one with `tooluse-screener --config-file PATH ls` after editing it.

## Contributing

```bash
bin/setup  # Install the tools
bin/ci     # Run the checks
bin/ci --fix  # Fix what can be fixed automatically
```

## License

Copyright (c) 2026 Joseph Hale, All Rights Reserved

Provided under the terms of the [Mozilla Public License, version 2.0](./LICENSE)

<details>

<summary><b>What does the MPL-2.0 license allow/require?</b></summary>

### TL;DR

You can use files from this project in both open source and proprietary
applications, provided you include the above attribution. However, if
you modify any code in this project, or copy blocks of it into your own
code, you must publicly share the resulting files (note, not your whole
program) under the MPL-2.0. The best way to do this is via a Pull
Request back into this project.

If you have any other questions, you may also find Mozilla's [official
FAQ](https://www.mozilla.org/en-US/MPL/2.0/FAQ/) for the MPL-2.0 license
insightful.

If you dislike this license, you can contact me about negotiating a paid
contract with different terms.

**Disclaimer:** This TL;DR is just a summary. All legal questions
regarding usage of this project must be handled according to the
official terms specified in the `LICENSE` file.

### Why the MPL-2.0 license?

I believe that an open-source software license should ensure that code
can be used everywhere.

Strict copyleft licenses, like the GPL family of licenses, fail to
fulfill that vision because they only permit code to be used in other
GPL-licensed projects. Permissive licenses, like the MIT and Apache
licenses, allow code to be used everywhere but fail to prevent
proprietary or GPL-licensed projects from limiting access to any
improvements they make.

In contrast, the MPL-2.0 license allows code to be used in any software
project, while ensuring that any improvements remain available for
everyone.

</details>
