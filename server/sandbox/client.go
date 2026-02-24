package sandbox

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var DockerClient *client.Client

const PythonImage = "python:3.10-alpine"

// InitDockerClient connects to the local Docker Daemon and ensures the necessary image exists.
func InitDockerClient() error {
	var err error
	// client.FromEnv reads environment variables (e.g., DOCKER_HOST) or connects to the default socket
	// We force API version >= 1.44 because newer WSL daemons drop support for older API versions
	DockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.44"))
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Ping the daemon to test the connection immediately
	ping, err := DockerClient.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("could not ping Docker daemon: %w. Are you sure Docker Desktop/Daemon is running and WSL integration is on?", err)
	}
	log.Printf("Docker Daemon Connected! API Version: %s", ping.APIVersion)

	return ensurePythonImage()
}

// ensurePythonImage checks if python:3.10-alpine exists locally, and pulls it if it doesn't.
func ensurePythonImage() error {
	ctx := context.Background()
	_, _, err := DockerClient.ImageInspectWithRaw(ctx, PythonImage)
	if err == nil {
		log.Printf("Image %s already exists locally.", PythonImage)
		return nil
	}

	log.Printf("Image %s not found locally. Pulling from Docker Hub...", PythonImage)
	reader, err := DockerClient.ImagePull(ctx, PythonImage, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", PythonImage, err)
	}
	defer reader.Close()

	// Print pull progress to stdout
	io.Copy(os.Stdout, reader)
	log.Printf("Successfully pulled image %s.", PythonImage)
	return nil
}
