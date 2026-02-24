package scheduler

import (
	"fmt"
	"log"
	"pt_lpoj/models"
	"pt_lpoj/sandbox"
	"pt_lpoj/storage"
	"time"
)

// StartWorkerPool initializes and starts a given number of worker goroutines
// to consume PENDING submissions from the database.
func StartWorkerPool(workerCount int) {
	// 0. Crash Recovery Hook
	recoverOrphanedSubmissions()

	for i := 0; i < workerCount; i++ {
		go judgeWorker(i)
	}
	log.Printf("Started %d Judge Workers", workerCount)
}

// recoverOrphanedSubmissions reverts any 'RUNNING' submissions back to 'PENDING'.
// This is critical if the main server process crashed or was killed while workers were executing tasks,
// ensuring those requests are eventually evaluated.
func recoverOrphanedSubmissions() {
	res := storage.DB.Model(&models.Submission{}).
		Where("status = ?", models.StatusRunning).
		Update("status", models.StatusPending)

	if res.RowsAffected > 0 {
		log.Printf("[Crash Recovery] Reverted %d orphaned submissions back to PENDING queue", res.RowsAffected)
	}
}

func judgeWorker(workerID int) {
	for {
		// 1. Fetch the oldest PENDING submission
		var submission models.Submission
		// Fetch one PENDING submission
		result := storage.DB.Where("status = ?", models.StatusPending).
			Order("created_at asc").
			First(&submission)

		if result.Error != nil {
			// No pending submissions, sleep and poll again
			time.Sleep(2 * time.Second)
			continue
		}

		// 2. Atomically mark as RUNNING using Optimistic/Where Locking
		// This prevents another worker from grabbing the same submission
		updateResult := storage.DB.Model(&submission).
			Where("id = ? AND status = ?", submission.ID, models.StatusPending).
			Update("status", models.StatusRunning)

		if updateResult.RowsAffected == 0 {
			// Another worker grabbed it first, continue polling
			continue
		}

		log.Printf("[Worker %d] Processing Submission: %s for Problem: %d", workerID, submission.ID, submission.ProblemID)

		// 3. Process the submission
		processSubmission(&submission)

		log.Printf("[Worker %d] Finished Submission: %s", workerID, submission.ID)
	}
}

func processSubmission(s *models.Submission) {
	// A. Fetch Problem Details (Limits & Tests)
	problemMeta := storage.ProblemCache[s.ProblemID]
	if problemMeta.ID == 0 {
		storage.UpdateSubmissionStatus(s.ID, models.StatusCE, "System Error: Problem not found", 0, 0, 0)
		return
	}

	hiddenTestsRaw, err := storage.GetProblemFile(s.ProblemID, "tests.txt")
	if err != nil {
		storage.UpdateSubmissionStatus(s.ID, models.StatusCE, "System Error: Tests not found", 0, 0, 0)
		return
	}

	// B. Build Payload
	payload := sandbox.BuildExecutableCode(s.Code, hiddenTestsRaw)

	// C. Execute in Sandbox
	// Note: memory stringency can cause instant 137 exit if too low. We default to at least 32MB
	memLimit := problemMeta.MemoryLimitKb
	if memLimit < 32768 {
		memLimit = 32768
	}

	res, err := sandbox.RunCode(s.ID.String(), payload, problemMeta.TimeLimitMs, memLimit)
	if err != nil {
		storage.UpdateSubmissionStatus(s.ID, models.StatusCE, fmt.Sprintf("Sandbox Error: %v", err), 0, 0, 0)
		return
	}

	// D. Determine Final Status
	status := models.StatusAC
	msg := "Accepted"

	if res.ExitCode != 0 {
		if res.ExitCode == 137 || res.OOMKilled {
			status = models.StatusOLE // or MLE (Memory Limit Exceeded)
			msg = "Memory Limit Exceeded or Killed by System"
		} else if res.ExitCode == 124 || res.ExecuteTimeMs >= problemMeta.TimeLimitMs {
			status = models.StatusTLE
			msg = "Time Limit Exceeded"
		} else if res.FailedAtCase > 0 {
			status = models.StatusWA
			msg = "Wrong Answer"
		} else {
			status = models.StatusRE
			msg = "Runtime Error"
		}
	} else if res.FailedAtCase > 0 { // Fallback check just in case ExitCode was 0 but Doctest failed somehow
		status = models.StatusWA
		msg = "Wrong Answer"
	}

	// E. Save to DB
	storage.UpdateSubmissionStatus(s.ID, status, msg, res.ExecuteTimeMs, res.MemoryPeakKb, res.FailedAtCase)
}
