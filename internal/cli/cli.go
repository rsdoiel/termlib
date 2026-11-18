// Copyright (C) 2026 R. S. Doiel
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

// Package cli holds the command line behavior of tldemo so that plain go test
// covers it. The program under cmd/ only calls Tldemo.
package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/rsdoiel/termlib"
)

// The exit codes tldemo uses (workspace DR-0003).
const (
	ExitOK    = 0
	ExitUsage = 2
)

/** Tldemo runs the tldemo program: it parses args, answers -help, -license
 * and -version, and otherwise runs the demo on stdout.
 *
 * Parameters:
 *   appName (string) — the name the program was run as
 *   args ([]string) — the command line without the program name
 *   stdout (io.Writer) — where help, license, version and the demo go
 *   stderr (io.Writer) — where a one line error goes
 *
 * Returns:
 *   int — the exit code: 0, or 2 for a bad command line
 *
 * Example:
 *   code := cli.Tldemo("tldemo", []string{"-version"}, os.Stdout, os.Stderr)
 */
func Tldemo(appName string, args []string, stdout, stderr io.Writer) int {
	var help, license, version bool
	fset := flag.NewFlagSet(appName, flag.ContinueOnError)
	fset.SetOutput(io.Discard)
	fset.BoolVar(&help, "h", false, "display help")
	fset.BoolVar(&help, "help", false, "display help")
	fset.BoolVar(&license, "l", false, "display license")
	fset.BoolVar(&license, "license", false, "display license")
	fset.BoolVar(&version, "v", false, "display version")
	fset.BoolVar(&version, "version", false, "display version")
	if err := fset.Parse(args); err != nil {
		fmt.Fprintf(stderr, "%s: %v (try %s -help)\n", appName, err, appName)
		return ExitUsage
	}
	switch {
	case help:
		fmt.Fprintln(stdout, termlib.FmtHelp(termlib.DemoHelpText, appName, termlib.Version, termlib.ReleaseDate, termlib.ReleaseHash))
		return ExitOK
	case license:
		fmt.Fprintln(stdout, strings.TrimSpace(termlib.LicenseText))
		return ExitOK
	case version:
		fmt.Fprintf(stdout, "%s %s %s\n", appName, termlib.Version, termlib.ReleaseHash)
		return ExitOK
	}
	if rest := fset.Args(); len(rest) > 0 {
		fmt.Fprintf(stderr, "%s: surplus argument %q (try %s -help)\n", appName, rest[0], appName)
		return ExitUsage
	}
	runDemo(stdout)
	return ExitOK
}
