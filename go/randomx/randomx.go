// Package randomx provides Go bindings for the RandomX proof-of-work library.
//
// RandomX is a proof-of-work (PoW) algorithm optimized for general-purpose CPUs.
// It uses random code execution together with memory-hard techniques to minimize
// the efficiency advantage of specialized hardware.
//
// This package wraps the C API defined in randomx.h via CGo. Before using this
// package, the RandomX C library (librandomx.a) must be compiled and available
// at the path referenced by the CGo LDFLAGS directive.
//
// Basic usage example:
//
//	flags := randomx.GetFlags()
//	cache := randomx.AllocCache(flags)
//	defer cache.Release()
//	cache.Init([]byte("example key"))
//	vm := randomx.CreateVM(flags, cache, nil)
//	defer vm.Destroy()
//	hash := vm.CalcHash([]byte("example input"))
//	fmt.Printf("%x\n", hash)
package randomx

/*
#cgo CXXFLAGS: -std=c++11
#cgo LDFLAGS: -L${SRCDIR}/../../build -lrandomx -lstdc++ -lm -lpthread
#include "../../src/randomx.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

// HashSize is the size of a RandomX hash output in bytes.
const HashSize = 32

// DatasetItemSize is the size of a single RandomX dataset item in bytes.
const DatasetItemSize = 64

// Flags controls the behavior and resource usage of RandomX components.
type Flags int

const (
	// FlagDefault uses portable software implementations; works on all platforms.
	FlagDefault Flags = C.RANDOMX_FLAG_DEFAULT
	// FlagLargePages allocates memory in large pages for improved performance.
	FlagLargePages Flags = C.RANDOMX_FLAG_LARGE_PAGES
	// FlagHardAES uses hardware-accelerated AES instructions.
	FlagHardAES Flags = C.RANDOMX_FLAG_HARD_AES
	// FlagFullMem uses the full 2 GiB dataset for fast-mode mining.
	FlagFullMem Flags = C.RANDOMX_FLAG_FULL_MEM
	// FlagJIT enables the JIT compiler for faster hash computation.
	FlagJIT Flags = C.RANDOMX_FLAG_JIT
	// FlagSecure enforces W^X policy when combined with FlagJIT.
	FlagSecure Flags = C.RANDOMX_FLAG_SECURE
	// FlagArgon2SSSE3 uses SSSE3-optimized Argon2 for cache initialization.
	FlagArgon2SSSE3 Flags = C.RANDOMX_FLAG_ARGON2_SSSE3
	// FlagArgon2AVX2 uses AVX2-optimized Argon2 for cache initialization.
	FlagArgon2AVX2 Flags = C.RANDOMX_FLAG_ARGON2_AVX2
	// FlagArgon2 enables all available Argon2 optimizations.
	FlagArgon2 Flags = C.RANDOMX_FLAG_ARGON2
)

// GetFlags returns the recommended flags for the current machine.
// The returned flags do not include FlagLargePages, FlagFullMem, or FlagSecure;
// add those manually if needed.
func GetFlags() Flags {
	return Flags(C.randomx_get_flags())
}

// Cache wraps the randomx_cache C structure. Use AllocCache to create one.
type Cache struct {
	ptr *C.randomx_cache
}

// AllocCache creates and allocates memory for a RandomX cache using the given flags.
// Returns nil if allocation fails or if the requested flags are unsupported.
// The caller must call Release when done with the cache.
func AllocCache(flags Flags) *Cache {
	ptr := C.randomx_alloc_cache(C.randomx_flags(flags))
	if ptr == nil {
		return nil
	}
	return &Cache{ptr: ptr}
}

// Init initializes (or reinitializes) the cache with the given key.
// The key must not be empty.
func (c *Cache) Init(key []byte) {
	if len(key) == 0 {
		return
	}
	C.randomx_init_cache(c.ptr, unsafe.Pointer(&key[0]), C.size_t(len(key)))
}

// Memory returns the raw internal memory buffer of the cache.
func (c *Cache) Memory() unsafe.Pointer {
	return C.randomx_get_cache_memory(c.ptr)
}

// Release frees all memory associated with the cache.
// After calling Release the Cache must not be used again.
func (c *Cache) Release() {
	if c.ptr != nil {
		C.randomx_release_cache(c.ptr)
		c.ptr = nil
	}
}

// Dataset wraps the randomx_dataset C structure. Use AllocDataset to create one.
type Dataset struct {
	ptr *C.randomx_dataset
}

// AllocDataset creates and allocates memory for a RandomX dataset using the given flags.
// Returns nil if allocation fails.
// The caller must call Release when done with the dataset.
func AllocDataset(flags Flags) *Dataset {
	ptr := C.randomx_alloc_dataset(C.randomx_flags(flags))
	if ptr == nil {
		return nil
	}
	return &Dataset{ptr: ptr}
}

// ItemCount returns the total number of items in the dataset.
// All items from 0 to ItemCount()-1 must be initialized before use.
func DatasetItemCount() uint64 {
	return uint64(C.randomx_dataset_item_count())
}

// Init initializes a contiguous range of dataset items starting at startItem.
// cache must be a previously initialized Cache.
func (d *Dataset) Init(cache *Cache, startItem, itemCount uint64) {
	C.randomx_init_dataset(d.ptr, cache.ptr, C.ulong(startItem), C.ulong(itemCount))
}

// Memory returns the raw internal memory buffer of the dataset.
func (d *Dataset) Memory() unsafe.Pointer {
	return C.randomx_get_dataset_memory(d.ptr)
}

// Release frees all memory associated with the dataset.
// After calling Release the Dataset must not be used again.
func (d *Dataset) Release() {
	if d.ptr != nil {
		C.randomx_release_dataset(d.ptr)
		d.ptr = nil
	}
}

// VM wraps the randomx_vm C structure. Use CreateVM to create one.
type VM struct {
	ptr *C.randomx_vm
}

// CreateVM creates and initializes a RandomX virtual machine.
//
// cache may be nil when flags includes FlagFullMem.
// dataset may be nil when flags does not include FlagFullMem.
// Returns nil if creation fails (e.g. unsupported flags, memory allocation failure,
// or missing required cache/dataset).
// The caller must call Destroy when done.
func CreateVM(flags Flags, cache *Cache, dataset *Dataset) *VM {
	var cachePtr *C.randomx_cache
	var datasetPtr *C.randomx_dataset
	if cache != nil {
		cachePtr = cache.ptr
	}
	if dataset != nil {
		datasetPtr = dataset.ptr
	}
	ptr := C.randomx_create_vm(C.randomx_flags(flags), cachePtr, datasetPtr)
	if ptr == nil {
		return nil
	}
	return &VM{ptr: ptr}
}

// SetCache reinitializes the VM with a new cache.
// Must be called whenever the cache is reinitialized with a new key.
func (vm *VM) SetCache(cache *Cache) {
	C.randomx_vm_set_cache(vm.ptr, cache.ptr)
}

// SetDataset reinitializes the VM with a new dataset.
func (vm *VM) SetDataset(dataset *Dataset) {
	C.randomx_vm_set_dataset(vm.ptr, dataset.ptr)
}

// Destroy releases all memory associated with the VM.
// After calling Destroy the VM must not be used again.
func (vm *VM) Destroy() {
	if vm.ptr != nil {
		C.randomx_destroy_vm(vm.ptr)
		vm.ptr = nil
	}
}

// CalcHash calculates a 32-byte RandomX hash of the given input.
func (vm *VM) CalcHash(input []byte) []byte {
	output := make([]byte, HashSize)
	if len(input) == 0 {
		return output
	}
	C.randomx_calculate_hash(vm.ptr, unsafe.Pointer(&input[0]), C.size_t(len(input)), unsafe.Pointer(&output[0]))
	return output
}

// CalcHashFirst begins a pipelined hash calculation for input.
// Call CalcHashNext or CalcHashLast to retrieve the result.
//
// WARNING: These pipelined functions may alter the floating point rounding mode
// of the calling goroutine's OS thread.
func (vm *VM) CalcHashFirst(input []byte) {
	if len(input) == 0 {
		return
	}
	C.randomx_calculate_hash_first(vm.ptr, unsafe.Pointer(&input[0]), C.size_t(len(input)))
}

// CalcHashNext outputs the hash of the previous input and begins computing
// the hash of nextInput. Must be preceded by CalcHashFirst.
func (vm *VM) CalcHashNext(nextInput []byte) []byte {
	output := make([]byte, HashSize)
	if len(nextInput) == 0 {
		return output
	}
	C.randomx_calculate_hash_next(vm.ptr, unsafe.Pointer(&nextInput[0]), C.size_t(len(nextInput)), unsafe.Pointer(&output[0]))
	return output
}

// CalcHashLast outputs the hash of the previous input.
// Must be preceded by CalcHashFirst or CalcHashNext.
func (vm *VM) CalcHashLast() []byte {
	output := make([]byte, HashSize)
	C.randomx_calculate_hash_last(vm.ptr, unsafe.Pointer(&output[0]))
	return output
}

// CalcCommitment calculates a RandomX commitment from a hash and its original input.
// hash must be a HashSize-byte slice produced by CalcHash (or equivalent).
// Returns a new HashSize-byte commitment value.
func CalcCommitment(input []byte, hash []byte) []byte {
	output := make([]byte, HashSize)
	if len(input) == 0 || len(hash) < HashSize {
		return output
	}
	C.randomx_calculate_commitment(
		unsafe.Pointer(&input[0]), C.size_t(len(input)),
		unsafe.Pointer(&hash[0]),
		unsafe.Pointer(&output[0]),
	)
	return output
}
