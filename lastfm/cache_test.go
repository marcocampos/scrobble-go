package lastfm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMemoryCache_GetSet(t *testing.T) {
	c := NewMemoryCache()

	if _, ok := c.Get("missing"); ok {
		t.Error("Get on empty cache should return false")
	}

	c.Set("key1", "value1")
	v, ok := c.Get("key1")
	if !ok {
		t.Error("Get after Set returned false")
	}
	if v != "value1" {
		t.Errorf("Get = %q, want %q", v, "value1")
	}
}

func TestMemoryCache_Overwrite(t *testing.T) {
	c := NewMemoryCache()
	c.Set("k", "v1")
	c.Set("k", "v2")
	v, _ := c.Get("k")
	if v != "v2" {
		t.Errorf("Get = %q, want %q", v, "v2")
	}
}

func TestBoltCache_GetSet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	c, err := NewBoltCache(path)
	if err != nil {
		t.Fatalf("NewBoltCache: %v", err)
	}
	defer c.Close()

	if _, ok := c.Get("missing"); ok {
		t.Error("Get on empty cache should return false")
	}

	c.Set("key1", "value1")
	v, ok := c.Get("key1")
	if !ok {
		t.Error("Get after Set returned false")
	}
	if v != "value1" {
		t.Errorf("Get = %q, want %q", v, "value1")
	}
}

func TestBoltCache_Persistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "persist.db")

	// Write in one instance.
	c1, err := NewBoltCache(path)
	if err != nil {
		t.Fatalf("NewBoltCache: %v", err)
	}
	c1.Set("persisted", "hello")
	c1.Close()

	// Read back in a new instance.
	c2, err := NewBoltCache(path)
	if err != nil {
		t.Fatalf("NewBoltCache (reopen): %v", err)
	}
	defer c2.Close()

	v, ok := c2.Get("persisted")
	if !ok {
		t.Fatal("expected persisted key to be found after reopen")
	}
	if v != "hello" {
		t.Errorf("persisted value = %q, want %q", v, "hello")
	}
}

func TestBoltCache_InvalidPath(t *testing.T) {
	_, err := NewBoltCache("/nonexistent/dir/test.db")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

func TestBoltCache_Close(t *testing.T) {
	path := filepath.Join(t.TempDir(), "close.db")
	c, err := NewBoltCache(path)
	if err != nil {
		t.Fatalf("NewBoltCache: %v", err)
	}
	if err := c.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
	// The file should still exist after close.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("database file should still exist after Close")
	}
}
