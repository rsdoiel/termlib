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
version: 0.0.12


programming_language:
  - Go >= 1.27.3


date_released: 2026-11-18
---

About this software
===================

## termlib 0.0.12

- Behavior change: Terminal.Printf now formats exactly as fmt.Printf does. It used to skip formatting when it had no arguments, so Printf("100%% done") printed "100%% done"; it now prints "100% done". A caller that passed a literal containing a lone percent sign and no arguments should use Print.
- Bug fix: Terminal.UpdateTerminalSize wrote the stored size without taking the lock every other method takes, so calling it from a resize handler while another goroutine drew or read the size was a data race (reproduced under go test -race, now fixed). Terminal is safe to use from several goroutines.
- Every exported symbol outside the generated version.go now has a documentation block with its parameters, return values and an example. ReadKey's comment now says an unrecognised escape sequence returns Key(0x1b), which is what the code does.
- Tests: the LineEditor.Prompt pseudo-terminal tests now run on macOS as well as Linux, and go test -race is clean on macOS (darwin/arm64, Go 1.27.3).
- The API is otherwise unchanged from v0.0.11.

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


