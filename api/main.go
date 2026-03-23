package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := os.MkdirAll(TextDir, 0755); err != nil {
		slog.Error("failed to create text directory", "error", err)
		os.Exit(1)
	}

	initialScan()

	server := &http.Server{
		Addr:         Port,
		Handler:      setupRouter("../web"),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go cleaner(ctx)

	go func() {
		slog.Info("server started", "port", Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
	slog.Info("server stopped")
}

func setupRouter(webDir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /msg/{id}", handleGetMsg)
	mux.HandleFunc("POST /msg", handlePostMsg)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(webDir, r.URL.Path)
		if stat, err := os.Stat(path); err != nil || stat.IsDir() {
			http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
			return
		}
		http.FileServer(http.Dir(webDir)).ServeHTTP(w, r)
	})

	return middleware(mux)
}
