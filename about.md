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
version: 0.0.10


programming_language:
  - Go >= 1.25


date_released: 2026-09-27
---

About this software
===================

## termlib 0.0.10

- Bug fix: LineEditor.Prompt() corrupted the display when the prompt was wider than the terminal or contained an embedded newline. redraw()'s per-keystroke cursor math assumes the prompt fits on one terminal row; a wider prompt broke that assumption, since "\r" only returns to column 0 of the terminal's current row, so reprinting a multi-row prompt on every keystroke pushed the display down further each time and left typed input effectively invisible. splitSafePrompt() now prints anything up to and including the prompt's last embedded newline once, up front, and hands only a genuinely one-row-safe tail to redraw().

## Authors

- [R. S. Doiel](https://orcid.org/0000-0003-0900-6903)






A minimalist terminal display library written as a Go module. It explores
the space between fmt package and a rich library like tcell.


- [Code Repository](https://github.com/rsdoiel/termlib)
  - [Issue Tracker](https://github.com/rsdoiel/termlib/issues)

## Programming languages

- Go >= 1.25




## Software Requirements

- Go >= 1.26
- CMTools >= 0.0.45b


## Software Suggestions

- Pandoc >= 3.9
- GNU Make >= 3.8


