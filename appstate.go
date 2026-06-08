package main

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/gen2brain/beeep"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type AppState struct {
	mu sync.Mutex

	wailsApp    *application.App
	window      *application.WebviewWindow
	systray     *application.SystemTray
	logger      *slog.Logger
	iconPath    string
	sessionChan chan Session
	answerChan  chan SessionAnswer
	done        chan struct{}
	timeoutChan chan struct{}
	pending     bool
}

func NewAppState(logger *slog.Logger) *AppState {
	return &AppState{
		logger:      logger,
		sessionChan: make(chan Session, 1),
		answerChan:  make(chan SessionAnswer),
		done:        make(chan struct{}),
		timeoutChan: make(chan struct{}, 1),
	}
}

func (s *AppState) SetPending(v bool) {
	s.mu.Lock()
	s.pending = v
	s.mu.Unlock()
}

func (s *AppState) IsPending() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pending
}

func (s *AppState) SubmitAnswers(answer SessionAnswer) {
	s.SetPending(false)
	s.systray.SetIcon(trayIdle)
	s.answerChan <- answer
}

func (s *AppState) ServiceStartup(ctx context.Context, opts application.ServiceOptions) error {
	s.iconPath = writeTempIcon(s.logger)

	go runMCPServer(s, s.logger)

	go s.dispatchSessions()
	go s.watchTimeouts()
	go s.watchDone()

	s.logger.Info("service started")
	return nil
}

func (s *AppState) ServiceShutdown() error {
	if s.iconPath != "" {
		os.Remove(s.iconPath)
	}
	s.logger.Info("service shutting down")
	return nil
}

func writeTempIcon(logger *slog.Logger) string {
	f, err := os.CreateTemp("", "askd-notify-*.png")
	if err != nil {
		logger.Warn("failed to create temp icon for notifications", "error", err)
		return ""
	}
	if _, err := f.Write(trayPending); err != nil {
		f.Close()
		os.Remove(f.Name())
		logger.Warn("failed to write temp icon", "error", err)
		return ""
	}
	f.Close()
	return f.Name()
}

func (s *AppState) dispatchSessions() {
	for session := range s.sessionChan {
		s.SetPending(true)
		s.systray.SetIcon(trayPending)
		if err := beeep.Notify("askd", "New question from agent", s.iconPath); err != nil {
			s.logger.Warn("beeep notify failed", "error", err)
		}
		s.window.EmitEvent("new-session", session)
	}
}

func (s *AppState) watchTimeouts() {
	for range s.timeoutChan {
		s.SetPending(false)
		s.systray.SetIcon(trayIdle)
		s.window.EmitEvent("session-timeout")
	}
}

func (s *AppState) watchDone() {
	<-s.done
	s.wailsApp.Quit()
}
