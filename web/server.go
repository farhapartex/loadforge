package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	cfg       *WebConfig
	sessions  *SessionStore
	templates *templateCache
	logs      *LogBroadcaster
	stats     *RunStats
	logFile   *os.File
	mux       *http.ServeMux
	http      *http.Server
}

func newServer(configPath string) (*Server, error) {
	cfg, err := loadWebConfig(configPath)
	if err != nil {
		return nil, err
	}

	tmplCache, err := newTemplateCache()
	if err != nil {
		return nil, fmt.Errorf("build template cache: %w", err)
	}

	logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file %q: %w", cfg.LogFile, err)
	}
	log.SetOutput(io.MultiWriter(os.Stderr, logFile))
	log.SetFlags(log.Ltime)

	mux := http.NewServeMux()
	srv := &Server{
		cfg:       cfg,
		sessions:  newSessionStore(cfg.parsedSessionTTL()),
		templates: tmplCache,
		logs:      newLogBroadcaster(),
		stats:     newRunStats(),
		logFile:   logFile,
		mux:       mux,
	}
	srv.registerRoutes()

	srv.http = &http.Server{
		Addr:        cfg.Addr,
		Handler:     mux,
		ReadTimeout: 15 * time.Second,
		IdleTimeout: 60 * time.Second,
		// WriteTimeout intentionally omitted: SSE connections are long-lived
	}

	return srv, nil
}

func Start(configPath string) error {
	srv, err := newServer(configPath)
	if err != nil {
		return fmt.Errorf("initialize server: %w", err)
	}
	defer srv.logFile.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go srv.sessions.cleanupLoop(ctx)
	go srv.logs.tailFile(ctx, srv.cfg.LogFile)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := srv.http.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown error: %v", err)
		}
	}()

	log.Printf("LoadForge web server listening on http://localhost%s", srv.cfg.Addr)
	if err := srv.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
