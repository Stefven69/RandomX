# RandomX Go Bindings

This directory contains Go bindings for the [RandomX](https://github.com/tevador/RandomX) proof-of-work library. The bindings use [CGo](https://pkg.go.dev/cmd/cgo) to wrap the C API exposed by `src/randomx.h`.

## Prerequisites

The C library (`librandomx.a`) must be compiled before building or testing the Go code. From the repository root:

```bash
mkdir -p build && cd build
cmake ..
make -j$(nproc) randomx
```

## Usage

```go
import "github.com/Stefven69/RandomX/go/randomx"

func main() {
    flags := randomx.GetFlags()

    cache := randomx.AllocCache(flags)
    defer cache.Release()
    cache.Init([]byte("my key"))

    vm := randomx.CreateVM(flags, cache, nil)
    defer vm.Destroy()

    hash := vm.CalcHash([]byte("my input"))
    fmt.Printf("%x\n", hash)
}
```

## Running the example

```bash
# From the repository root, after building the C library:
cd go
go run ./cmd/randomx-example/
```

## Running tests

```bash
# From the repository root, after building the C library:
cd go
go test ./...
```

## Package structure

| Path | Description |
|------|-------------|
| `randomx/` | Core CGo package wrapping the C API |
| `cmd/randomx-example/` | CLI equivalent of `src/tests/api-example1.c` |

## API

The Go API mirrors the C API from `src/randomx.h`:

| Go | C equivalent |
|----|-------------|
| `randomx.GetFlags()` | `randomx_get_flags()` |
| `randomx.AllocCache(flags)` | `randomx_alloc_cache()` |
| `cache.Init(key)` | `randomx_init_cache()` |
| `cache.Release()` | `randomx_release_cache()` |
| `randomx.AllocDataset(flags)` | `randomx_alloc_dataset()` |
| `randomx.DatasetItemCount()` | `randomx_dataset_item_count()` |
| `dataset.Init(cache, start, count)` | `randomx_init_dataset()` |
| `dataset.Release()` | `randomx_release_dataset()` |
| `randomx.CreateVM(flags, cache, dataset)` | `randomx_create_vm()` |
| `vm.SetCache(cache)` | `randomx_vm_set_cache()` |
| `vm.SetDataset(dataset)` | `randomx_vm_set_dataset()` |
| `vm.Destroy()` | `randomx_destroy_vm()` |
| `vm.CalcHash(input)` | `randomx_calculate_hash()` |
| `vm.CalcHashFirst(input)` | `randomx_calculate_hash_first()` |
| `vm.CalcHashNext(nextInput)` | `randomx_calculate_hash_next()` |
| `vm.CalcHashLast()` | `randomx_calculate_hash_last()` |
| `randomx.CalcCommitment(input, hash)` | `randomx_calculate_commitment()` |
