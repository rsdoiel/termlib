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

-h, -help
: Display this help

-v, -version
: Display {app_name} version

-l, -license
: Display {app_name} license

# EXIT STATUS

0
: Success.

2
: Usage: an unknown option or a surplus argument. Nothing was run.


`
)
