---
title: termlib
abstract: "A minimalist terminal display library written as a Go module. It explores
the space between fmt package and a rich library like tcell."
authors:
  - family_name: Doiel
    given_name: R. S.
    id: https://orcid.org/0000-0003-0900-6903



repository_code: https://github.com/rsdoiel/termlib
version: 0.0.7


programming_language:
  - Go &gt;&#x3D; 1.25


date_released: 2026-05-12
---

About this software
===================

## termlib 0.0.7

- Ctrl+J (0x0a) now inserts \n into the buffer and moves to a new visual line with "...  " continuation prompt. Enter (\r) is the only submit key.
- redraw was refactored to be line-aware: it draws only the current visual line (from the last \n in buf to the next \n or end), choosing prompt for line 1 and "...  " for subsequent
  lines.
- Backspace across a \n clears the current visual line (\r\033[K), moves up (\033[1A), and redraws the merged previous line.
- Left/Right arrows, Home/End, Ctrl+A/E/K are all clamped to the current line — they won't jump across \n boundaries.
- History navigation (Up/Down) is disabled when lineCount > 0.

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
- CMTools >= 0.0.43


## Software Suggestions

- Pandoc >= 3.9
- GNU Make >= 3.8


