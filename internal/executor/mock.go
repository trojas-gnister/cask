package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// RecordedCommand is a command captured by MockExecutor.
type RecordedCommand struct {
	Cmd   []string
	Sudo  bool
	Shell bool
}

// MockExecutor records commands without executing them.
// It can return pre-configured responses for specific command patterns.
type MockExecutor struct {
	mu        sync.Mutex
	Commands  []RecordedCommand
	Files     map[string]string
	responses map[string]*CommandResult
}

// NewMockExecutor creates a MockExecutor ready for use.
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		Files:     make(map[string]string),
		responses: make(map[string]*CommandResult),
	}
}

// SetResponse pre-configures a response for a command key (space-joined args).
func (m *MockExecutor) SetResponse(cmdKey string, result *CommandResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[cmdKey] = result
}

func (m *MockExecutor) getResponse(cmd []string) *CommandResult {
	key := strings.Join(cmd, " ")
	if r, ok := m.responses[key]; ok {
		return r
	}
	return &CommandResult{Success: true, ReturnCode: 0, Stdout: "", Command: cmd}
}

func (m *MockExecutor) IsDryRun() bool {
	return false
}

func (m *MockExecutor) Execute(_ context.Context, cmd []string) (*CommandResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Commands = append(m.Commands, RecordedCommand{Cmd: cmd, Sudo: false})
	return m.getResponse(cmd), nil
}

func (m *MockExecutor) ExecuteSudo(_ context.Context, cmd []string) (*CommandResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Commands = append(m.Commands, RecordedCommand{Cmd: cmd, Sudo: true})
	return m.getResponse(cmd), nil
}

func (m *MockExecutor) ExecuteShell(_ context.Context, cmd string) (*CommandResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	parts := []string{"sh", "-c", cmd}
	m.Commands = append(m.Commands, RecordedCommand{Cmd: parts, Shell: true})
	return m.getResponse(parts), nil
}

func (m *MockExecutor) FileExists(path string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.Files[path]
	return ok
}

func (m *MockExecutor) ReadFile(path string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	content, ok := m.Files[path]
	if !ok {
		return "", fmt.Errorf("mock file not found: %s", path)
	}
	return content, nil
}

func (m *MockExecutor) WriteFile(path string, content string, sudo bool) (*CommandResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Files[path] = content
	cmd := []string{"write", path}
	m.Commands = append(m.Commands, RecordedCommand{Cmd: cmd, Sudo: sudo})
	return &CommandResult{Success: true, ReturnCode: 0, Command: cmd}, nil
}

// CommandCount returns the number of recorded commands.
func (m *MockExecutor) CommandCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Commands)
}

// HasCommand checks if a command matching the given prefix was recorded.
func (m *MockExecutor) HasCommand(prefix ...string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, rec := range m.Commands {
		if len(rec.Cmd) >= len(prefix) {
			match := true
			for i, p := range prefix {
				if rec.Cmd[i] != p {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}
