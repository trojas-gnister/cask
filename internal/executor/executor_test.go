package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMockExecutorRecordsCommands(t *testing.T) {
	m := NewMockExecutor()
	ctx := context.Background()

	_, _ = m.Execute(ctx, []string{"echo", "hello"})
	_, _ = m.ExecuteSudo(ctx, []string{"pacman", "-S", "vim"})
	_, _ = m.ExecuteShell(ctx, "echo foo | grep foo")

	if m.CommandCount() != 3 {
		t.Errorf("expected 3 commands, got %d", m.CommandCount())
	}
	if m.Commands[0].Cmd[0] != "echo" {
		t.Error("first command should be echo")
	}
	if !m.Commands[1].Sudo {
		t.Error("second command should be sudo")
	}
	if !m.Commands[2].Shell {
		t.Error("third command should be shell")
	}
}

func TestMockExecutorPreConfiguredResponse(t *testing.T) {
	m := NewMockExecutor()
	m.SetResponse("flatpak list", &CommandResult{
		Success:    true,
		ReturnCode: 0,
		Stdout:     "org.mozilla.Firefox\norg.gimp.GIMP",
	})

	r, err := m.Execute(context.Background(), []string{"flatpak", "list"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Stdout != "org.mozilla.Firefox\norg.gimp.GIMP" {
		t.Errorf("unexpected stdout: %s", r.Stdout)
	}
}

func TestMockExecutorDefaultSuccess(t *testing.T) {
	m := NewMockExecutor()
	r, err := m.Execute(context.Background(), []string{"anything"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.Success {
		t.Error("default response should be success")
	}
}

func TestMockExecutorFileOperations(t *testing.T) {
	m := NewMockExecutor()

	if m.FileExists("/tmp/test.txt") {
		t.Error("file should not exist initially")
	}

	_, err := m.WriteFile("/tmp/test.txt", "hello world", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !m.FileExists("/tmp/test.txt") {
		t.Error("file should exist after write")
	}

	content, err := m.ReadFile("/tmp/test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", content)
	}
}

func TestMockExecutorReadFileNotFound(t *testing.T) {
	m := NewMockExecutor()
	_, err := m.ReadFile("/nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestMockExecutorHasCommand(t *testing.T) {
	m := NewMockExecutor()
	_, _ = m.Execute(context.Background(), []string{"flatpak", "install", "org.mozilla.Firefox"})

	if !m.HasCommand("flatpak", "install") {
		t.Error("should find flatpak install command")
	}
	if m.HasCommand("podman") {
		t.Error("should not find podman command")
	}
}

func TestSystemExecutorDryRun(t *testing.T) {
	e := NewSystemExecutor(true)
	ctx := context.Background()

	r, err := e.Execute(ctx, []string{"rm", "-rf", "/"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.Success {
		t.Error("dry-run should succeed")
	}
	if r.Stdout == "" {
		t.Error("dry-run should produce output")
	}
}

func TestSystemExecutorDryRunSudo(t *testing.T) {
	e := NewSystemExecutor(true)
	r, _ := e.ExecuteSudo(context.Background(), []string{"pacman", "-S", "vim"})
	if !r.Success {
		t.Error("dry-run sudo should succeed")
	}
}

func TestSystemExecutorDryRunShell(t *testing.T) {
	e := NewSystemExecutor(true)
	r, _ := e.ExecuteShell(context.Background(), "echo test | grep test")
	if !r.Success {
		t.Error("dry-run shell should succeed")
	}
}

func TestSystemExecutorRealCommand(t *testing.T) {
	e := NewSystemExecutor(false)
	r, err := e.Execute(context.Background(), []string{"echo", "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.Success {
		t.Errorf("echo should succeed, stderr: %s", r.Stderr)
	}
	if r.Stdout != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", r.Stdout)
	}
}

func TestSystemExecutorCommandNotFound(t *testing.T) {
	e := NewSystemExecutor(false)
	r, err := e.Execute(context.Background(), []string{"nonexistent-command-xyz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Success {
		t.Error("nonexistent command should fail")
	}
}

func TestSystemExecutorEmptyCommand(t *testing.T) {
	e := NewSystemExecutor(false)
	_, err := e.Execute(context.Background(), []string{})
	if err == nil {
		t.Error("expected error for empty command")
	}
}

func TestSystemExecutorFileOps(t *testing.T) {
	e := NewSystemExecutor(false)
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if e.FileExists(path) {
		t.Error("file should not exist yet")
	}

	_, err := e.WriteFile(path, "test content", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !e.FileExists(path) {
		t.Error("file should exist after write")
	}

	content, err := e.ReadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "test content" {
		t.Errorf("expected 'test content', got %q", content)
	}
}

func TestSystemExecutorWriteFileDryRun(t *testing.T) {
	e := NewSystemExecutor(true)
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	r, err := e.WriteFile(path, "content", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.Success {
		t.Error("dry-run write should succeed")
	}

	// File should NOT have been written
	if _, err := os.Stat(path); err == nil {
		t.Error("dry-run should not actually write file")
	}
}

func TestSystemExecutorIsDryRun(t *testing.T) {
	if NewSystemExecutor(false).IsDryRun() {
		t.Error("should not be dry run")
	}
	if !NewSystemExecutor(true).IsDryRun() {
		t.Error("should be dry run")
	}
}
