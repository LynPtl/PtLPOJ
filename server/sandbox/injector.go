package sandbox

import (
	"fmt"
	"strings"
)

// BuildExecutableCode combines the user's submitted pure Python code with the
// server's hidden doctest cases into a single executable payload payload.
// We use doctest in verbose mode, and if any test fails, we force a non-zero exit code (1).
func BuildExecutableCode(userCode string, hiddenTests string) string {
	// The injection template wraps the user logic, injects the hidden tests into
	// the docstring of function 'f', and then rigorously runs the test suite.

	// We handle the tricky issue of injecting the docs: since the user's code might
	// redefine 'f' entirely, we inject the __doc__ dynamically *after* their code runs.

	template := `
import doctest
import sys
import traceback

# === [ User Submitted Code START ] ===
%s
# === [ User Submitted Code END ] ===

# === [ Hidden Sandbox Injection START ] ===
try:
    # Attempt to attach the hidden tests safely.
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
        # One or more tests failed. Force exit 1 (WA)
        sys.exit(1)
    else:
        # All passed
        sys.exit(0)
except Exception as e:
    # Explicitly catch syntax/runtime errors during the user load phase
    traceback.print_exc()
    sys.exit(2)
`
	// Sanitize tests string slightly so it doesn't break the triple quotes
	safeTests := strings.ReplaceAll(hiddenTests, "\"\"\"", "\\\"\\\"\\\"")

	return fmt.Sprintf(template, userCode, safeTests)
}
