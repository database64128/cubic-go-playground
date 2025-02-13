package udpheader

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"io"
	"testing"

	"golang.org/x/crypto/hkdf"
	"lukechampine.com/blake3"
)

var (
	key       = make([]byte, 32)
	plaintext = make([]byte, aes.BlockSize)
	c         cipher.Block
)

func init() {
	rand.Read(key)
	rand.Read(plaintext)

	var err error
	c, err = aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
}

func BenchmarkGenSaltHkdfSha1(b *testing.B) {
	salt := make([]byte, 32)
	subkey := make([]byte, 32)

	for b.Loop() {
		rand.Read(salt)

		r := hkdf.New(sha1.New, key, salt, []byte("ss-subkey"))

		if _, err := io.ReadFull(r, subkey); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenSaltBlake3(b *testing.B) {
	keyMaterial := make([]byte, 64)
	copy(keyMaterial, key)
	subkey := make([]byte, 32)

	for b.Loop() {
		rand.Read(keyMaterial[32:])

		blake3.DeriveKey(subkey, "shadowsocks 2022 session subkey", keyMaterial)
	}
}

func BenchmarkAesEcbHeaderEncryption(b *testing.B) {
	ciphertext := make([]byte, 16)

	for b.Loop() {
		c.Encrypt(ciphertext, plaintext)
	}
}

func BenchmarkAesEcbHeaderDecryption(b *testing.B) {
	ciphertext := make([]byte, 16)
	c.Encrypt(ciphertext, plaintext)

	decrypted := make([]byte, 16)

	for b.Loop() {
		c.Decrypt(decrypted, ciphertext)
	}
}
