package scheduler

import (
	"context"
	"fmt"
	"log"
	"pt_lpoj/models"
	"pt_lpoj/sandbox"
	"pt_lpoj/storage"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	// DefaultJobQueueBuffer is the buffer size for the jobs channel
	DefaultJobQueueBuffer = 100
	// subscriptionCleanupInterval is how often we clean up orphaned SSE subscriptions
	subscriptionCleanupInterval = 5 * time.Minute
	// subscriptionTimeout is how long a subscription lives before it's considered orphaned
	subscriptionTimeout = 10 * time.Minute
)

// Result represents the outcome of a submission evaluation.
type Result struct {
	SubmissionID   uuid.UUID
	Status         models.SubmissionStatus
	Message        string
	ExecutionTimeMs int
	MemoryPeakKb   int
	FailedCase     int
}

// subEntry holds an SSE subscription with its creation time for timeout tracking.
type subEntry struct {
	ch      chan Result
	created time.Time
}

// JobQueue manages the job distribution via Go channels.
type JobQueue struct {
	jobs    chan uuid.UUID                 // 待处理的 submission ID
	sseSubs map[uuid.UUID]*subEntry        // SSE 订阅表
	subsMu  sync.RWMutex                   // 保护 sseSubs

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// GlobalQueue is the shared job queue instance used across the server.
// It must be initialized via SetGlobalQueue before use.
var GlobalQueue *JobQueue

// SetGlobalQueue initializes the global job queue. Call once from main.
func SetGlobalQueue(q *JobQueue) {
	GlobalQueue = q
}

// NewJobQueue creates a JobQueue with a buffered channel of the given size.
func NewJobQueue(bufferSize int) *JobQueue {
	if bufferSize <= 0 {
		bufferSize = DefaultJobQueueBuffer
	}
	ctx, cancel := context.WithCancel(context.Background())
	q := &JobQueue{
		jobs:    make(chan uuid.UUID, bufferSize),
		sseSubs: make(map[uuid.UUID]*subEntry),
		ctx:     ctx,
		cancel:  cancel,
	}
	go q.cleanupLoop()
	return q
}

// Enqueue adds a submission ID to the job queue.
func (q *JobQueue) Enqueue(subID uuid.UUID) {
	select {
	case q.jobs <- subID:
	default:
		// Queue full, log and drop (should not happen in practice)
		log.Printf("[JobQueue] WARNING: job channel full, dropping submission %s", subID)
	}
}

// Subscribe registers an SSE client for results of the given submission.
// Returns a channel that will receive the result when ready.
func (q *JobQueue) Subscribe(subID uuid.UUID) chan Result {
	q.subsMu.Lock()
	defer q.subsMu.Unlock()

	// If already subscribed, close old channel and replace
	if entry, exists := q.sseSubs[subID]; exists {
		close(entry.ch)
	}

	ch := make(chan Result, 1)
	q.sseSubs[subID] = &subEntry{
		ch:      ch,
		created: time.Now(),
	}
	return ch
}

// Unsubscribe removes the SSE subscription for the given submission.
func (q *JobQueue) Unsubscribe(subID uuid.UUID) {
	q.subsMu.Lock()
	defer q.subsMu.Unlock()

	if entry, exists := q.sseSubs[subID]; exists {
		close(entry.ch)
		delete(q.sseSubs, subID)
	}
}

// Close shuts down the job queue and all workers.
func (q *JobQueue) Close() {
	q.cancel()
	close(q.jobs) // 阻止新任务进入
	q.wg.Wait()
}

// cleanupLoop periodically removes orphaned SSE subscriptions.
func (q *JobQueue) cleanupLoop() {
	ticker := time.NewTicker(subscriptionCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-q.ctx.Done():
			return
		case <-ticker.C:
			q.removeOrphanedSubs()
		}
	}
}

// removeOrphanedSubs removes subscriptions that have been around too long
// without receiving a result (e.g. client disconnected without unsubscribing).
func (q *JobQueue) removeOrphanedSubs() {
	q.subsMu.Lock()
	defer q.subsMu.Unlock()

	now := time.Now()
	for subID, entry := range q.sseSubs {
		if now.Sub(entry.created) > subscriptionTimeout {
			close(entry.ch)
			delete(q.sseSubs, subID)
			log.Printf("[JobQueue] Cleaned up orphaned SSE subscription for %s", subID)
		}
	}
}

