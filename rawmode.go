// Copyright (C) 2025-2026 R. S. Doiel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
