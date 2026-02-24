package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProblemRepo(t *testing.T) {
	// Let's test against the real generated data structure in PtLPOJ/data
	// Assuming the test runs from within PtLPOJ/server/storage

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	dataDir := filepath.Join(filepath.Dir(wd), "..", "data")

	err = InitProblemRepo(dataDir)
	if err != nil {
		// If problems.json is missing, skip the test safely instead of crashing
		if os.IsNotExist(err) || err.Error() == "problems.json not found at "+filepath.Join(dataDir, "problems", "problems.json")+". Please generate it first" {
			t.Skipf("Skipping integration test because problems.json is not yet generated: %v", err)
			return
		}
		t.Fatalf("Failed to initialize ProblemRepo: %v", err)
	}

	// Verify we loaded something
	problems := GetAllProblems()
	if len(problems) == 0 {
		t.Errorf("Expected problems to be loaded, got 0")
	}

	// Assume problem 1001 exists as per our python script
	p, err := GetProblemByID(1001)
	if err != nil {
		t.Errorf("Failed to get problem 1001: %v", err)
	}
	if p != nil && p.Title == "" {
		t.Errorf("Loaded problem has no title")
	}

	// Test GetProblemFile
	scaffold, err := GetProblemFile(1001, "scaffold.py")
	if err != nil {
		t.Errorf("Failed to read scaffold.py: %v", err)
	}
	if len(scaffold) < 5 {
		t.Errorf("Read scaffold.py seems incorrectly short or empty")
	}
}