// StartWorkerPool initializes and starts a given number of worker goroutines
// that consume submissions from the provided JobQueue.
func StartWorkerPool(workerCount int, queue *JobQueue) {
	// Crash Recovery: revert any RUNNING submissions to PENDING
	ids := recoverOrphanedSubmissions()
	for _, id := range ids {
		queue.Enqueue(id)
	}

	for i := 0; i < workerCount; i++ {
		queue.wg.Add(1)
		go queue.runWorker(i)
	}
	log.Printf("Started %d Judge Workers with channel-based queue", workerCount)
}

// recoverOrphanedSubmissions reverts any 'RUNNING' submissions back to 'PENDING'
// and returns their IDs so they can be re-enqueued.
func recoverOrphanedSubmissions() []uuid.UUID {
	var subs []models.Submission
	storage.DB.Where("status = ?", models.StatusRunning).Find(&subs)

	if len(subs) == 0 {
		return nil
	}

	storage.DB.Model(&models.Submission{}).
		Where("status = ?", models.StatusRunning).
		Update("status", models.StatusPending)

	ids := make([]uuid.UUID, len(subs))
	for i, s := range subs {
		ids[i] = s.ID
	}

	log.Printf("[Crash Recovery] Reverted %d orphaned submissions back to PENDING", len(subs))
	return ids
}

func (q *JobQueue) runWorker(workerID int) {
	defer q.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Worker %d] PANIC recovered: %v", workerID, r)
			go q.runWorker(workerID)
		}
	}()
	for {
		select {
		case <-q.ctx.Done():
			return
		case subID, ok := <-q.jobs:
			if !ok {
				return
			}
			q.processSubmission(subID, workerID)
		}
	}
}

func (q *JobQueue) processSubmission(subID uuid.UUID, workerID int) {
	var submission models.Submission
	if err := storage.DB.Where("id = ?", subID).First(&submission).Error; err != nil {
		log.Printf("[Worker %d] Submission %s not found in DB: %v", workerID, subID, err)
		return
	}

	result := storage.DB.Model(&submission).
		Where("id = ? AND status = ?", subID, models.StatusPending).
		Update("status", models.StatusRunning)
	if result.RowsAffected == 0 {
		return
	}

	log.Printf("[Worker %d] Processing Submission: %s for Problem: %d", workerID, subID, submission.ProblemID)

	status, msg, timeMs, memKb, failedCase := q.evaluate(&submission)
	storage.UpdateSubmissionStatus(subID, status, msg, timeMs, memKb, failedCase)

	q.pushResult(Result{
		SubmissionID:   subID,
		Status:         status,
		Message:        msg,
		ExecutionTimeMs: timeMs,
		MemoryPeakKb:   memKb,
		FailedCase:     failedCase,
	})

	log.Printf("[Worker %d] Finished Submission: %s -> %s", workerID, subID, status)
}

func (q *JobQueue) evaluate(s *models.Submission) (status models.SubmissionStatus, msg string, timeMs int, memKb int, failedCase int) {
	problemMeta := storage.ProblemCache[s.ProblemID]
	if problemMeta.ID == 0 {
		return models.StatusCE, "System Error: Problem not found", 0, 0, 0
	}

	hiddenTestsRaw, err := storage.GetProblemFile(s.ProblemID, "tests.txt")
	if err != nil {
		return models.StatusCE, "System Error: Tests not found", 0, 0, 0
	}

	payload := sandbox.BuildExecutableCode(s.Code, hiddenTestsRaw)

	memLimit := problemMeta.MemoryLimitKb
	if memLimit < 32768 {
		memLimit = 32768
	}

	res, err := sandbox.RunCode(s.ID.String(), payload, problemMeta.TimeLimitMs, memLimit)
	if err != nil {
		return models.StatusCE, fmt.Sprintf("Sandbox Error: %v", err), 0, 0, 0
	}

	if res.ExitCode != 0 {
		if res.ExitCode == 137 || res.OOMKilled {
			status = models.StatusOLE
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
	} else if res.FailedAtCase > 0 {
		status = models.StatusWA
		msg = "Wrong Answer"
	} else {
		status = models.StatusAC
		msg = "Accepted"
	}

	return status, msg, res.ExecuteTimeMs, res.MemoryPeakKb, res.FailedAtCase
}

func (q *JobQueue) pushResult(r Result) {
	q.subsMu.RLock()
	entry, ok := q.sseSubs[r.SubmissionID]
	q.subsMu.RUnlock()

	if !ok {
		return
	}

	select {
	case entry.ch <- r:
	default:
		log.Printf("[JobQueue] WARNING: result channel for %s was full or blocked", r.SubmissionID)
	}

	go func(id uuid.UUID) {
		time.Sleep(500 * time.Millisecond)
		q.Unsubscribe(id)
	}(r.SubmissionID)
}
