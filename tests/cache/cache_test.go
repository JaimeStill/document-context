package cache_test

import (
	"testing"

	"github.com/JaimeStill/document-context/pkg/cache"
)

func TestGenerateKey_Deterministic(t *testing.T) {
	input := "/path/to/document.pdf/1.png?dpi=300&format=png&quality=90"

	key1 := cache.GenerateKey(input)
	key2 := cache.GenerateKey(input)

	if key1 != key2 {
		t.Error("expected same input to generate same key")
	}
}

func TestGenerateKey_DifferentInputs(t *testing.T) {
	input1 := "/path/to/document.pdf/1.png?dpi=300&format=png&quality=90"
	input2 := "/path/to/document.pdf/2.png?dpi=300&format=png&quality=90"

	key1 := cache.GenerateKey(input1)
	key2 := cache.GenerateKey(input2)

	if key1 == key2 {
		t.Error("expected different inputs to generate different keys")
	}
}

func TestGenerateKey_Format(t *testing.T) {
	input := "/path/to/document.pdf/1.png?dpi=300&format=png&quality=90"

	key := cache.GenerateKey(input)

	// SHA256 hex encoding produces 64-character string
	if len(key) != 64 {
		t.Errorf("expected 64-character key, got %d characters", len(key))
	}

	// Should be all lowercase hexadecimal
	for _, c := range key {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("expected hex character, got %q", c)
		}
	}
}

func TestGenerateKey_EmptyInput(t *testing.T) {
	key := cache.GenerateKey("")

	// Should still generate valid key for empty input
	if len(key) != 64 {
		t.Errorf("expected 64-character key for empty input, got %d characters", len(key))
	}
}

func TestGenerateKey_SensitiveToOrder(t *testing.T) {
	input1 := "/path/to/document.pdf/1.png?dpi=300&quality=90"
	input2 := "/path/to/document.pdf/1.png?quality=90&dpi=300"

	key1 := cache.GenerateKey(input1)
	key2 := cache.GenerateKey(input2)

	// Different order should produce different keys
	// (this is expected - caller must normalize parameter order)
	if key1 == key2 {
		t.Error("expected different parameter order to generate different keys")
	}
}

func TestCacheEntry_Structure(t *testing.T) {
	entry := &cache.CacheEntry{
		Key:      "test-key",
		Data:     []byte{0x89, 'P', 'N', 'G'},
		Filename: "document.1.png",
	}

	if entry.Key != "test-key" {
		t.Errorf("expected key %q, got %q", "test-key", entry.Key)
	}

	if len(entry.Data) != 4 {
		t.Errorf("expected data length 4, got %d", len(entry.Data))
	}

	if entry.Filename != "document.1.png" {
		t.Errorf("expected filename %q, got %q", "document.1.png", entry.Filename)
	}
}

func TestErrCacheEntryNotFound(t *testing.T) {
	err := cache.ErrCacheEntryNotFound

	if err == nil {
		t.Fatal("expected non-nil error")
	}

	expected := "cache entry not found"
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}
