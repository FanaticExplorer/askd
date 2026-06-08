package main

import (
	"embed"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed frontend/dist
var assets embed.FS

//go:embed assets/tray-idle.png
var trayIdle []byte

//go:embed assets/tray-pending.png
var trayPending []byte

func main() {
	logger, closeLog := initLogging()
	defer closeLog()

	unlock, err := acquireLock()
	if err != nil {
		logger.Warn("failed to acquire lock", "error", err)
		os.Exit(1)
	}
	defer unlock()

	state := NewAppState(logger)

	app := application.New(application.Options{
		Name:        "askd",
		Description: "MCP server with GUI for asking user questions",
		Icon:        trayPending,
		Logger:      logger,
		Services: []application.Service{
			application.NewService(state),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		Windows: application.WindowsOptions{
			DisableQuitOnLastWindowClosed: true,
		},
	})

	state.wailsApp = app

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "askd",
		Width:  550,
		Height: 600,
		Hidden: true,
		URL:    "/",
	})

	state.window = window

	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		e.Cancel()
		window.Hide()
	})

	systray := app.SystemTray.New()
	systray.SetLabel("askd")
	systray.SetIcon(trayIdle)

	systray.OnClick(func() {
		if window.IsVisible() {
			window.Hide()
		} else {
			window.Center()
			window.Show()
			window.Focus()
		}
	})
	systray.OnRightClick(func() {
		systray.OpenMenu()
	})

	state.systray = systray

	menu := app.NewMenu()
	menu.Add("Open").OnClick(func(ctx *application.Context) {
		window.Center()
		window.Show()
		window.Focus()
	})
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
	systray.SetMenu(menu)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		app.Quit()
	}()

	if err := app.Run(); err != nil {
		logger.Error("application error", "error", err)
		os.Exit(1)
	}
}

func acquireLock() (func(), error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	lockDir := filepath.Join(home, ".askd")
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return nil, fmt.Errorf("create lock dir: %w", err)
	}
	lockPath := filepath.Join(lockDir, "lock")

	data, readErr := os.ReadFile(lockPath)
	if readErr == nil {
		var pid int
		fmt.Sscanf(string(data), "%d", &pid)
		if pid > 0 && isProcessAlive(pid) {
			killProcess(pid)
		}
		os.Remove(lockPath)
	}

	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("acquire lock: %w", err)
	}
	fmt.Fprintf(f, "%d", os.Getpid())
	f.Close()
	return func() { os.Remove(lockPath) }, nil
}
