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

package cli

import (
	"bytes"
	"strings"
	"testing"
)

func runCLI(args ...string) (string, string, int) {
	var out, errOut bytes.Buffer
	code := Tldemo("tldemo", args, &out, &errOut)
	return out.String(), errOut.String(), code
}

/** TestTldemoAnswersWithoutRunningTheDemo checks help, license and version in
 * process: exit 0, the text on standard output, nothing on standard error, and
 * no terminal escape sequences (the demo clears the screen first).
 */
func TestTldemoAnswersWithoutRunningTheDemo(t *testing.T) {
	for _, tc := range []struct{ arg, want string }{
		{"-h", "tldemo(1)"},
		{"-help", "# EXIT STATUS"},
		{"-l", "GNU AFFERO GENERAL PUBLIC LICENSE"},
		{"-v", "tldemo "},
	} {
		out, errOut, code := runCLI(tc.arg)
		if code != ExitOK || errOut != "" || !strings.Contains(out, tc.want) || strings.Contains(out, "\033") {
			t.Errorf("%s: code %d, stderr %q, stdout %q (want to contain %q, no ESC)", tc.arg, code, errOut, out, tc.want)
		}
	}
}

/** TestTldemoHelpHasNoPlaceholders checks that FmtHelp filled in every
 * {placeholder} of the help text.
 */
func TestTldemoHelpHasNoPlaceholders(t *testing.T) {
	out, _, _ := runCLI("-help")
	if strings.Contains(out, "{") {
		t.Errorf("help still has a placeholder:\n%s", out)
	}
}

/** TestTldemoUsageErrors checks that a bad command line returns ExitUsage and
 * writes one line to standard error and nothing to standard output.
 */
func TestTldemoUsageErrors(t *testing.T) {
	for _, args := range [][]string{{"-bogus"}, {"extra"}, {"-h=maybe"}} {
		out, errOut, code := runCLI(args...)
		if code != ExitUsage || out != "" || !strings.HasPrefix(errOut, "tldemo: ") || strings.Count(strings.TrimSpace(errOut), "\n") != 0 {
			t.Errorf("%v: code %d, stdout %q, stderr %q", args, code, out, errOut)
		}
	}
}
