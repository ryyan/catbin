package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log/slog"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	Port         = ":5000"
	Day          = time.Hour * 24
	Week         = Day * 7
	Month        = Day * 31
	DateFormat   = time.RFC3339
	// MaxPasteSize: 64KB allows 10k chars (40k worst-case bytes) + 33% Base64 + URL overhead
	MaxPasteSize = 64 * 1024
	MaxChars     = 10000
)

var (
	TextDir         = ".text"
	CleanerInterval = 11 * time.Minute
	expirations     = []string{"burn", "hour", "day", "week", "month"}
	textCache       sync.Map // key=ID (string), value=expirationDate (time.Time)
)

type response struct {
	Expiration string `json:"expiration"`
	Text       string `json:"text"`
}

// getText reads a file from disk and parses its expiration and ciphertext
func getText(id string) (string, string, error) {
	if id == "" || id == "." {
		return "", "", errors.New("invalid id")
	}

	data, err := os.ReadFile(filepath.Join(TextDir, id))
	if err != nil {
		return "", "", err
	}

	dataStr := string(data)
	idx := strings.IndexByte(dataStr, '\n')
	if idx == -1 {
		return "", "", errors.New("invalid format")
	}

	expiration := dataStr[:idx]
	text := dataStr[idx+1:]

	res, _ := json.Marshal(&response{
		Expiration: expiration,
		Text:       text,
	})
	return string(res), expiration, nil
}

// saveText creates a new file and records the expiration + ciphertext
func saveText(text, expiration string) (string, string, error) {
	if text == "" {
		return "", "", errors.New("missing field: text")
	}
	if !stringInSlice(expiration, expirations) {
		return "", "", errors.New("invalid field: expiration")
	}

	expirationStr := expiration
	var expirationDate time.Time

	if expiration != "burn" {
		expirationDate = time.Now().UTC()
		switch expiration {
		case "hour":  expirationDate = expirationDate.Add(time.Hour)
		case "day":   expirationDate = expirationDate.Add(Day)
		case "week":  expirationDate = expirationDate.Add(Week)
		case "month": expirationDate = expirationDate.Add(Month)
		}
		expirationStr = expirationDate.Format(DateFormat)
	} else {
		// Burn-on-read pastes expire in 24 hours if not read
		expirationDate = time.Now().UTC().Add(Day)
	}

	id, err := generateId(33)
	if err != nil {
		return "", "", err
	}

	file, err := os.Create(filepath.Join(TextDir, id))
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	if _, err := file.WriteString(expirationStr + "\n" + text); err != nil {
		return "", "", err
	}

	textCache.Store(id, expirationDate)
	return id, expirationStr, nil
}

// initialScan populates the in-memory cache and deletes already-expired files
func initialScan() {
	deleted := 0
	scanned := 0
	now := time.Now().UTC()

	files, _ := os.ReadDir(TextDir)
	for _, f := range files {
		if f.IsDir() { continue }
		scanned++

		p := filepath.Join(TextDir, f.Name())
		file, err := os.Open(p)
		if err != nil { continue }

		scanner := bufio.NewScanner(file)
		if scanner.Scan() {
			expStr := scanner.Text()
			file.Close()

			var expDate time.Time
			if expStr == "burn" {
				info, _ := f.Info()
				expDate = info.ModTime().Add(Day)
			} else {
				expDate, _ = time.Parse(DateFormat, expStr)
			}

			if !expDate.IsZero() && expDate.Before(now) {
				os.Remove(p)
				deleted++
			} else {
				textCache.Store(f.Name(), expDate)
			}
		} else {
			file.Close()
		}
	}
	slog.Info("initial scan complete", "scanned", scanned, "deleted", deleted)
}

// cleaner is a background task that periodically clears expired pastes
func cleaner(ctx context.Context) {
	ticker := time.NewTicker(CleanerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			deleted := 0
			now := time.Now().UTC()
			textCache.Range(func(key, value any) bool {
				id := key.(string)
				expDate := value.(time.Time)
				if !expDate.IsZero() && expDate.Before(now) {
					osRemove(id)
					deleted++
				}
				return true
			})
			if deleted > 0 {
				slog.Info("cleaner finished", "deleted", deleted)
			}
		}
	}
}

// osRemove deletes a file from disk and its key from the in-memory map
func osRemove(id string) {
	os.Remove(filepath.Join(TextDir, id))
	textCache.Delete(id)
}

func stringInSlice(s string, list []string) bool {
	for _, v := range list {
		if v == s { return true }
	}
	return false
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

// generateId creates a cryptographically secure random Base62 ID
func generateId(n int) (string, error) {
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		if err != nil { return "", err }
		b[i] = letterBytes[num.Int64()]
	}
	return string(b), nil
}
