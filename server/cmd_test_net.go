package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"pt_lpoj/sandbox"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

func main() {
	err := sandbox.InitDockerClient()
	if err != nil {
		panic(err)
	}

	code := `
import urllib.request
import sys
try:
	urllib.request.urlopen("http://example.com", timeout=1)
except Exception as e:
	print("Caught:", type(e).__name__, getattr(e, "reason", str(e)))
	sys.exit(1)
`
	ctx := context.Background()
	resp, err := sandbox.DockerClient.ContainerCreate(ctx, &container.Config{
		Image:           sandbox.PythonImage,
		Cmd:             []string{"python", "-c", code},
		NetworkDisabled: true,
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	err = sandbox.DockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	statusCh, errCh := sandbox.DockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case <-statusCh:
	case <-errCh:
	}

	outReader, err := sandbox.DockerClient.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		panic(err)
	}

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	limitReader := io.LimitReader(outReader, 2*1024*1024)
	stdcopy.StdCopy(stdoutBuf, stderrBuf, limitReader)
	outReader.Close()
	sandbox.DockerClient.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})

	fmt.Printf("StdOut: %q\n", stdoutBuf.String())
	fmt.Printf("StdErr: %q\n", stderrBuf.String())
}
