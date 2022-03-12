package csprng

import (
	"crypto/rand"
	"io"
	"testing"

	"lukechampine.com/blake3"
)

const (
	testSmallSize = 24
	testBigSize   = 1024
)

func BenchmarkCryptoRandomSmall(b *testing.B) {
	buf := make([]byte, testSmallSize)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := rand.Read(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCryptoRandomBig(b *testing.B) {
	buf := make([]byte, testBigSize)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := rand.Read(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func initBlake3KeyedHash(b *testing.B, size int) io.Reader {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		b.Fatal(err)
	}

	h := blake3.New(32, key)
	return h.XOF()
}

func BenchmarkBlake3KeyedHashSmall(b *testing.B) {
	buf := make([]byte, testSmallSize)
	r := initBlake3KeyedHash(b, testSmallSize)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := r.Read(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBlake3KeyedHashBig(b *testing.B) {
	buf := make([]byte, testBigSize)
	r := initBlake3KeyedHash(b, testBigSize)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := r.Read(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}
