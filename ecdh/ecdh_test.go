package ecdh

import (
	"crypto/ecdh"
	"crypto/rand"
	"testing"
	"unsafe"
)

type publicKeyHeader struct {
	curve     ecdh.Curve
	publicKey []byte
	boring    unsafe.Pointer
}

func BenchmarkX25519(b *testing.B) {
	key, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	pubkey := key.PublicKey()
	pubkeyHeader := (*publicKeyHeader)(unsafe.Pointer(&pubkey))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pubkeyHeader.publicKey, err = key.ECDH(pubkey)
		if err != nil {
			b.Fatal(err)
		}
	}
}
