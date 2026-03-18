package flatpak

import (
	"context"
	"testing"

	"github.com/iskry/cask/internal/executor"
)

func TestInstallEmpty(t *testing.T) {
	mock := executor.NewMockExecutor()
	r, err := Install(context.Background(), mock, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.Success {
		t.Error("empty install should succeed")
	}
	if mock.CommandCount() != 0 {
		t.Error("empty install should not execute commands")
	}
}

func TestInstallPackages(t *testing.T) {
	mock := executor.NewMockExecutor()
	_, err := Install(context.Background(), mock, []string{"org.mozilla.Firefox", "org.gimp.GIMP"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.HasCommand("flatpak", "install") {
		t.Error("should execute flatpak install")
	}
	// Should be sudo
	if !mock.Commands[0].Sudo {
		t.Error("flatpak install should use sudo")
	}
}

func TestRemovePackages(t *testing.T) {
	mock := executor.NewMockExecutor()
	_, err := Remove(context.Background(), mock, []string{"org.mozilla.Firefox"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.HasCommand("flatpak", "uninstall") {
		t.Error("should execute flatpak uninstall")
	}
}

func TestListFlatpak(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.SetResponse("flatpak list --app --columns=application", &executor.CommandResult{
		Success: true,
		Stdout:  "org.mozilla.Firefox\norg.gimp.GIMP\n",
	})

	apps, err := List(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 2 {
		t.Errorf("expected 2 apps, got %d", len(apps))
	}
	if apps[0] != "org.mozilla.Firefox" {
		t.Errorf("expected Firefox, got %s", apps[0])
	}
}

func TestListFlatpakEmpty(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.SetResponse("flatpak list --app --columns=application", &executor.CommandResult{
		Success: true,
		Stdout:  "",
	})

	apps, err := List(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 0 {
		t.Errorf("expected 0 apps, got %d", len(apps))
	}
}

func TestAddRemote(t *testing.T) {
	mock := executor.NewMockExecutor()
	_, err := AddRemote(context.Background(), mock, "flathub", "https://flathub.org/repo/flathub.flatpakrepo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.HasCommand("flatpak", "remote-add") {
		t.Error("should execute flatpak remote-add")
	}
}
