package sandbox

import (
	"testing"
)

func TestSandboxRunner(t *testing.T) {
	err := InitDockerClient()
	if err != nil {
		t.Skipf("Skipping sandbox tests, docker is not available: %v", err)
	}

	// Case 1: Safe AC code
	userCode := "def f(x):\n    return x + 1\n"
	hiddenTest := ">>> f(1)\n2\n>>> f(2)\n3\n"

	payload := BuildExecutableCode(userCode, hiddenTest)

	res, err := RunCode("test_ac_01", payload, 2000, 32768)
	if err != nil {
		t.Fatalf("Failed to run safe code: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("Expected ExitCode 0, got %d. Stdout: %s", res.ExitCode, res.Stdout)
	}
	if res.FailedAtCase != 0 {
		t.Errorf("Expected 0 failed cases, got %d", res.FailedAtCase)
	}

	// Case 2: Time Limit Exceeded (Infinite Loop)
	loopCode := "def f(x):\n    while True:\n        pass\n"
	payload2 := BuildExecutableCode(loopCode, hiddenTest)
	res2, err := RunCode("test_tle_02", payload2, 1000, 32768)
	if err != nil {
		t.Fatalf("Failed to run code: %v", err)
	}
	// It should be killed (ExitCode 137 or 124 timeout equivalent)
	if res2.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code for TLE, got 0")
	}

	// Case 3: Wrong Answer
	waCode := "def f(x):\n    return x + 2\n"
	payload3 := BuildExecutableCode(waCode, hiddenTest)
	res3, err := RunCode("test_wa_03", payload3, 2000, 32768)
	if err != nil {
		t.Fatalf("Failed to run code: %v", err)
	}
	if res3.ExitCode == 0 {
		t.Errorf("WA code should not have exit code 0")
	}
	if res3.FailedAtCase != 1 {
		t.Errorf("Expected to fail at case 1, got %d", res3.FailedAtCase)
	}

	// Case 4: Network Isolation (urllib fetch)
	netCode := "import urllib.request\ndef f(x):\n    urllib.request.urlopen('http://example.com', timeout=1)\n    return x\n"
	payload4 := BuildExecutableCode(netCode, hiddenTest)
	res4, err := RunCode("test_net_04", payload4, 2000, 32768)
	if err != nil {
		t.Fatalf("Failed to run code: %v", err)
	}
	if res4.ExitCode == 0 {
		t.Errorf("Network code should have crashed due to isolation")
	}
	// Note: We don't check exact verbose strings because doctest intercepts raw urllib
	// stacktraces into a standard generic "Failed example" trace when wrapped.
}
