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

// tldemo is part of termlib, a minimalist terminal display library for Go.
// Run tldemo -help.
package main

import (
	"os"
	"path"

	"github.com/rsdoiel/termlib/internal/cli"
)

func main() {
	os.Exit(cli.Tldemo(path.Base(os.Args[0]), os.Args[1:], os.Stdout, os.Stderr))
}
