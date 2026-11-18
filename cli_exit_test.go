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

package termlib

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These tests run tldemo as a built binary, the way a user does. They cover
// the standard options and the exit codes of workspace DR-0003. The demo itself
// needs a terminal and is not run.

var builtTldemo string

func tldemo(t *testing.T) string {
	t.Helper()
	if builtTldemo != "" {
		return builtTldemo
	}
	dir, err := os.MkdirTemp("", "termlib-cli-")
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "tldemo")
	if out, err := exec.Command("go", "build", "-o", bin, "./cmd/tldemo").CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	builtTldemo = bin
	return bin
}

func TestMain(m *testing.M) {
	code := m.Run()
	if builtTldemo != "" {
		os.RemoveAll(filepath.Dir(builtTldemo))
	}
	os.Exit(code)
}

func runTldemo(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(tldemo(t), args...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	code := 0
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			t.Fatal(err)
		}
		code = ee.ExitCode()
	}
	return out.String(), errOut.String(), code
}

/** TestStandardOptions checks the short and long spellings of -help, -license
 * and -version: each prints to standard output, writes nothing to standard
 * error and exits 0. The help carries an EXIT STATUS section.
 */
func TestStandardOptions(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"-h"}, "EXIT STATUS"},
		{[]string{"-help"}, "EXIT STATUS"},
		{[]string{"-l"}, "GNU AFFERO GENERAL PUBLIC LICENSE"},
		{[]string{"-license"}, "GNU AFFERO GENERAL PUBLIC LICENSE"},
		{[]string{"-v"}, "tldemo "},
		{[]string{"-version"}, "tldemo "},
	} {
		out, errOut, code := runTldemo(t, tc.args...)
		if code != 0 || errOut != "" || !strings.Contains(out, tc.want) {
			t.Errorf("%v: code %d, stderr %q, stdout lacks %q", tc.args, code, errOut, tc.want)
		}
	}
}

/** TestUsageErrors checks that a bad command line attempts nothing and exits
 * 2 with one plain line on standard error and nothing on standard output
 * (DR-0003, DR-0011). Before this was fixed a surplus argument was ignored and
 * the demo ran.
 */
func TestUsageErrors(t *testing.T) {
	for _, args := range [][]string{{"-nosuchoption"}, {"surplus"}} {
		out, errOut, code := runTldemo(t, args...)
		if code != 2 {
			t.Errorf("%v: exit %d, want 2", args, code)
		}
		if out != "" {
			t.Errorf("%v: stdout %q, want empty", args, out)
		}
		if !strings.HasPrefix(errOut, "tldemo: ") || strings.Count(strings.TrimRight(errOut, "\n"), "\n") != 0 {
			t.Errorf("%v: stderr %q, want one line prefixed with tldemo:", args, errOut)
		}
	}
}
