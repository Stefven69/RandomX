// Command randomx-example demonstrates basic usage of the RandomX Go bindings.
// It computes a RandomX hash for a fixed key and input, mirroring the behavior
// of the C api-example1.c included with the RandomX library.
package main

import (
	"fmt"
	"os"

	"github.com/Stefven69/RandomX/go/randomx"
)

func main() {
	key := []byte("RandomX example key")
	input := []byte("RandomX example input")

	flags := randomx.GetFlags()

	cache := randomx.AllocCache(flags)
	if cache == nil {
		fmt.Fprintln(os.Stderr, "Cache allocation failed")
		os.Exit(1)
	}
	defer cache.Release()

	cache.Init(key)

	vm := randomx.CreateVM(flags, cache, nil)
	if vm == nil {
		fmt.Fprintln(os.Stderr, "Failed to create virtual machine")
		os.Exit(1)
	}
	defer vm.Destroy()

	hash := vm.CalcHash(input)
	fmt.Printf("%x\n", hash)
}
