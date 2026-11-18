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

//go:build linux

package termlib

import (
	"strconv"
	"testing"

	"golang.org/x/sys/unix"
)

// openPtyPair opens a pty and returns the master and slave descriptors.
func openPtyPair(t *testing.T) (master, slave int, err error) {
	t.Helper()
	master, err = unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		return 0, 0, err
	}
	if err = unix.IoctlSetPointerInt(master, unix.TIOCSPTLCK, 0); err != nil {
		return 0, 0, err
	}
	n, err := unix.IoctlGetInt(master, unix.TIOCGPTN)
	if err != nil {
		return 0, 0, err
	}
	slave, err = unix.Open("/dev/pts/"+strconv.Itoa(n), unix.O_RDWR|unix.O_NOCTTY, 0)
	return master, slave, err
}

// stableTermios returns the settings with the bits the kernel manages dropped, so
// that two reads can be compared. Linux has none.
func stableTermios(t *unix.Termios) unix.Termios { return *t }

// getTermios reads the terminal settings of fd.
func getTermios(fd int) (*unix.Termios, error) {
	return unix.IoctlGetTermios(fd, unix.TCGETS)
}
