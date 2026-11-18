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
	"unicode/utf8"
)

/** Key represents a single keystroke. For printable ASCII and Unicode
 * characters the value equals the rune value, so Key('q') is the letter q.
 * Control characters retain their ASCII values (e.g. Key(0x03) is Ctrl+C).
 * Special keys such as arrow keys use values >= 1000 defined as constants
 * below.
 *
 * Example:
 *   k, _ := termlib.ReadKey(os.Stdin)
 *   switch k {
 *   case Key(' '):          // space bar
 *   case termlib.KeyUp:     // up arrow
 *   case Key(0x03):         // Ctrl+C
 *   }
 */
type Key rune

const (
	// Special keys — all >= 1000 to avoid collision with printable characters.
	KeyUnknown  Key = 1000 + iota
	KeyUp           // up arrow
	KeyDown         // down arrow
	KeyLeft         // left arrow
	KeyRight        // right arrow
	KeyHome         // Home key
	KeyEnd          // End key
	KeyPageUp       // Page Up key
	KeyPageDown     // Page Down key
)

/** ReadKey reads exactly one keystroke from in, which must already be in raw
 * mode (see EnterRawMode). Multi-byte escape sequences (arrow keys, etc.)
 * are consumed and returned as a single Key constant. Multi-byte UTF-8
 * printable characters are decoded and returned as Key(rune).
 *
 * Parameters:
 *   in (*os.File) — input file in raw mode, typically os.Stdin.
 *
 * Returns:
 *   Key   — the keystroke. An escape sequence that is not recognised (Delete,
 *           function keys, a lone ESC) is consumed and returned as Key(0x1b).
 *           KeyUnknown is returned for invalid UTF-8 and together with an error.
 *   error — non-nil on I/O failure or EOF.
 *
 * Example:
 *   restore, _ := termlib.EnterRawMode(os.Stdin)
 *   defer restore()
 *   k, err := termlib.ReadKey(os.Stdin)
 */
func ReadKey(in *os.File) (Key, error) {
	b := make([]byte, 1)
	if _, err := in.Read(b); err != nil {
		return KeyUnknown, err
	}

	switch {
	case b[0] == 0x1b: // ESC or escape sequence
		seq := readKeyEscSeq(in)
		switch seq {
		case "[A", "OA":
			return KeyUp, nil
		case "[B", "OB":
			return KeyDown, nil
		case "[C", "OC":
			return KeyRight, nil
		case "[D", "OD":
			return KeyLeft, nil
		case "[H", "OH", "[1~":
			return KeyHome, nil
		case "[F", "OF", "[4~":
			return KeyEnd, nil
		case "[5~":
			return KeyPageUp, nil
		case "[6~":
			return KeyPageDown, nil
		default:
			return Key(0x1b), nil // bare ESC
		}

	case b[0] >= 0xc0: // UTF-8 multi-byte lead byte
		r := readKeyUTF8Tail(in, b[0])
		if r == utf8.RuneError {
			return KeyUnknown, nil
		}
		return Key(r), nil

	default:
		return Key(b[0]), nil
	}
}

/** KeyReader starts a goroutine that reads keystrokes from in and sends them
 * on the returned channel. in must already be in raw mode. The goroutine
 * exits and closes the channel when in returns an error (including EOF).
 * There is no way to stop it: if the caller stops receiving, the goroutine
 * stays blocked in ReadKey or on a full channel until in fails, so call
 * KeyReader once per input and keep receiving, or use ReadKey directly.
 *
 * Parameters:
 *   in (*os.File) — input file in raw mode, typically os.Stdin.
 *
 * Returns:
 *   <-chan Key — receive keystrokes from this channel.
 *
 * Example:
 *   restore, _ := termlib.EnterRawMode(os.Stdin)
 *   defer restore()
 *   keys := termlib.KeyReader(os.Stdin)
 *   for k := range keys {
 *       if k == termlib.Key('q') { break } // the reader goroutine remains
 *   }
 */
func KeyReader(in *os.File) <-chan Key {
	ch := make(chan Key, 8)
	go func() {
		defer close(ch)
		for {
			k, err := ReadKey(in)
			if err != nil {
				return
			}
			ch <- k
		}
	}()
	return ch
}

// readKeyEscSeq reads the bytes following an ESC and returns a short string
// identifying the sequence, e.g. "[A" for up-arrow.
func readKeyEscSeq(in *os.File) string {
	b := make([]byte, 1)
	if _, err := in.Read(b); err != nil {
		return ""
	}
	switch b[0] {
	case '[': // CSI sequence
		var seq []byte
		for {
			if _, err := in.Read(b); err != nil {
				break
			}
			seq = append(seq, b[0])
			if (b[0] >= 'A' && b[0] <= 'Z') || (b[0] >= 'a' && b[0] <= 'z') || b[0] == '~' {
				break
			}
		}
		return "[" + string(seq)
	case 'O': // SS3 sequence (application cursor keys)
		if _, err := in.Read(b); err != nil {
			return ""
		}
		return "O" + string(b[:1])
	default:
		return ""
	}
}

// readKeyUTF8Tail reads continuation bytes for a multi-byte UTF-8 sequence
// whose lead byte is lead, and returns the decoded rune.
func readKeyUTF8Tail(in *os.File, lead byte) rune {
	var tailLen int
	switch {
	case lead >= 0xf0:
		tailLen = 3
	case lead >= 0xe0:
		tailLen = 2
	default:
		tailLen = 1
	}
	tail := make([]byte, tailLen)
	in.Read(tail)
	full := append([]byte{lead}, tail...)
	r, _ := utf8.DecodeRune(full)
	return r
}
