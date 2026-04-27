// rawmode.go — terminal raw mode management.
// Copyright (C) 2025 R. S. Doiel
package termlib

import (
	"os"

	"golang.org/x/term"
)

/** EnterRawMode puts the terminal associated with in into raw mode and
 * returns a restore function that the caller must defer. If the terminal
 * cannot be placed into raw mode (e.g. in is a pipe), an error is returned
 * and the restore function is a no-op.
 *
 * Parameters:
 *   in (*os.File) — the file to place in raw mode, typically os.Stdin.
 *
 * Returns:
 *   restore (func()) — call (or defer) this to return the terminal to its
 *                      original state.
 *   err (error)      — non-nil if raw mode could not be entered.
 *
 * Example:
 *   restore, err := termlib.EnterRawMode(os.Stdin)
 *   if err != nil {
 *       log.Fatal(err)
 *   }
 *   defer restore()
 */
func EnterRawMode(in *os.File) (restore func(), err error) {
	fd := int(in.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return func() {}, err
	}
	return func() { term.Restore(fd, oldState) }, nil
}
