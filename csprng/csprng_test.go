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

	for b.Loop() {
		rand.Read(buf)
	}
}

func BenchmarkCryptoRandomBig(b *testing.B) {
	buf := make([]byte, testBigSize)

	for b.Loop() {
		rand.Read(buf)
	}
}

func initBlake3KeyedHash(size int) io.Reader {
	key := make([]byte, 32)
	rand.Read(key)

	h := blake3.New(size, key)
	return h.XOF()
}

func BenchmarkBlake3KeyedHashSmall(b *testing.B) {
	buf := make([]byte, testSmallSize)
	r := initBlake3KeyedHash(testSmallSize)

	for b.Loop() {
		_, err := r.Read(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBlake3KeyedHashBig(b *testing.B) {
	buf := make([]byte, testBigSize)
	r := initBlake3KeyedHash(testBigSize)

	for b.Loop() {
		_, err := r.Read(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func initAesCtr(b *testing.B, keySize int) cipher.Stream {
	b.Helper()

	key := make([]byte, keySize)
	rand.Read(key)

	cb, err := aes.NewCipher(key)
	if err != nil {
		b.Fatal(err)
	}

	iv := make([]byte, cb.BlockSize())
	rand.Read(iv)

	return cipher.NewCTR(cb, iv)
}

func BenchmarkAes128CtrSmall(b *testing.B) {
	buf := make([]byte, testSmallSize)
	cs := initAesCtr(b, 16)

	for b.Loop() {
		cs.XORKeyStream(buf, buf)
	}
}

func BenchmarkAes128CtrBig(b *testing.B) {
	buf := make([]byte, testBigSize)
	cs := initAesCtr(b, 16)

	for b.Loop() {
		cs.XORKeyStream(buf, buf)
	}
}

func BenchmarkAes256CtrSmall(b *testing.B) {
	buf := make([]byte, testSmallSize)
	cs := initAesCtr(b, 32)

	for b.Loop() {
		cs.XORKeyStream(buf, buf)
	}
}

func BenchmarkAes256CtrBig(b *testing.B) {
	buf := make([]byte, testBigSize)
	cs := initAesCtr(b, 32)

	for b.Loop() {
		cs.XORKeyStream(buf, buf)
	}
}

func initChaCha20(b *testing.B) cipher.Stream {
	b.Helper()

	key := make([]byte, chacha20.KeySize)
	rand.Read(key)

	nonce := make([]byte, chacha20.NonceSize)
	rand.Read(nonce)

	cs, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		b.Fatal(err)
	}
	return cs
}

func BenchmarkChaCha20Small(b *testing.B) {
	buf := make([]byte, testSmallSize)
	cs := initChaCha20(b)

	for b.Loop() {
		cs.XORKeyStream(buf, buf)
	}
}

func BenchmarkChaCha20Big(b *testing.B) {
	buf := make([]byte, testBigSize)
	cs := initChaCha20(b)

	for b.Loop() {
		cs.XORKeyStream(buf, buf)
	}
}
