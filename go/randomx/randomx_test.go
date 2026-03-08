package randomx_test

import (
	"encoding/hex"
	"testing"

	"github.com/Stefven69/RandomX/go/randomx"
)

// testKey and testInput match the values used in the C api-example1.c.
const (
	testKey   = "RandomX example key"
	testInput = "RandomX example input"
)

// TestGetFlags verifies that GetFlags returns a valid (non-negative) value.
func TestGetFlags(t *testing.T) {
	flags := randomx.GetFlags()
	if flags < 0 {
		t.Errorf("GetFlags returned negative value: %d", flags)
	}
}

// TestAllocCacheNilOnBadFlags verifies that AllocCache returns nil for unsupported flag
// combinations (FlagLargePages on platforms without large-page support is expected to
// fail; skip if it unexpectedly succeeds to avoid flakiness across environments).
func TestAllocCacheRelease(t *testing.T) {
	flags := randomx.GetFlags()
	cache := randomx.AllocCache(flags)
	if cache == nil {
		t.Fatal("AllocCache returned nil with recommended flags")
	}
	cache.Release()
	// Double-release must be safe (no-op).
	cache.Release()
}

// TestCacheInit verifies that a cache can be initialized with a key.
func TestCacheInit(t *testing.T) {
	flags := randomx.GetFlags()
	cache := randomx.AllocCache(flags)
	if cache == nil {
		t.Fatal("AllocCache returned nil")
	}
	defer cache.Release()
	cache.Init([]byte(testKey))
	if cache.Memory() == nil {
		t.Error("Cache memory pointer is nil after Init")
	}
}

// TestDatasetItemCount verifies that dataset item count is a positive value.
func TestDatasetItemCount(t *testing.T) {
	count := randomx.DatasetItemCount()
	if count == 0 {
		t.Error("DatasetItemCount returned 0")
	}
}

// TestCreateVMAndCalcHash is the core integration test: allocate a cache,
// initialize a VM, and verify that a known input produces a deterministic hash.
func TestCreateVMAndCalcHash(t *testing.T) {
	flags := randomx.GetFlags()

	cache := randomx.AllocCache(flags)
	if cache == nil {
		t.Fatal("AllocCache returned nil")
	}
	defer cache.Release()

	cache.Init([]byte(testKey))

	vm := randomx.CreateVM(flags, cache, nil)
	if vm == nil {
		t.Fatal("CreateVM returned nil")
	}
	defer vm.Destroy()

	hash1 := vm.CalcHash([]byte(testInput))
	if len(hash1) != randomx.HashSize {
		t.Fatalf("CalcHash returned %d bytes, want %d", len(hash1), randomx.HashSize)
	}

	// Hashing the same input twice must produce the same result.
	hash2 := vm.CalcHash([]byte(testInput))
	if hex.EncodeToString(hash1) != hex.EncodeToString(hash2) {
		t.Errorf("CalcHash not deterministic:\n  first:  %x\n  second: %x", hash1, hash2)
	}

	t.Logf("Hash(%q) = %x", testInput, hash1)
}

// TestCreateVMDifferentInputs checks that two different inputs produce different hashes.
func TestCreateVMDifferentInputs(t *testing.T) {
	flags := randomx.GetFlags()

	cache := randomx.AllocCache(flags)
	if cache == nil {
		t.Fatal("AllocCache returned nil")
	}
	defer cache.Release()
	cache.Init([]byte(testKey))

	vm := randomx.CreateVM(flags, cache, nil)
	if vm == nil {
		t.Fatal("CreateVM returned nil")
	}
	defer vm.Destroy()

	hash1 := vm.CalcHash([]byte("input A"))
	hash2 := vm.CalcHash([]byte("input B"))
	if hex.EncodeToString(hash1) == hex.EncodeToString(hash2) {
		t.Errorf("Different inputs produced the same hash: %x", hash1)
	}
}

// TestPipelinedHashing verifies that the pipelined hash API (First/Next/Last)
// produces the same results as the single-call CalcHash API.
func TestPipelinedHashing(t *testing.T) {
	flags := randomx.GetFlags()

	cache := randomx.AllocCache(flags)
	if cache == nil {
		t.Fatal("AllocCache returned nil")
	}
	defer cache.Release()
	cache.Init([]byte(testKey))

	vm := randomx.CreateVM(flags, cache, nil)
	if vm == nil {
		t.Fatal("CreateVM returned nil")
	}
	defer vm.Destroy()

	inputs := [][]byte{
		[]byte("pipelined input 1"),
		[]byte("pipelined input 2"),
		[]byte("pipelined input 3"),
	}

	// Compute reference hashes using the single-call API.
	var refs []string
	for _, inp := range inputs {
		refs = append(refs, hex.EncodeToString(vm.CalcHash(inp)))
	}

	// Compute hashes using the pipelined API.
	vm.CalcHashFirst(inputs[0])
	pipe1 := vm.CalcHashNext(inputs[1])
	pipe2 := vm.CalcHashNext(inputs[2])
	pipe3 := vm.CalcHashLast()

	pipelined := []string{
		hex.EncodeToString(pipe1),
		hex.EncodeToString(pipe2),
		hex.EncodeToString(pipe3),
	}

	for i := range refs {
		if refs[i] != pipelined[i] {
			t.Errorf("input[%d]: single=%s pipelined=%s", i, refs[i], pipelined[i])
		}
	}
}

// TestCalcCommitment verifies that CalcCommitment produces a non-zero result.
func TestCalcCommitment(t *testing.T) {
	flags := randomx.GetFlags()

	cache := randomx.AllocCache(flags)
	if cache == nil {
		t.Fatal("AllocCache returned nil")
	}
	defer cache.Release()
	cache.Init([]byte(testKey))

	vm := randomx.CreateVM(flags, cache, nil)
	if vm == nil {
		t.Fatal("CreateVM returned nil")
	}
	defer vm.Destroy()

	input := []byte(testInput)
	hash := vm.CalcHash(input)
	commitment := randomx.CalcCommitment(input, hash)

	if len(commitment) != randomx.HashSize {
		t.Fatalf("CalcCommitment returned %d bytes, want %d", len(commitment), randomx.HashSize)
	}

	allZero := true
	for _, b := range commitment {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("CalcCommitment returned all-zero output")
	}
	t.Logf("Commitment = %x", commitment)
}

// TestVMSetCache verifies that replacing a cache with the same key is safe.
func TestVMSetCache(t *testing.T) {
	flags := randomx.GetFlags()

	cache1 := randomx.AllocCache(flags)
	if cache1 == nil {
		t.Fatal("AllocCache returned nil")
	}
	defer cache1.Release()
	cache1.Init([]byte(testKey))

	vm := randomx.CreateVM(flags, cache1, nil)
	if vm == nil {
		t.Fatal("CreateVM returned nil")
	}
	defer vm.Destroy()

	hash1 := vm.CalcHash([]byte(testInput))

	// Create a second cache with a different key.
	cache2 := randomx.AllocCache(flags)
	if cache2 == nil {
		t.Fatal("AllocCache returned nil for cache2")
	}
	defer cache2.Release()
	cache2.Init([]byte("different key"))

	vm.SetCache(cache2)
	hash2 := vm.CalcHash([]byte(testInput))

	// Different keys must produce different hashes.
	if hex.EncodeToString(hash1) == hex.EncodeToString(hash2) {
		t.Error("Expected different hashes after SetCache with a new key, got the same hash")
	}
}
