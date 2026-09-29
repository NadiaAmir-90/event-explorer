package models

import "testing"

func TestCacheClear(t *testing.T) {
	cache := NewCache()

	fetchCount := 0

	fetch := func() interface{} {
		fetchCount++
		return "data"
	}

	// First request: cache miss.
	result1 := cache.GetOrFetch("key", fetch)

	if result1 != "data" {
		t.Fatalf("expected data, got %v", result1)
	}

	if fetchCount != 1 {
		t.Fatalf("expected fetch to be called once, got %d", fetchCount)
	}

	// Second request: cache hit.
	result2 := cache.GetOrFetch("key", fetch)

	if result2 != "data" {
		t.Fatalf("expected data, got %v", result2)
	}

	if fetchCount != 1 {
		t.Fatalf("expected fetch to still be called once, got %d", fetchCount)
	}

	// Explicitly clear the cache.
	cache.Clear()

	// Third request: cache miss because cache was cleared.
	result3 := cache.GetOrFetch("key", fetch)

	if result3 != "data" {
		t.Fatalf("expected data, got %v", result3)
	}

	if fetchCount != 2 {
		t.Fatalf("expected fetch to be called twice after Clear(), got %d", fetchCount)
	}

	// Fourth request: cache hit again.
	result4 := cache.GetOrFetch("key", fetch)

	if result4 != "data" {
		t.Fatalf("expected data, got %v", result4)
	}

	if fetchCount != 2 {
		t.Fatalf("expected fetch to still be called twice, got %d", fetchCount)
	}
}