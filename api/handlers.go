package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
)

// Security middleware: sets restrictive headers and limits request body size
func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self'; img-src 'self' data:; frame-ancestors 'none';")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")

		// 64KB limit protects against large-file-based DoS
		r.Body = http.MaxBytesReader(w, r.Body, MaxPasteSize)
		next.ServeHTTP(w, r)
	})
}

// GET /msg/{id} handles retrieving a paste and "burning" it if needed
func handleGetMsg(w http.ResponseWriter, r *http.Request) {
	// filepath.Base ensures no directory traversal can occur
	id := filepath.Base(r.PathValue("id"))
	result, expirationStr, err := getText(id)
	if err != nil {
		slog.Warn("get message failed", "id", id, "error", err)
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "Text not found")
		return
	}

	// Burn-on-read logic deletes file immediately after read
	if expirationStr == "burn" {
		osRemove(id)
		slog.Info("burned message on read", "id", id)
	}

	io.WriteString(w, result)
}

// POST /msg handles saving a new encrypted paste
func handlePostMsg(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	text := r.FormValue("text")
	expiration := r.FormValue("expiration")

	// Validate character length
	if len(text) > MaxChars {
		http.Error(w, "Paste content too large", http.StatusRequestEntityTooLarge)
		return
	}

	id, expirationStr, err := saveText(text, expiration)
	if err != nil {
		slog.Error("save message failed", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.Info("saved message", "id", id, "expiration", expiration)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":         id,
		"expiration": expirationStr,
	})
}
