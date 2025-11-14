package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestNewBlake2b(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{"valid size 32", 32, false},
		{"valid size 64", 64, false},
		{"valid size 1", 1, false},
		{"invalid size 0", 0, true},
		{"invalid size 65", 65, true},
		{"invalid size negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := NewBlake2b(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBlake2b() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && h == nil {
				t.Error("NewBlake2b() returned nil hash without error")
			}
			if !tt.wantErr && h.size != tt.size {
				t.Errorf("NewBlake2b() size = %d, want %d", h.size, tt.size)
			}
		})
	}
}

func TestBlake2bWrite(t *testing.T) {
	h, _ := NewBlake2b(32)

	data := []byte("test data")
	n, err := h.Write(data)

	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != len(data) {
		t.Errorf("Write() wrote %d bytes, want %d", n, len(data))
	}
}

func TestBlake2bSumMethod(t *testing.T) {
	tests := []struct {
		name string
		data string
		size int
	}{
		{"empty data", "", 32},
		{"short data", "test", 32},
		{"longer data", "The quick brown fox jumps over the lazy dog", 64},
		{"unicode", "你好世界", 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _ := NewBlake2b(tt.size)
			h.Write([]byte(tt.data))
			sum := h.Sum(nil)

			if len(sum) != tt.size {
				t.Errorf("Sum() length = %d, want %d", len(sum), tt.size)
			}
		})
	}
}

func TestBlake2bDeterministic(t *testing.T) {
	data := []byte("test data for deterministic check")

	h1, _ := NewBlake2b(32)
	h1.Write(data)
	sum1 := h1.Sum(nil)

	h2, _ := NewBlake2b(32)
	h2.Write(data)
	sum2 := h2.Sum(nil)

	if !bytes.Equal(sum1, sum2) {
		t.Error("Blake2b should produce deterministic results")
	}
}

func TestBlake2bDifferentData(t *testing.T) {
	data1 := []byte("test data 1")
	data2 := []byte("test data 2")

	h1, _ := NewBlake2b(32)
	h1.Write(data1)
	sum1 := h1.Sum(nil)

	h2, _ := NewBlake2b(32)
	h2.Write(data2)
	sum2 := h2.Sum(nil)

	if bytes.Equal(sum1, sum2) {
		t.Error("Blake2b should produce different hashes for different data")
	}
}

func TestBlake2bMultipleWrites(t *testing.T) {
	data := []byte("test data")

	// Single write
	h1, _ := NewBlake2b(32)
	h1.Write(data)
	sum1 := h1.Sum(nil)

	// Multiple writes
	h2, _ := NewBlake2b(32)
	h2.Write(data[:4])
	h2.Write(data[4:])
	sum2 := h2.Sum(nil)

	if !bytes.Equal(sum1, sum2) {
		t.Error("Blake2b should produce same hash regardless of write chunking")
	}
}

func TestBlake2bSum512(t *testing.T) {
	data := []byte("test data")
	sum := Blake2bSum512(data)

	if len(sum) != 64 {
		t.Errorf("Blake2bSum512() length = %d, want 64", len(sum))
	}

	// Test determinism
	sum2 := Blake2bSum512(data)
	if !bytes.Equal(sum, sum2) {
		t.Error("Blake2bSum512() should be deterministic")
	}
}

func TestBlake2bSum256(t *testing.T) {
	data := []byte("test data")
	sum := Blake2bSum256(data)

	if len(sum) != 32 {
		t.Errorf("Blake2bSum256() length = %d, want 32", len(sum))
	}

	// Test determinism
	sum2 := Blake2bSum256(data)
	if !bytes.Equal(sum, sum2) {
		t.Error("Blake2bSum256() should be deterministic")
	}
}

func TestBlake2bSum(t *testing.T) {
	data := []byte("test")

	tests := []struct {
		name string
		size int
	}{
		{"size 16", 16},
		{"size 32", 32},
		{"size 48", 48},
		{"size 64", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sum, err := Blake2bSum(data, tt.size)
			if err != nil {
				t.Errorf("Blake2bSum() error = %v", err)
			}
			if len(sum) != tt.size {
				t.Errorf("Blake2bSum() length = %d, want %d", len(sum), tt.size)
			}
		})
	}
}

func TestBlake2bSumInvalidSize(t *testing.T) {
	data := []byte("test")

	_, err := Blake2bSum(data, 0)
	if err == nil {
		t.Error("Blake2bSum() with size 0 should return error")
	}

	_, err = Blake2bSum(data, 65)
	if err == nil {
		t.Error("Blake2bSum() with size 65 should return error")
	}
}

// Test against known test vectors (if available)
func TestBlake2bTestVectors(t *testing.T) {
	// Test empty input
	h, _ := NewBlake2b(64)
	sum := h.Sum(nil)

	// Blake2b-512 of empty string
	// Expected: 786a02f742015903c6c6fd852552d272912f4740e15847618a86e217f71f5419d25e1031afee585313896444934eb04b903a685b1448b755d56f701afe9be2ce
	expected := "786a02f742015903c6c6fd852552d272912f4740e15847618a86e217f71f5419d25e1031afee585313896444934eb04b903a685b1448b755d56f701afe9be2ce"
	actual := hex.EncodeToString(sum)

	if actual != expected {
		t.Errorf("Blake2b-512 empty string:\ngot:  %s\nwant: %s", actual, expected)
	}
}

func TestBlake2bLargeData(t *testing.T) {
	// Test with large data
	data := make([]byte, 1024*1024) // 1 MB
	for i := range data {
		data[i] = byte(i % 256)
	}

	h, _ := NewBlake2b(32)
	h.Write(data)
	sum := h.Sum(nil)

	if len(sum) != 32 {
		t.Errorf("Blake2b with large data: length = %d, want 32", len(sum))
	}
}

func TestBlake2bSumWithPrefix(t *testing.T) {
	h, _ := NewBlake2b(32)
	h.Write([]byte("test"))

	prefix := []byte("prefix")
	sum := h.Sum(prefix)

	// Sum should include prefix
	if len(sum) != len(prefix)+32 {
		t.Errorf("Sum() with prefix length = %d, want %d", len(sum), len(prefix)+32)
	}

	// Prefix should be intact
	if !bytes.Equal(sum[:len(prefix)], prefix) {
		t.Error("Sum() did not preserve prefix")
	}
}
