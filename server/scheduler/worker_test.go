package scheduler

import (
	"pt_lpoj/models"
	"pt_lpoj/sandbox"
	"pt_lpoj/storage"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJudgeWorkerIntegration(t *testing.T) {
	err := storage.InitDB("file::memory:")
	if err != nil {
		t.Fatalf("Failed to init in-memory DB: %v", err)
	}

	err = storage.InitProblemRepo("../../data")
	if err != nil {
		t.Fatalf("Failed to load problems: %v", err)
	}

	err = sandbox.InitDockerClient()
	if err != nil {
		t.Skipf("Skipping integration test; Docker not available: %v", err)
	}

	user, _ := storage.CreateUser("worker@test.com", models.RoleStudent)

	code := "def f(x):\n    return x + 2\n"
	sub, err := storage.CreateSubmission(user.ID, 1001, code)
	if err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}

	queue := NewJobQueue(10)
	SetGlobalQueue(queue)
	StartWorkerPool(1, queue)

	queue.Enqueue(sub.ID)

	resultCh := queue.Subscribe(sub.ID)
	defer queue.Unsubscribe(sub.ID)

	select {
	case <-time.After(15 * time.Second):
		t.Fatalf("Worker integration test timed out")
	case result := <-resultCh:
		if result.Status != models.StatusWA {
			t.Errorf("Expected status %s, got %s", models.StatusWA, result.Status)
		}
		if result.ExecutionTimeMs == 0 {
			t.Errorf("Expected execution time to be recorded")
		}
		if result.FailedCase == 0 {
			t.Errorf("Expected failed case to be recorded for WA")
		}
		dbSub, _ := storage.GetSubmissionByID(sub.ID)
		if dbSub.Status != models.StatusWA {
			t.Errorf("DB status mismatch: expected %s, got %s", models.StatusWA, dbSub.Status)
		}
	}
}

func TestJobQueueUnsubscribe(t *testing.T) {
	queue := NewJobQueue(10)
	subID := uuid.New()

	ch := queue.Subscribe(subID)
	queue.Unsubscribe(subID)

	select {
	case _, ok := <-ch:
		if ok {
			t.Errorf("Expected closed channel")
		}
	default:
	}
}

func TestJobQueueClose(t *testing.T) {
	queue := NewJobQueue(10)
	queue.Close()
}

func TestJobQueueEnqueueNonBlocking(t *testing.T) {
	queue := NewJobQueue(10)
	subID := uuid.New()

	select {
	case queue.jobs <- subID:
	default:
		t.Errorf("Expected non-blocking enqueue")
	}
	queue.Close()
}
