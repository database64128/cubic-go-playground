package csprng

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"testing"

	"golang.org/x/crypto/chacha20"
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

	h := blake3.New(size, key)
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

func initAesCtr(b *testing.B, keySize int) cipher.Stream {
	key := make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		b.Fatal(err)
	}

	cb, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	iv := make([]byte, cb.BlockSize())
	_, err = rand.Read(iv)
	if err != nil {
		b.Fatal(err)
	}

	return cipher.NewCTR(cb, iv)
}

func BenchmarkAes128CtrSmall(b *testing.B) {
	buf := make([]byte, testSmallSize)
	cs := initAesCtr(b, 16)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs.XORKeyStream(buf, buf)
	}
}

func BenchmarkAes128CtrBig(b *testing.B) {
	buf := make([]byte, testBigSize)
	cs := initAesCtr(b, 16)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs.XORKeyStream(buf, buf)
	}
}

func BenchmarkAes256CtrSmall(b *testing.B) {
	buf := make([]byte, testSmallSize)
	cs := initAesCtr(b, 32)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs.XORKeyStream(buf, buf)
	}
}

func BenchmarkAes256CtrBig(b *testing.B) {
	buf := make([]byte, testBigSize)
	cs := initAesCtr(b, 32)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs.XORKeyStream(buf, buf)
	}
}

func initChaCha20(b *testing.B) cipher.Stream {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		b.Fatal(err)
	}

	nonce := make([]byte, chacha20.NonceSize)
	_, err = rand.Read(nonce)
	if err != nil {
		b.Fatal(err)
	}

	cs, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		b.Fatal(err)
	}
	return cs
}

func BenchmarkChaCha20Small(b *testing.B) {
	buf := make([]byte, testSmallSize)
	cs := initChaCha20(b)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs.XORKeyStream(buf, buf)
	}
}

func BenchmarkChaCha20Big(b *testing.B) {
	buf := make([]byte, testBigSize)
	cs := initChaCha20(b)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs.XORKeyStream(buf, buf)
	}
}
