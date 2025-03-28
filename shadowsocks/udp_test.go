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

var key = make([]byte, 32)

func init() {
	rand.Read(key)
}

func BenchmarkShadowsocksAEADAes256GcmEncryption(b *testing.B) {
	b.SetBytes(testPayloadLength)

	nonce := make([]byte, 12)
	subkey := make([]byte, 32)
	buf := make([]byte, 32+testPayloadLength+16)

	// Random payload
	payload := buf[32 : 32+testPayloadLength]
	rand.Read(payload)

	for b.Loop() {
		// Generate random salt
		rand.Read(buf[:32])

		// Derive subkey
		r := hkdf.New(sha1.New, key, buf[:32], []byte("ss-subkey"))

		if _, err := io.ReadFull(r, subkey); err != nil {
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
	b.SetBytes(testPayloadLength)

	nonce := make([]byte, 12)
	subkey := make([]byte, 32)
	buf := make([]byte, 64+testPayloadLength+16)

	// Random payload
	payload := buf[64 : 64+testPayloadLength]
	rand.Read(payload)

	for b.Loop() {
		// Copy key so buf[:64] can be used as key material
		copy(buf, key)

		// Generate random salt
		rand.Read(buf[32:64])

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
	b.SetBytes(testPayloadLength)

	var counter uint64

	buf := make([]byte, 16+testPayloadLength+16)

	// Random payload
	payload := buf[16 : 16+testPayloadLength]
	rand.Read(payload)

	// Header block cipher
	aesecb, err := aes.NewCipher(key)
	if err != nil {
		b.Fatal(err)
	}

	// AEAD
	keyMaterial := make([]byte, 32+8) // key + session id
	sid := keyMaterial[32:]
	copy(keyMaterial, key)
	rand.Read(sid)

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

	for b.Loop() {
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
	b.SetBytes(testPayloadLength)

	buf := make([]byte, 24+testPayloadLength+16)

	// Random payload
	payload := buf[24 : 24+testPayloadLength]
	rand.Read(payload)

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		// Random nonce
		rand.Read(buf[:24])

		// Seal AEAD
		aead.Seal(payload[:0], buf[:24], payload, nil)
	}
}
