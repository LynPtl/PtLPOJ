package sandbox

import (
	"fmt"
	"strings"
)

// BuildExecutableCode combines the user's submitted pure Python code with the
// server's hidden doctest cases into a single executable payload payload.
// We use doctest in verbose mode, and if any test fails, we force a non-zero exit code (1).
//
// Security: user's print() calls are silenced during the user code phase,
// so they cannot pollute the doctest output that Go parses.
func BuildExecutableCode(userCode string, hiddenTests string) string {
	template := `
import doctest
import sys
import traceback
import builtins

# Silence user's print() calls by replacing the builtin
# They will be restored after the user code phase so doctest output is unaffected
_original_print = builtins.print
def _silenced_print(*args, **kwargs):
    pass  # discard all print output
builtins.print = _silenced_print

# === [ User Submitted Code START ] ===
%s
# === [ User Submitted Code END ] ===

# Restore print before running doctest so doctest's output goes to stdout
builtins.print = _original_print

# === [ Hidden Sandbox Injection START ] ===
try:
    if 'f' in globals():
        f.__doc__ = """
%s
        """
    else:
        print("ERROR: Function 'f' was not defined by the user.")
        sys.exit(2) # CE/RE indicator

    # Run the tests.
    # verbose=True emits clear text traces we can grep using Go
    results = doctest.testmod(verbose=True)
    if results.failed > 0:
        sys.exit(1)
    else:
        sys.exit(0)
except Exception as e:
    traceback.print_exc()
    sys.exit(2)
`
	safeTests := strings.ReplaceAll(hiddenTests, "\"\"\"", "\\\"")

	return fmt.Sprintf(template, userCode, safeTests)
}
