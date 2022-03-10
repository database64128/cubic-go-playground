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
	key       []byte
	plaintext []byte
	c         cipher.Block
)

func init() {
	key = make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}

	plaintext = make([]byte, 16)
	_, err = rand.Read(plaintext)
	if err != nil {
		panic(err)
	}

	c, err = aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
}

func BenchmarkGenSaltHkdfSha1(b *testing.B) {
	salt := make([]byte, 32)
	subkey := make([]byte, 32)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := rand.Read(salt)
		if err != nil {
			b.Fatal(err)
		}

		r := hkdf.New(sha1.New, key, salt, []byte("ss-subkey"))

		_, err = io.ReadFull(r, subkey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenSaltBlake3(b *testing.B) {
	keyMaterial := make([]byte, 64)
	copy(keyMaterial, key)
	subkey := make([]byte, 32)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := rand.Read(keyMaterial[32:])
		if err != nil {
			b.Fatal(err)
		}

		blake3.DeriveKey(subkey, "shadowsocks 2022 session subkey", keyMaterial)
	}
}

func BenchmarkAesEcbHeaderEncryption(b *testing.B) {
	ciphertext := make([]byte, 16)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Encrypt(ciphertext, plaintext)
	}
}

func BenchmarkAesEcbHeaderDecryption(b *testing.B) {
	ciphertext := make([]byte, 16)
	c.Encrypt(ciphertext, plaintext)

	decrypted := make([]byte, 16)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Decrypt(decrypted, ciphertext)
	}
}
