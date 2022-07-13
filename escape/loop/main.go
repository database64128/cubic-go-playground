package main

import (
	"crypto/rand"
	"fmt"
	"unsafe"
)

func main() {
	for i := 0; i < 10; i++ {
		v := randomValue()
		fmt.Println("Iter", i, "value", v, "memaddr", uintptr(unsafe.Pointer(&v)))
	}
}

func randomValue() (v [16]byte) {
	_, err := rand.Read(v[:])
	if err != nil {
		panic(err)
	}
	return
}
