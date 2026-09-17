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
$ tooluse-screener 'rm -rf /'          # prints `allow|deny|ask: reason`
deny: Command matches a denied rule: Emptying a whole tree
```

The exit code carries the same answer: 0 allowed, 1 denied, 2 to ask.

```bash
echo "$payload" | tooluse-screener --hook   # answers the agent that asked
```

Each command is checked against a [policy](#configuration), and one
`deny` blocks the whole line it was written on.

## Installation

```bash
mise use github:thehale/tooluse-screener
```

Point the agent's pre-tool hook at `tooluse-screener --hook`, which
reads the tool-use payload on stdin and answers in the envelope that
agent expects.

```json
{
  "hooks": {
    "PreToolUse": [
      { "hooks": [{ "type": "command", "command": "tooluse-screener --hook" }] }
    ]
  }
}
```

Claude Code and Codex are both recognised. Claude Code enforces a `deny`
on `PreToolUse`; Codex reads both `PreToolUse` and `PermissionRequest`
and enforces on the first, so point it at both. A `deny` is written to
stderr as well as into the envelope, because Codex ignores one that
carries no reason there. A payload holding no Bash command is met with
silence, and so is a verdict of `ask`, which leaves the agent to prompt
as it normally would.

Commands are matched as text rather than parsed as a shell, so quoting
defeats the matching. A policy is a guard rail for an agent that means
well, not a sandbox for one that does not.

## Configuration

### Policy File

The policy says which commands are allowed and which are denied. It is
read from the first of these that answers:

| Policy                          | Where                    |
| ------------------------------- | ------------------------ |
| `--config-file PATH`            | The command line         |
| `$TOOLUSE_SCREENER_POLICY_FILE` | The environment          |
| `tooluse-screener/policy.yaml`  | Your platform's config¹  |
| The built-in one                | Compiled into the binary |

¹ `$XDG_CONFIG_HOME` or `~/.config` on Linux,
`~/Library/Application Support` on macOS, `%AppData%` on Windows.

Only the highest one applies. No merging, no inheritance, so a policy of
your own replaces the built-in one rather than adding to it.

A policy that will not parse enforces nothing and says why on stderr.
Check one after editing it:

```bash
tooluse-screener --config-file PATH ls
```

### Syntax

A policy has three blocks, each described below: `denied` and `allowed`
hold rules, and `trusted_git_directories` binds git to the repositories
you name.

#### Rules

A rule is a command written plainly, or a mapping naming several:

```yaml
- ls
- commands: ls
- commands: [ls, cat]
- patterns: '^ls\b'
- patterns: ['^ls\b', '^cat\b']
```

`commands` are literal text. `patterns` are regexes, specifically
[RE2](https://github.com/google/re2/wiki/Syntax), which has no
lookarounds and reads `\b` as ASCII.

Rules check the entire command for whole-word matches, so a rule about
`ls` matches `ENV=thing ls -la` and not `lsblk`. An allowed rule has to
be the command itself rather than merely appear in it, so it says
nothing about `rm ls`.

Add a `description`, which is what the rule calls itself in a verdict,
and a `reason`, which is shown alongside it:

```yaml
- description: Emptying a whole tree
  reason: Ask a human first.
  patterns: ['^rm -(rf|fr) /$']
```

An `only` block holds a rule to where it applies. A rule whose `only` is
not satisfied does not match at all:

```yaml
- commands: git push
  only:
    dirs: [~/src/one-repo]
    branches:
      - ci-fix
      - not: [main, master, trunk]
```

`dirs` holds the rule to those directories, read against every directory
the command names with `-C`, `--git-dir` or `--work-tree`, or against
the working directory when it names none.

`branches` holds a push to the branches it lists. A plain name is one
the push may land on, and `not` names one it may not. Where both are
written, both hold. This is read from where the commits will actually
land rather than from the words, which is why it is a restriction and
not part of the expression beside it.

##### Rule Precedence

The basic rules are straightforward:

- If **any** commands match a `denied` rule, the whole compound command
  is denied.

  ```bash
  friendly && scary  # whole command denied
  ```

- If **every** command matches an `allowed` rule, the whole command is allowed.

  ```bash
  friendly && cheerful  # whole command approved
  ```

- Everything else is returned for the model to "ask" you or an auto-evaluator.

  ```bash
  friendly && unknown  # "ask" for permission.
  ```

When `allowed` and `denied` rules both match the same command, conflict
resolution occurs.

```bash
friendly -exec "scary"
```

- To be safe, if a `denied` rule matches **any** part of the command,
  the denial wins ...
- ... UNLESS, there's an explicit `allowed` rule that matches the
  **entire** command.
- In the case of a tie, denial wins.

#### Trusted Git Directories

```yaml
trusted_git_directories: [~/src]
```

This is where an agent may reach out to a repository other than the one
it is standing in. An allowed git rule holds only where the command
points inside one of these, so `git -C /elsewhere status` is left to be
asked about however plainly `git status` is allowed.

It is not the same question as `only.dirs`, and a git command has to get
past both. This names the repositories an agent may touch at all.
`only.dirs` names where a single rule applies.

#### Example

```yaml
trusted_git_directories:
  - ~/src

denied:
  - description: Pushing
    reason: A push is the operator's to approve.
    commands: git push

allowed:
  - ls

  - description: Pushing where CI can run against the commit
    patterns: ['^git push origin \S+$']
    only:
      dirs: [~/src/one-repo]
      branches:
        - not: [main, master, trunk]
```

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
