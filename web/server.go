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

	logBroadcaster := newLogBroadcaster()
	log.SetOutput(io.MultiWriter(os.Stderr, logBroadcaster))
	log.SetFlags(0)

	mux := http.NewServeMux()
	srv := &Server{
		cfg:       cfg,
		sessions:  newSessionStore(cfg.parsedSessionTTL()),
		templates: tmplCache,
		logs:      logBroadcaster,
		stats:     newRunStats(),
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go srv.sessions.cleanupLoop(ctx)

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
