// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// testenv creates and tears down Docker test environments.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// defaultImage is the base image for test environments.
const defaultImage = "golang:1.27"

// defaultTimeout is the maximum time to wait for container operations.
const defaultTimeout = 2 * time.Minute

// Env represents a running test environment.
type Env struct {
	ID   string
	Name string
	cli  *client.Client
}

// Create creates a new test environment container. The container is started
// and ready when Create returns. The caller must call Remove to clean up.
func Create(ctx context.Context, name string, img string, dind bool) (*Env, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}

	pullResp, err := cli.ImagePull(ctx, img, client.ImagePullOptions{})
	if err != nil {
		return nil, fmt.Errorf("pull %s: %w", img, err)
	}
	io.Copy(io.Discard, pullResp)
	pullResp.Close()

	config := &container.Config{
		Image: img,
		Cmd:   []string{"sleep", "infinity"},
		Labels: map[string]string{
			"gobook-testenv": "true",
		},
	}

	hostConfig := &container.HostConfig{}
	if dind {
		hostConfig.Privileged = true
	}

	createResult, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     config,
		HostConfig: hostConfig,
		Name:       name,
	})
	if err != nil {
		return nil, fmt.Errorf("create container: %w", err)
	}

	if _, err := cli.ContainerStart(ctx, createResult.ID, client.ContainerStartOptions{}); err != nil {
		cli.ContainerRemove(ctx, createResult.ID, client.ContainerRemoveOptions{Force: true})
		return nil, fmt.Errorf("start container: %w", err)
	}

	return &Env{
		ID:   createResult.ID,
		Name: name,
		cli:  cli,
	}, nil
}

// Remove stops and removes the test environment container.
func (e *Env) Remove(ctx context.Context) error {
	timeout := 10
	_, stopErr := e.cli.ContainerStop(ctx, e.ID, client.ContainerStopOptions{Timeout: &timeout})
	if stopErr != nil {
		e.cli.ContainerRemove(ctx, e.ID, client.ContainerRemoveOptions{Force: true})
		return fmt.Errorf("stop container %s: %w", e.Name, stopErr)
	}
	if _, err := e.cli.ContainerRemove(ctx, e.ID, client.ContainerRemoveOptions{}); err != nil {
		return fmt.Errorf("remove container %s: %w", e.Name, err)
	}
	return nil
}

func run(args []string) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "usage: testenv <create|remove> [options]\n")
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	switch args[0] {
	case "create":
		name := "gobook-test"
		img := defaultImage
		dind := false
		if len(args) > 1 {
			name = args[1]
		}
		if len(args) > 2 {
			img = args[2]
		}
		if len(args) > 3 && args[3] == "--dind" {
			dind = true
		}
		env, err := Create(ctx, name, img, dind)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		fmt.Printf("created: %s (%s)\n", env.Name, env.ID[:12])
		return 0

	case "remove":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "usage: testenv remove <name-or-id>\n")
			return 1
		}
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: docker client: %v\n", err)
			return 1
		}
		env := &Env{ID: args[1], Name: args[1], cli: cli}
		if err := env.Remove(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		fmt.Printf("removed: %s\n", args[1])
		return 0

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		return 1
	}
}

func main() {
	os.Exit(run(os.Args[1:]))
}
