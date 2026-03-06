package main

import (
	"errors"
	"strings"
	"testing"
)

func TestRunDefaultsToTUI(t *testing.T) {
	var stdout, stderr strings.Builder
	called := ""
	debugRenderer := ""

	exitCode := run(nil, &stdout, &stderr,
		func(debug string) error {
			called = "cli"
			debugRenderer = debug
			return nil
		},
		func(debug string) error {
			called = "tui"
			debugRenderer = debug
			return nil
		},
	)

	if exitCode != 0 {
		t.Fatalf("expected zero exit code, got %d", exitCode)
	}
	if called != "tui" {
		t.Fatalf("expected tui runner, got %q", called)
	}
	if debugRenderer != "" {
		t.Fatalf("expected empty debug renderer, got %q", debugRenderer)
	}
}

func TestRunCLIFlagOverridesMode(t *testing.T) {
	var stdout, stderr strings.Builder
	called := ""

	exitCode := run([]string{"--mode=tui", "--cli", "--debug-renderer=engine"}, &stdout, &stderr,
		func(debug string) error {
			called = "cli:" + debug
			return nil
		},
		func(debug string) error {
			called = "tui:" + debug
			return nil
		},
	)

	if exitCode != 0 {
		t.Fatalf("expected zero exit code, got %d", exitCode)
	}
	if called != "cli:engine" {
		t.Fatalf("expected cli runner with debug renderer, got %q", called)
	}
}

func TestRunModeCLIUsesCLIRunner(t *testing.T) {
	var stdout, stderr strings.Builder
	called := ""

	exitCode := run([]string{"--mode=cli"}, &stdout, &stderr,
		func(debug string) error {
			called = "cli"
			return nil
		},
		func(debug string) error {
			called = "tui"
			return nil
		},
	)

	if exitCode != 0 {
		t.Fatalf("expected zero exit code, got %d", exitCode)
	}
	if called != "cli" {
		t.Fatalf("expected cli runner, got %q", called)
	}
}

func TestRunRejectsInvalidMode(t *testing.T) {
	var stdout, stderr strings.Builder

	exitCode := run([]string{"--mode=bad"}, &stdout, &stderr,
		func(string) error { return nil },
		func(string) error { return nil },
	)

	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), `invalid mode "bad"; expected tui or cli`) {
		t.Fatalf("expected invalid mode error, got %q", stderr.String())
	}
}

func TestRunRejectsInvalidDebugRenderer(t *testing.T) {
	var stdout, stderr strings.Builder

	exitCode := run([]string{"--debug-renderer=bad"}, &stdout, &stderr,
		func(string) error { return nil },
		func(string) error { return nil },
	)

	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), `invalid debug renderer "bad"; expected view or engine`) {
		t.Fatalf("expected invalid debug renderer error, got %q", stderr.String())
	}
}

func TestRunReturnsOneWhenRunnerFails(t *testing.T) {
	var stdout, stderr strings.Builder

	exitCode := run([]string{"--cli"}, &stdout, &stderr,
		func(string) error { return errors.New("boom") },
		func(string) error { return nil },
	)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "Error starting SwapChess: boom") {
		t.Fatalf("expected launch failure in stderr, got %q", stderr.String())
	}
}

func TestRunHelpWritesUsage(t *testing.T) {
	var stdout, stderr strings.Builder

	exitCode := run([]string{"--help"}, &stdout, &stderr,
		func(string) error { return nil },
		func(string) error { return nil },
	)

	if exitCode != 0 {
		t.Fatalf("expected zero exit code for help, got %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "Usage: swapchess [--cli] [--mode=tui|cli]") {
		t.Fatalf("expected usage in stdout, got %q", stdout.String())
	}
}
