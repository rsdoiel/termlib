
# Action Items

## Modernization (2026-11-16)

To match the workspace standard (knowledge and opml are the models; Go 1.27.3 across repositories). Do one repository at a time; run the tests and `make -n release` after each step.

- [x] Go 1.27.3: change `go 1.26.2` in go.mod, the `Go >=` lines in codemeta.json, README.md and INSTALL.md; run `go mod tidy`, `go vet ./...` and the tests.
- [x] Makefile: add `generate` (cmt for version.go, CITATION.cff, about.md, installers); make `build` stop running cmt. Model: opml/Makefile.
- [x] Makefile: add `check-release` and make `release` run `generate` then `check-release`.
- [x] Makefile: add `check-go-version` (go.mod, codemeta.json and the Go in use agree); `check-release` calls it.
- [x] Drop the armv7l target and 32-bit Pi wording (dropped 2026-10-27).
- [x] publish.bash/publish.ps1: take knowledge's checked versions (a failed pull no longer publishes nothing silently).
- [x] release.bash: differs from fountain's only in two echo lines (termlib's are cleaner), so it stays.
- [x] release.ps1: take knowledge's (existing-draft check, checksums).

## Release 0.0.11 (2026-11-18)

- [x] Port DR-0012: `check-committed`, `release` without `generate`, `release.bash --dry-run`, `makefile_test.go`.
- [x] CLI to the workspace practice: `internal/cli`, `-h/-l/-v`, exit 2 on a bad command line, `cli_exit_test.go`.
- [x] gofmt the sources.
- [ ] Exported symbols carry one-line `//` comments, not the `/** */` blocks with parameters, returns and an example that CLAUDE.md asks for (about 60 symbols). Not release-blocking.

## Tests (2026-11-18)

- [x] Coverage 32.8% to 93%: keys, history, Terminal, FmtHelp, `internal/cli`, and `Prompt` through a pty (`pty_unix_test.go`; `golang.org/x/sys/unix` is a direct dependency).
- [x] `Terminal.Printf` formats like `fmt.Printf` with or without arguments (`%%` collapses). Decided with the author 2026-11-18.
- [x] The pty tests run on Linux and macOS (`pty_linux_test.go`, `pty_darwin_test.go` open the pair and read termios; macOS sets `PENDIN` itself, masked in the restore test).
- [x] `go test -race` run on the Mac Mini (darwin/arm64, Go 1.27.3): clean, three runs, pty tests included.
- [x] `UpdateTerminalSize` wrote the size without the lock; reproduced under `-race` (`TestResizeWhileDrawing`) and fixed.
- [x] Exported symbols carry the `/** */` blocks CLAUDE.md asks for. Left alone: `FmtHelp` and the other symbols in the generated `version.go`.

## Release 0.0.12 (2026-11-18)

- [x] Printf formats like fmt; UpdateTerminalSize race fixed; doc blocks; macOS pty tests; race run on the Mac Mini.
- [ ] Not blocking: a pty-driven `Prompt` test in harvey could copy `pty_unix_test.go`.
