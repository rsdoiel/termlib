/*
termlib is a light weight terminal interface library. Sort of a ncurse light.
Copyright (C) 2025 R. S. Doiel

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
*/
package termlib

const (
	DemoHelpText = `%{app_name}(1) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

{app_name}

# SYNOPSIS

{app_name} [OPTIONS]

# DESCRIPTION

**{app_name}** demonstrate the simple terminal UI that can be implemented with termlib. This is a bare bones approach not a full TUI package like tcell.

# OPTIONS

-help,
: Display this help

-version
: Display {app_name} version

-license
: Display {app_name} license


`
)
