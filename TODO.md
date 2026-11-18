
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

- [x] Coverage 32.8% to 93%: keys, history, Terminal, FmtHelp, `internal/cli`, and `Prompt` through a pty (`pty_linux_test.go`, Linux only, uses `golang.org/x/sys/unix`, now a direct dependency).
- [ ] `Terminal.Printf` with no arguments skips formatting, so `Printf("100%% done")` prints `100%% done`. Decide whether `%%` should collapse; not pinned by a test.
- [ ] The pty test does not run on macOS. Its ioctls differ (`TIOCPTYGRANT`, `TIOCPTYUNLK`, `TIOCPTYGNAME`); add a `pty_darwin_test.go` if the Mac Mini should cover `Prompt`.
- [ ] `go test -race` has not been run on the editor (the Pi's race runtime fails); run it on the Mac Mini for the platform-independent tests.
