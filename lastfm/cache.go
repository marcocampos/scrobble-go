package lastfm

import (
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

// CacheBackend is the interface for request-response caching.
// Implementations must be safe for concurrent use.
type CacheBackend interface {
	// Get retrieves a cached response by key.
	// Returns the value and true if found, or ("", false) if not.
	Get(key string) (string, bool)
	// Set stores a response string under the given key.
	Set(key, value string)
}

// ── MemoryCache ───────────────────────────────────────────────────────────────

type memoryCacheEntry struct {
	value     string
	expiresAt time.Time // zero value means no expiry
}

// MemoryCache is a thread-safe in-memory CacheBackend.
// Entries are lost when the process exits; use BoltCache for persistence.
// When created with NewMemoryCacheWithTTL, entries expire after the given
// duration and are lazily evicted on the next Get.
type MemoryCache struct {
	mu    sync.RWMutex
	store map[string]memoryCacheEntry
	ttl   time.Duration // 0 means no expiry
}

// NewMemoryCache returns an initialised MemoryCache with no TTL.
// Entries are retained indefinitely. For long-running applications,
// prefer NewMemoryCacheWithTTL to bound memory usage.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{store: make(map[string]memoryCacheEntry)}
}

// NewMemoryCacheWithTTL returns a MemoryCache where entries expire after
// the given duration. Expired entries are lazily evicted on the next Get.
// A TTL of 0 or less disables expiry (equivalent to NewMemoryCache).
func NewMemoryCacheWithTTL(ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		store: make(map[string]memoryCacheEntry),
		ttl:   ttl,
	}
}

// Get implements CacheBackend.
func (c *MemoryCache) Get(key string) (string, bool) {
	c.mu.RLock()
	e, ok := c.store[key]
	c.mu.RUnlock()

	if !ok {
		return "", false
	}
	if !e.expiresAt.IsZero() && time.Now().After(e.expiresAt) {
		c.mu.Lock()
		delete(c.store, key)
		c.mu.Unlock()
		return "", false
	}
	return e.value, true
}

// Set implements CacheBackend.
func (c *MemoryCache) Set(key, value string) {
	e := memoryCacheEntry{value: value}
	if c.ttl > 0 {
		e.expiresAt = time.Now().Add(c.ttl)
	}
	c.mu.Lock()
	c.store[key] = e
	c.mu.Unlock()
}

// ── BoltCache ─────────────────────────────────────────────────────────────────

var boltBucket = []byte("lastfm")

// BoltCache is a persistent CacheBackend backed by a bbolt database file.
// It survives process restarts, making it useful for long-running applications
// or CLI tools that make repeated read-only API calls.
//
// Call Close when you are done to release the file lock.
type BoltCache struct {
	db *bolt.DB
}

// NewBoltCache opens (or creates) a bbolt database at path and returns a
// BoltCache. Returns an error if the file cannot be opened.
func NewBoltCache(path string) (*BoltCache, error) {
	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, err
	}
	// Ensure the bucket exists.
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(boltBucket)
		return err
	}); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &BoltCache{db: db}, nil
}

// Get retrieves a cached value. Returns ("", false) if not found or on error.
func (c *BoltCache) Get(key string) (string, bool) {
	var value string
	var found bool
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(boltBucket)
		if b == nil {
			return nil
		}
		v := b.Get([]byte(key))
		if v != nil {
			value = string(v)
			found = true
		}
		return nil
	})
	if err != nil {
		return "", false
	}
	return value, found
}

// Set stores a value in the cache. Silently ignores write errors.
func (c *BoltCache) Set(key, value string) {
	_ = c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(boltBucket)
		if b == nil {
			return nil
		}
		return b.Put([]byte(key), []byte(value))
	})
}

// Close releases the database file lock. Always call this when done.
func (c *BoltCache) Close() error {
	return c.db.Close()
}
