package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "catbin_test_*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	oldDir := TextDir
	TextDir = filepath.Join(tmpDir, ".text")
	os.MkdirAll(TextDir, 0755)
	t.Cleanup(func() { TextDir = oldDir })

	// Reset cache
	textCache.Range(func(k, v any) bool {
		textCache.Delete(k)
		return true
	})
}

func TestStore(t *testing.T) {
	setupTest(t)

	t.Run("SaveAndGet", func(t *testing.T) {
		id, _, err := saveText("blob", "hour")
		if err != nil {
			t.Fatal(err)
		}
		res, exp, _ := getText(id)
		var r response
		json.Unmarshal([]byte(res), &r)
		if r.Text != "blob" || exp == "burn" {
			t.Error("getText mismatch")
		}
	})

	t.Run("BurnOnRead", func(t *testing.T) {
		id, _, _ := saveText("burn", "burn")
		_, exp, _ := getText(id)
		if exp != "burn" {
			t.Error("should be burn")
		}
		osRemove(id)
		_, _, err := getText(id)
		if err == nil {
			t.Error("should be deleted")
		}
	})

	t.Run("EdgeCases", func(t *testing.T) {
		if _, _, err := saveText("", "hour"); err == nil { t.Error("empty text") }
		if _, _, err := saveText("a", "invalid"); err == nil { t.Error("invalid exp") }
		if _, _, err := getText(""); err == nil { t.Error("empty id") }
		if _, _, err := getText("."); err == nil { t.Error("invalid id .") }
		
		// Bad file format
		os.WriteFile(filepath.Join(TextDir, "bad"), []byte("no-newline"), 0644)
		if _, _, err := getText("bad"); err == nil { t.Error("bad format") }
	})
}

func TestHandlers(t *testing.T) {
	setupTest(t)
	router := setupRouter(".")

	t.Run("Lifecycle", func(t *testing.T) {
		form := url.Values{"text": {"hello"}, "expiration": {"hour"}}
		req := httptest.NewRequest("POST", "/msg", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		
		var postRes map[string]string
		json.Unmarshal(rr.Body.Bytes(), &postRes)
		id := postRes["id"]

		req = httptest.NewRequest("GET", "/msg/"+id, nil)
		req.SetPathValue("id", id) 
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != 200 { t.Errorf("GET failed %d", rr.Code) }
	})

	t.Run("Errors", func(t *testing.T) {
		// Large text
		form := url.Values{"text": {strings.Repeat("a", MaxChars+1)}, "expiration": {"hour"}}
		req := httptest.NewRequest("POST", "/msg", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusRequestEntityTooLarge { t.Error("expected 413") }

		// Not found
		req = httptest.NewRequest("GET", "/msg/missing", nil)
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != 404 { t.Error("expected 404") }
	})
}

func TestSPARouting(t *testing.T) {
	setupTest(t)
	webDir, _ := os.MkdirTemp("", "web")
	defer os.RemoveAll(webDir)
	os.WriteFile(filepath.Join(webDir, "index.html"), []byte("index"), 0644)
	router := setupRouter(webDir)

	req := httptest.NewRequest("GET", "/any-path", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Body.String() != "index" { t.Error("SPA fallback failed") }
}

func TestInitialScan(t *testing.T) {
	setupTest(t)
	
	// Valid
	os.WriteFile(filepath.Join(TextDir, "v"), []byte(time.Now().Add(time.Hour).Format(DateFormat)+"\nok"), 0644)
	// Expired
	os.WriteFile(filepath.Join(TextDir, "e"), []byte(time.Now().Add(-time.Hour).Format(DateFormat)+"\nold"), 0644)
	// Burn
	os.WriteFile(filepath.Join(TextDir, "b"), []byte("burn\nnow"), 0644)
	// Garbage
	os.WriteFile(filepath.Join(TextDir, "g"), []byte("garbage"), 0644)

	initialScan()
	if _, ok := textCache.Load("v"); !ok { t.Error("v missing") }
	if _, ok := textCache.Load("e"); ok { t.Error("e should be gone") }
	if _, ok := textCache.Load("b"); !ok { t.Error("b missing") }
}

func TestCleaner(t *testing.T) {
	setupTest(t)
	oldInterval := CleanerInterval
	CleanerInterval = 10 * time.Millisecond
	t.Cleanup(func() { CleanerInterval = oldInterval })

	id, _, _ := saveText("expired", "hour")
	// Manually backdate the cache
	textCache.Store(id, time.Now().Add(-time.Hour))

	ctx, cancel := context.WithCancel(context.Background())
	go cleaner(ctx)
	
	time.Sleep(50 * time.Millisecond)
	cancel()

	if _, ok := textCache.Load(id); ok {
		t.Error("cleaner failed to delete expired item")
	}
}

func TestUtilities(t *testing.T) {
	if !stringInSlice("a", []string{"a"}) { t.Error("fail") }
	id, _ := generateId(33)
	if len(id) != 33 { t.Error("id len") }
}
