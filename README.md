

# termlib

A minimalist terminal display library written as a Go module. It explores
the space between fmt package and a rich library like tcell.

## Release Notes

- version: 0.0.7
- status: concept
- released: 2026-05-12

- Ctrl+J (0x0a) now inserts \n into the buffer and moves to a new visual line with &quot;...  &quot; continuation prompt. Enter (\r) is the only submit key.
- redraw was refactored to be line-aware: it draws only the current visual line (from the last \n in buf to the next \n or end), choosing prompt for line 1 and &quot;...  &quot; for subsequent
  lines.
- Backspace across a \n clears the current visual line (\r\033[K), moves up (\033[1A), and redraws the merged previous line.
- Left/Right arrows, Home/End, Ctrl+A/E/K are all clamped to the current line — they won&#x27;t jump across \n boundaries.
- History navigation (Up/Down) is disabled when lineCount &gt; 0.


### Authors

- Doiel, R. S.



## Software Requirements

- Go >= 1.26
- CMTools >= 0.0.43

### Software Suggestions

- Pandoc >= 3.9
- GNU Make >= 3.8



## Related resources



- [Getting Help, Reporting bugs](https://github.com/rsdoiel/termlib/issues)

- [Installation](INSTALL.md)
- [About](about.md)

