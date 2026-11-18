---
title: termlib
abstract: |-
  A minimalist terminal display library written as a Go module. It explores
  the space between fmt package and a rich library like tcell.
authors:
  - family_name: Doiel
    given_name: R. S.
    id: https://orcid.org/0000-0003-0900-6903



repository_code: https://github.com/rsdoiel/termlib
version: 0.0.11


programming_language:
  - Go >= 1.27.3


date_released: 2026-11-18
---

About this software
===================

## termlib 0.0.11

- tldemo now follows the workspace command line practice: -h, -l and -v short spellings beside -help, -license and -version; an unknown option or a surplus argument exits 2 with one plain line on standard error instead of the flag package's usage dump (before, a surplus argument was ignored and the demo ran); the help has an EXIT STATUS section. The program is a 12-line main over internal/cli.
- Go 1.27.3 is the minimum version; golang.org/x/term is v0.48.0 and golang.org/x/sys v0.50.0.
- Release process (workspace DR-0012): make release no longer runs generate or commits; check-committed fails a release unless the tree is committed and HEAD is pushed; release.bash has --dry-run, creates the draft at the pushed commit and commits nothing; the armv7l target is gone.
- The termlib package API is unchanged from v0.0.10. Source files are gofmt-clean and every one carries the full AGPL-3.0-or-later notice (four had it truncated, six had only a copyright line); the doc comments of Prompt, Truncate and KeyReader are corrected. The tests grew from 32.8% to 93% statement coverage, including LineEditor.Prompt driven through a real pseudo-terminal (Linux only), and make test reports coverage. ReadKey's comment now says an unrecognised escape sequence returns Key(0x1b), which is what the code does.

## Authors

- [R. S. Doiel](https://orcid.org/0000-0003-0900-6903)






A minimalist terminal display library written as a Go module. It explores
the space between fmt package and a rich library like tcell.


- [Code Repository](https://github.com/rsdoiel/termlib)
  - [Issue Tracker](https://github.com/rsdoiel/termlib/issues)

## Programming languages

- Go >= 1.27.3




## Software Requirements

- Go >= 1.27.3
- CMTools >= 0.0.45b


## Software Suggestions

- Pandoc >= 3.9
- GNU Make >= 3.8


