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

//go:build darwin

package termlib

import (
	"testing"
	"unsafe"

	"golang.org/x/sys/unix"
)

// openPtyPair opens a pty and returns the master and slave descriptors. macOS
// grants and unlocks the slave with ioctls and returns its name in a 128 byte
// buffer.
func openPtyPair(t *testing.T) (master, slave int, err error) {
	t.Helper()
	master, err = unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		return 0, 0, err
	}
	if err = unix.IoctlSetInt(master, unix.TIOCPTYGRANT, 0); err != nil {
		return 0, 0, err
	}
	if err = unix.IoctlSetInt(master, unix.TIOCPTYUNLK, 0); err != nil {
		return 0, 0, err
	}
	var name [128]byte
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(master), uintptr(unix.TIOCPTYGNAME), uintptr(unsafe.Pointer(&name[0]))); errno != 0 {
		return 0, 0, errno
	}
	n := 0
	for n < len(name) && name[n] != 0 {
		n++
	}
	slave, err = unix.Open(string(name[:n]), unix.O_RDWR|unix.O_NOCTTY, 0)
	return master, slave, err
}

// stableTermios returns the settings with the bits the kernel manages dropped, so
// that two reads can be compared. macOS sets PENDIN itself when a terminal leaves
// raw mode; it is not a setting the editor changed.
func stableTermios(t *unix.Termios) unix.Termios {
	c := *t
	c.Lflag &^= unix.PENDIN
	return c
}

// getTermios reads the terminal settings of fd.
func getTermios(fd int) (*unix.Termios, error) {
	return unix.IoctlGetTermios(fd, unix.TIOCGETA)
}
