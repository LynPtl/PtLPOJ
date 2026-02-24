package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
)

// RunResult holds the output from the Docker sandbox
type RunResult struct {
	ExitCode      int
	Stdout        string
	Stderr        string
	ExecuteTimeMs int
	FailedAtCase  int
	OOMKilled     bool
}

// RunCode spins up a highly restricted Docker container to execute the provided payload.
func RunCode(submissionID string, payload string, timeLimitMs int, memoryLimitKB int) (*RunResult, error) {
	// 1. Create a temporary directory on the host to hold the payload
	hostTempDir := filepath.Join(os.TempDir(), "ptlpoj_sandbox", submissionID)
	err := os.MkdirAll(hostTempDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox dir: %w", err)
	}
	defer os.RemoveAll(hostTempDir) // Cleanup after execution

	entryPointPath := filepath.Join(hostTempDir, "main.py")
	err = os.WriteFile(entryPointPath, []byte(payload), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write payload to disk: %w", err)
	}

	// 2. Prepare container configurations
	// Memory limits in bytes
	memBytes := int64(memoryLimitKB) * 1024

	ctx := context.Background()

	resp, err := DockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image:           PythonImage,
			Cmd:             []string{"python", "/sandbox/main.py"},
			User:            "nobody", // Deep Security: Prevent root access
			Tty:             false,
			NetworkDisabled: true, // Cut off internet access
		},
		&container.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:/sandbox:ro", hostTempDir), // Read-only mount
			},
			Resources: container.Resources{
				Memory:     memBytes,
				MemorySwap: memBytes,                                      // Prevent swapping to disk
				PidsLimit:  func() *int64 { p := int64(20); return &p }(), // Anti-forkbomb
			},
			ReadonlyRootfs: true,
			CapDrop:        []string{"ALL"}, // Principle of least privilege
			NetworkMode:    "none",
			AutoRemove:     false, // We need to inspect it before removal
		},
		&network.NetworkingConfig{},
		nil,
		"ptlpoj_judge_"+submissionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	containerID := resp.ID

	// 3. Start Container & Execute
	startTime := time.Now()
	if err := DockerClient.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		_ = DockerClient.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// 4. Wait for completion or Timeout
	// We add a realistic buffer (+500ms) over the problem's time limit to account for Docker startup overhead
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeLimitMs+500)*time.Millisecond)
	defer cancel()

	statusCh, errCh := DockerClient.ContainerWait(timeoutCtx, containerID, container.WaitConditionNotRunning)

	var exitCode int
	var timeoutTriggered bool

	select {
	case err := <-errCh:
		if err != nil && timeoutCtx.Err() == context.DeadlineExceeded {
			// TLE triggered by our context wrap
			timeoutTriggered = true
			DockerClient.ContainerKill(ctx, containerID, "SIGKILL")
		} else if err != nil {
			return nil, fmt.Errorf("wait error: %w", err)
		}
	case status := <-statusCh:
		exitCode = int(status.StatusCode)
	case <-timeoutCtx.Done(): // Context deadline hit
		timeoutTriggered = true
		DockerClient.ContainerKill(ctx, containerID, "SIGKILL")
	}

	execTime := int(time.Since(startTime).Milliseconds())

	// 5. Inspect for constraints and OOM
	inspect, err := DockerClient.ContainerInspect(ctx, containerID)
	oomKilled := false
	if err == nil {
		oomKilled = inspect.State.OOMKilled
	}

	// 6. Fetch Logs (Output)
	outReader, err := DockerClient.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	if err == nil {
		// Demultiplex docker streams using official tool
		limitReader := io.LimitReader(outReader, 2*1024*1024)
		_, _ = stdcopy.StdCopy(stdoutBuf, stderrBuf, limitReader)
		outReader.Close()
	}

	// Clean up container
	_ = DockerClient.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})

	// 7. Parse output logic
	cleanOut := stdoutBuf.String()
	if stderrBuf.Len() > 0 {
		cleanOut += "\n" + stderrBuf.String()
	}
	failedCase := 0

	// If it timed out, force an exit code that signifies TLE
	if timeoutTriggered || execTime >= timeLimitMs {
		exitCode = 124 // Standard timeout exit code equivalent
	} else if exitCode != 0 {
		// Attempt to extract line failure
		failedCase = extractFailedCaseNumberFromDoctest(cleanOut)
		if failedCase == 0 {
			failedCase = 1
		} // generic fail map
	}

	return &RunResult{
		ExitCode:      exitCode,
		Stdout:        cleanOut,
		ExecuteTimeMs: execTime,
		FailedAtCase:  failedCase,
		OOMKilled:     oomKilled,
	}, nil
}

// extractFailedCaseNumberFromDoctest uses Regex to parse out "Failed example:" lines inside verbose output
func extractFailedCaseNumberFromDoctest(output string) int {
	// Standard doctest verbose starts each example with "Trying:"
	reTrying := regexp.MustCompile(`(?m)^Trying:`)
	matchesTry := reTrying.FindAllStringIndex(output, -1)

	reFailed := regexp.MustCompile(`(?m)^Failed example:(\s*)(.*)`)
	matchesFail := reFailed.FindStringIndex(output)

	if matchesFail != nil && len(matchesTry) > 0 {
		failPos := matchesFail[0]
		// Count how many "Trying:" occurred before the first "Failed example"
		count := 0
		for _, idx := range matchesTry {
			if idx[0] < failPos {
				count++
			} else {
				break
			}
		}
		return count
	}
	return 0
}
