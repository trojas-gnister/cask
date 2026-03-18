// Package executor provides an abstraction for running system commands.
//
// The Executor interface decouples business logic from system calls,
// enabling testing with MockExecutor and production use with SystemExecutor.
package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DefaultTimeout for command execution.
const DefaultTimeout = 120 * time.Second

// CommandResult holds the output and status of an executed command.
type CommandResult struct {
	Success    bool
	ReturnCode int
	Stdout     string
	Stderr     string
	Command    []string
}

// Executor defines the interface for running system commands.
type Executor interface {
	// Execute runs a command without elevated privileges.
	Execute(ctx context.Context, cmd []string) (*CommandResult, error)

	// ExecuteSudo runs a command with sudo.
	ExecuteSudo(ctx context.Context, cmd []string) (*CommandResult, error)

	// ExecuteShell runs a shell command string (for piping, etc.).
	ExecuteShell(ctx context.Context, cmd string) (*CommandResult, error)

	// FileExists checks if a file exists at the given path.
	FileExists(path string) bool

	// ReadFile reads a file's contents.
	ReadFile(path string) (string, error)

	// WriteFile writes content to a file. If sudo is true, uses tee via shell.
	WriteFile(path string, content string, sudo bool) (*CommandResult, error)

	// IsDryRun returns whether the executor is in dry-run mode.
	IsDryRun() bool
}

// SystemExecutor runs real system commands via os/exec.
type SystemExecutor struct {
	DryRun  bool
	Timeout time.Duration
}

// NewSystemExecutor creates a SystemExecutor with the given options.
func NewSystemExecutor(dryRun bool) *SystemExecutor {
	return &SystemExecutor{
		DryRun:  dryRun,
		Timeout: DefaultTimeout,
	}
}

func (e *SystemExecutor) IsDryRun() bool {
	return e.DryRun
}

func (e *SystemExecutor) Execute(ctx context.Context, cmd []string) (*CommandResult, error) {
	if len(cmd) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	if e.DryRun {
		return &CommandResult{
			Success:    true,
			ReturnCode: 0,
			Stdout:     fmt.Sprintf("[dry-run] Would execute: %s", shellJoin(cmd)),
			Command:    cmd,
		}, nil
	}
	return e.run(ctx, cmd)
}

func (e *SystemExecutor) ExecuteSudo(ctx context.Context, cmd []string) (*CommandResult, error) {
	fullCmd := append([]string{"sudo"}, cmd...)
	if e.DryRun {
		return &CommandResult{
			Success:    true,
			ReturnCode: 0,
			Stdout:     fmt.Sprintf("[dry-run] Would execute: %s", shellJoin(fullCmd)),
			Command:    fullCmd,
		}, nil
	}
	return e.run(ctx, fullCmd)
}

func (e *SystemExecutor) ExecuteShell(ctx context.Context, cmd string) (*CommandResult, error) {
	if e.DryRun {
		return &CommandResult{
			Success:    true,
			ReturnCode: 0,
			Stdout:     fmt.Sprintf("[dry-run] Would execute: %s", cmd),
			Command:    []string{"sh", "-c", cmd},
		}, nil
	}
	return e.runShell(ctx, cmd)
}

func (e *SystemExecutor) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (e *SystemExecutor) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (e *SystemExecutor) WriteFile(path string, content string, sudo bool) (*CommandResult, error) {
	if sudo {
		return e.runShell(context.Background(),
			fmt.Sprintf("printf '%%s' %s | sudo tee %s > /dev/null",
				shellQuote(content), shellQuote(path)))
	}
	if e.DryRun {
		return &CommandResult{
			Success:    true,
			ReturnCode: 0,
			Stdout:     fmt.Sprintf("[dry-run] Would write to: %s", path),
			Command:    []string{"write", path},
		}, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return nil, err
	}
	return &CommandResult{
		Success:    true,
		ReturnCode: 0,
		Command:    []string{"write", path},
	}, nil
}

func (e *SystemExecutor) run(ctx context.Context, cmd []string) (*CommandResult, error) {
	timeout := e.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	c := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	var stdout, stderr strings.Builder
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	rc := 0
	if c.ProcessState != nil {
		rc = c.ProcessState.ExitCode()
	}
	if ctx.Err() == context.DeadlineExceeded {
		return &CommandResult{
			Success:    false,
			ReturnCode: -1,
			Stderr:     fmt.Sprintf("Command timed out after %s: %s", timeout, shellJoin(cmd)),
			Command:    cmd,
		}, nil
	}
	if err != nil {
		return &CommandResult{
			Success:    false,
			ReturnCode: rc,
			Stdout:     stdout.String(),
			Stderr:     stderr.String(),
			Command:    cmd,
		}, nil
	}
	return &CommandResult{
		Success:    true,
		ReturnCode: rc,
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		Command:    cmd,
	}, nil
}

func (e *SystemExecutor) runShell(ctx context.Context, cmd string) (*CommandResult, error) {
	return e.run(ctx, []string{"sh", "-c", cmd})
}

// shellJoin joins command parts into a shell-safe string for display.
func shellJoin(parts []string) string {
	quoted := make([]string, len(parts))
	for i, p := range parts {
		if strings.ContainsAny(p, " \t\n\"'\\$") {
			quoted[i] = shellQuote(p)
		} else {
			quoted[i] = p
		}
	}
	return strings.Join(quoted, " ")
}

// shellQuote returns a single-quoted shell string.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
