package scheduler

import (
	"pt_lpoj/models"
	"pt_lpoj/sandbox"
	"pt_lpoj/storage"
	"testing"
	"time"
)

func TestJudgeWorkerIntegration(t *testing.T) {
	// 1. Initialize DB and Storage in testing mode
	err := storage.InitDB("file::memory:")
	if err != nil {
		t.Fatalf("Failed to init in-memory DB: %v", err)
	}

	err = storage.InitProblemRepo("../../data")
	if err != nil {
		t.Fatalf("Failed to load problems: %v", err)
	}

	// 3. Initialize Docker
	err = sandbox.InitDockerClient()
	if err != nil {
		t.Skipf("Skipping integration test; Docker not available: %v", err)
	}

	// 4. Create a dummy user
	user, _ := storage.CreateUser("worker@test.com", models.RoleStudent)

	// 5. Create a PENDING submission (let's use problem 1001, assume it's valid)
	// We'll submit a wrong answer to test full integration
	code := "def f(x):\n    return x + 2\n"
	sub, err := storage.CreateSubmission(user.ID, 1001, code)
	if err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}

	// 6. Start 1 Worker
	StartWorkerPool(1)

	// 7. Wait for worker to process
	timeout := time.After(15 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var finalSub *models.Submission
	done := false

	for !done {
		select {
		case <-timeout:
			t.Fatalf("Worker integration test timed out")
		case <-ticker.C:
			// check status
			finalSub, _ = storage.GetSubmissionByID(sub.ID)
			if finalSub != nil && finalSub.Status != models.StatusPending && finalSub.Status != models.StatusRunning {
				done = true
			}
		}
	}

	// 8. Validate Result
	if finalSub.Status != models.StatusWA {
		t.Errorf("Expected status %s, got %s", models.StatusWA, finalSub.Status)
	}
	if finalSub.ExecutionTimeMs == 0 {
		t.Errorf("Expected execution time to be recorded")
	}
	if finalSub.MemoryPeakKb == 0 {
		t.Errorf("Expected memory limit to be recorded")
	}
	if finalSub.FailedAtCase == 0 {
		t.Errorf("Expected failed case to be recorded for WA")
	}
}
