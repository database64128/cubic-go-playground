package shadowsocks

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"io"
	"testing"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
	"lukechampine.com/blake3"
)

const testPayloadLength = 1400

var (
	key []byte
)

func init() {
	key = make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
}

func BenchmarkShadowsocksAEADAes256GcmEncryption(b *testing.B) {
	nonce := make([]byte, 12)
	subkey := make([]byte, 32)
	buf := make([]byte, 32+testPayloadLength+16)

	// Random payload
	payload := buf[32 : 32+testPayloadLength]
	_, err := rand.Read(payload)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generate random salt
		_, err := rand.Read(buf[:32])
		if err != nil {
			b.Fatal(err)
		}

		// Derive subkey
		r := hkdf.New(sha1.New, key, buf[:32], []byte("ss-subkey"))

		_, err = io.ReadFull(r, subkey)
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		cb, err := aes.NewCipher(subkey)
		if err != nil {
			b.Fatal(err)
		}

		aead, err := cipher.NewGCM(cb)
		if err != nil {
			b.Fatal(err)
		}

		aead.Seal(payload[:0], nonce, payload, nil)
	}
}

func BenchmarkShadowsocksAEADAes256GcmWithBlake3Encryption(b *testing.B) {
	nonce := make([]byte, 12)
	subkey := make([]byte, 32)
	buf := make([]byte, 64+testPayloadLength+16)

	// Random payload
	payload := buf[64 : 64+testPayloadLength]
	_, err := rand.Read(payload)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Copy key so buf[:64] can be used as key material
		copy(buf, key)

		// Generate random salt
		_, err := rand.Read(buf[32:64])
		if err != nil {
			b.Fatal(err)
		}

		// Derive subkey
		blake3.DeriveKey(subkey, "shadowsocks 2022 session subkey", buf[:64])

		// Seal AEAD
		cb, err := aes.NewCipher(subkey)
		if err != nil {
			b.Fatal(err)
		}

		aead, err := cipher.NewGCM(cb)
		if err != nil {
			b.Fatal(err)
		}

		aead.Seal(payload[:0], nonce, payload, nil)
	}
}

func BenchmarkDraftSeparateHeaderAes256GcmEncryption(b *testing.B) {
	var counter uint64

	buf := make([]byte, 16+testPayloadLength+16)

	// Random payload
	payload := buf[16 : 16+testPayloadLength]
	_, err := rand.Read(payload)
	if err != nil {
		b.Fatal(err)
	}

	// Header block cipher
	aesecb, err := aes.NewCipher(key)
	if err != nil {
		b.Fatal(err)
	}

	// AEAD
	keyMaterial := make([]byte, 32+8) // key + session id
	sid := keyMaterial[32:]
	copy(keyMaterial, key)
	_, err = rand.Read(sid)
	if err != nil {
		b.Fatal(err)
	}

	subkey := make([]byte, 32)

	blake3.DeriveKey(subkey, "shadowsocks 2022 session subkey", keyMaterial)

	cb, err := aes.NewCipher(subkey)
	if err != nil {
		b.Fatal(err)
	}

	aead, err := cipher.NewGCM(cb)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Session id
		copy(buf, sid)

		// Counter
		binary.BigEndian.PutUint64(buf[8:], counter)
		counter++

		// Header
		aesecb.Encrypt(buf[:16], buf[:16])

		// Seal AEAD
		aead.Seal(payload[:0], buf[4:16], payload, nil)
	}
}

func BenchmarkDraftXChaCha20Poly1305Encryption(b *testing.B) {
	buf := make([]byte, 24+testPayloadLength+16)

	// Random payload
	payload := buf[24 : 24+testPayloadLength]
	_, err := rand.Read(payload)
	if err != nil {
		b.Fatal(err)
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Random nonce
		_, err = rand.Read(buf[:24])
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		aead.Seal(payload[:0], buf[:24], payload, nil)
	}
}
