// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/moby/moby/client"
)

func dockerAvailable() bool {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = cli.Ping(ctx, client.PingOptions{})
	return err == nil
}

func TestCreateAndRemove(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("docker not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	env, err := Create(ctx, "gobook-test-unit", "alpine:latest", false)
	if err != nil {
		t.Fatalf("Create() = _, %v", err)
	}
	if env.ID == "" {
		t.Fatal("env.ID is empty")
	}
	if env.Name != "gobook-test-unit" {
		t.Errorf("env.Name = %q, want %q", env.Name, "gobook-test-unit")
	}

	if err := env.Remove(ctx); err != nil {
		t.Fatalf("Remove() = %v", err)
	}
}

func TestCreateDinD(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("docker not available")
	}
	if os.Getenv("GOBOOK_TEST_DIND") != "1" {
		t.Skip("DinD test requires GOBOOK_TEST_DIND=1")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	env, err := Create(ctx, "gobook-test-dind", "docker:dind", true)
	if err != nil {
		t.Fatalf("Create(dind) = _, %v", err)
	}
	defer env.Remove(ctx)

	if env.ID == "" {
		t.Fatal("env.ID is empty")
	}
}

func TestRunNoArgs(t *testing.T) {
	code := run(nil)
	if code != 1 {
		t.Errorf("run(nil) = %d, want 1", code)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	code := run([]string{"bad"})
	if code != 1 {
		t.Errorf("run(bad) = %d, want 1", code)
	}
}

func TestRunRemoveNoArgs(t *testing.T) {
	code := run([]string{"remove"})
	if code != 1 {
		t.Errorf("run(remove) = %d, want 1", code)
	}
}

func TestRunCreateAndRemove(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("docker not available")
	}

	code := run([]string{"create", "gobook-test-cli", "alpine:latest"})
	if code != 0 {
		t.Fatalf("run(create) = %d, want 0", code)
	}

	code = run([]string{"remove", "gobook-test-cli"})
	if code != 0 {
		t.Fatalf("run(remove) = %d, want 0", code)
	}
}
