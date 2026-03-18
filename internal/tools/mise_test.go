package tools

import (
	"context"
	"testing"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
)

func TestSetupMiseToolsNil(t *testing.T) {
	if err := SetupMiseTools(context.Background(), executor.NewMockExecutor(), nil); err != nil {
		t.Fatalf("nil config should not error: %v", err)
	}
}

func TestSetupMiseToolsEmpty(t *testing.T) {
	cfg := &config.ToolsConfig{}
	if err := SetupMiseTools(context.Background(), executor.NewMockExecutor(), cfg); err != nil {
		t.Fatalf("empty config should not error: %v", err)
	}
}

func TestSetupMiseToolsInstall(t *testing.T) {
	mock := executor.NewMockExecutor()
	cfg := &config.ToolsConfig{
		Tools: []config.ToolVersion{
			{Name: "node", Version: "20", GlobalInstall: false},
			{Name: "go", Version: "1.22", GlobalInstall: true},
		},
	}

	if err := SetupMiseTools(context.Background(), mock, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mock.HasCommand("mise", "install", "node@20") {
		t.Error("should install node@20")
	}
	if !mock.HasCommand("mise", "install", "go@1.22") {
		t.Error("should install go@1.22")
	}
	if !mock.HasCommand("mise", "use", "-g", "go@1.22") {
		t.Error("should set go@1.22 as global")
	}
	// node should not have global set
	for _, cmd := range mock.Commands {
		if len(cmd.Cmd) >= 4 && cmd.Cmd[0] == "mise" && cmd.Cmd[1] == "use" && cmd.Cmd[3] == "node@20" {
			t.Error("node should not be set as global")
		}
	}
}

func TestSetupMiseToolsInstallFailure(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.SetResponse("mise install node@20", &executor.CommandResult{
		Success: false, ReturnCode: 1, Stderr: "not found",
	})

	cfg := &config.ToolsConfig{
		Tools: []config.ToolVersion{{Name: "node", Version: "20"}},
	}

	err := SetupMiseTools(context.Background(), mock, cfg)
	if err == nil {
		t.Error("should return error on install failure")
	}
}
