package swgp

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	mrand "math/rand"
	"testing"

	"golang.org/x/crypto/chacha20poly1305"
	"lukechampine.com/blake3"
)

const (
	wireguardHandshakeInitiationMessageLength = 148
	maxPaddingLength                          = 1024
	maxPacketLength                           = 1452
)

var (
	key       []byte
	aesecb    cipher.Block
	xc20p1305 cipher.AEAD
	blake3xof *blake3.OutputReader
)

func init() {
	key = make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}

	aesecb, err = aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	xc20p1305, err = chacha20poly1305.NewX(key)
	if err != nil {
		panic(err)
	}

	h := blake3.New(24, key)
	blake3xof = h.XOF()
}

func writeWgHsInit(b *testing.B, buf []byte) {
	buf[0] = 1
	_, err := rand.Read(buf[1:wireguardHandshakeInitiationMessageLength])
	if err != nil {
		b.Fatal(err)
	}
}

func writeWgData(b *testing.B, buf []byte) {
	buf[0] = 4
	_, err := rand.Read(buf[1:])
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkZeroOverheadAesEncryptPartialWgHsInit(b *testing.B) {
	buf := make([]byte, maxPacketLength)
	writeWgHsInit(b, buf)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		aesecb.Encrypt(buf[:16], buf[:16])
	}
}

func BenchmarkZeroOverheadAesEncryptPartialWgHsInitRandomPadding(b *testing.B) {
	buf := make([]byte, maxPacketLength)
	writeWgHsInit(b, buf)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		aesecb.Encrypt(buf[:16], buf[:16])

		paddingLen := mrand.Intn(maxPaddingLength + 1)
		_, err := rand.Read(buf[wireguardHandshakeInitiationMessageLength : wireguardHandshakeInitiationMessageLength+paddingLen])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkZeroOverheadAesEncryptPartialWgHsInitBlake3KeyedHashPadding(b *testing.B) {
	buf := make([]byte, maxPacketLength)
	writeWgHsInit(b, buf)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		aesecb.Encrypt(buf[:16], buf[:16])

		paddingLen := mrand.Intn(maxPaddingLength + 1)
		_, err := blake3xof.Read(buf[wireguardHandshakeInitiationMessageLength : wireguardHandshakeInitiationMessageLength+paddingLen])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkZeroOverheadAesEncryptPartialWgData(b *testing.B) {
	buf := make([]byte, maxPacketLength)
	writeWgData(b, buf)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		aesecb.Encrypt(buf[:16], buf[:16])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptPartialWgHsInitRandomNonce(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	overhead := xc20p1305.Overhead()
	buf := make([]byte, maxPacketLength)
	wg := make([]byte, wireguardHandshakeInitiationMessageLength)
	writeWgHsInit(b, wg)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		validLen := nonceSize + wireguardHandshakeInitiationMessageLength + overhead
		nonce := buf[:nonceSize]

		// Generate nonce
		_, err := rand.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Copy plaintext after 16 bytes
		copy(buf[nonceSize+16+overhead:], wg[16:])

		// Seal first 16 bytes
		xc20p1305.Seal(nonce, nonce, wg[:16], buf[nonceSize+16+overhead:validLen])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptPartialWgHsInitBlake3KeyedHashNonce(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	overhead := xc20p1305.Overhead()
	buf := make([]byte, maxPacketLength)
	wg := make([]byte, wireguardHandshakeInitiationMessageLength)
	writeWgHsInit(b, wg)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		validLen := nonceSize + wireguardHandshakeInitiationMessageLength + overhead
		nonce := buf[:nonceSize]

		// Generate nonce
		_, err := blake3xof.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Copy plaintext after 16 bytes
		copy(buf[nonceSize+16+overhead:], wg[16:])

		// Seal first 16 bytes
		xc20p1305.Seal(nonce, nonce, wg[:16], buf[nonceSize+16+overhead:validLen])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptPartialWgHsInitRandomPadding(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	overhead := xc20p1305.Overhead()
	buf := make([]byte, maxPacketLength)
	wg := make([]byte, wireguardHandshakeInitiationMessageLength)
	writeWgHsInit(b, wg)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		paddingLen := mrand.Intn(maxPaddingLength + 1)
		validLen := nonceSize + wireguardHandshakeInitiationMessageLength + overhead
		totalLen := validLen + paddingLen
		nonce := buf[:nonceSize]
		padding := buf[validLen:totalLen]

		// Generate nonce
		_, err := rand.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Copy plaintext after 16 bytes
		copy(buf[nonceSize+16+overhead:], wg[16:])

		// Add padding
		_, err = rand.Read(padding)
		if err != nil {
			b.Fatal(err)
		}

		// Seal first 16 bytes
		xc20p1305.Seal(nonce, nonce, wg[:16], buf[nonceSize+16+overhead:totalLen])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptPartialWgHsInitBlake3KeyedHashPadding(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	overhead := xc20p1305.Overhead()
	buf := make([]byte, maxPacketLength)
	wg := make([]byte, wireguardHandshakeInitiationMessageLength)
	writeWgHsInit(b, wg)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		paddingLen := mrand.Intn(maxPaddingLength + 1)
		validLen := nonceSize + wireguardHandshakeInitiationMessageLength + overhead
		totalLen := validLen + paddingLen
		nonce := buf[:nonceSize]
		padding := buf[validLen:totalLen]

		// Generate nonce
		_, err := blake3xof.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Copy plaintext after 16 bytes
		copy(buf[nonceSize+16+overhead:], wg[16:])

		// Add padding
		_, err = blake3xof.Read(padding)
		if err != nil {
			b.Fatal(err)
		}

		// Seal first 16 bytes
		xc20p1305.Seal(nonce, nonce, wg[:16], buf[nonceSize+16+overhead:totalLen])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgHsInitRandomNonce(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	buf := make([]byte, maxPacketLength)
	writeWgHsInit(b, buf[nonceSize:])

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nonce := buf[:nonceSize]

		// Generate nonce
		_, err := rand.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, buf[nonceSize:nonceSize+wireguardHandshakeInitiationMessageLength], nil)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgHsInitBlake3KeyedHashNonce(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	buf := make([]byte, maxPacketLength)
	writeWgHsInit(b, buf[nonceSize:])

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nonce := buf[:nonceSize]

		// Generate nonce
		_, err := blake3xof.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, buf[nonceSize:nonceSize+wireguardHandshakeInitiationMessageLength], nil)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgHsInitRandomPadding(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	overhead := xc20p1305.Overhead()
	buf := make([]byte, maxPacketLength)
	writeWgHsInit(b, buf[nonceSize:])

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		paddingLen := mrand.Intn(maxPaddingLength + 1)
		validLen := nonceSize + wireguardHandshakeInitiationMessageLength + overhead
		totalLen := validLen + paddingLen
		nonce := buf[:nonceSize]
		padding := buf[validLen:totalLen]

		// Generate nonce
		_, err := rand.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Add padding
		_, err = rand.Read(padding)
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, buf[nonceSize:nonceSize+wireguardHandshakeInitiationMessageLength], padding)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgHsInitBlake3KeyedHashPadding(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	overhead := xc20p1305.Overhead()
	buf := make([]byte, maxPacketLength)
	writeWgHsInit(b, buf[nonceSize:])

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		paddingLen := mrand.Intn(maxPaddingLength + 1)
		validLen := nonceSize + wireguardHandshakeInitiationMessageLength + overhead
		totalLen := validLen + paddingLen
		nonce := buf[:nonceSize]
		padding := buf[validLen:totalLen]

		// Generate nonce
		_, err := blake3xof.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Add padding
		_, err = blake3xof.Read(padding)
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, buf[nonceSize:nonceSize+wireguardHandshakeInitiationMessageLength], padding)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgHsInitEncryptPaddingRandomNonce(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	buf := make([]byte, maxPacketLength)
	writeWgHsInit(b, buf[nonceSize:])

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		paddingLen := mrand.Intn(maxPaddingLength + 1)
		nonce := buf[:nonceSize]

		// Generate nonce
		_, err := rand.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, buf[nonceSize:nonceSize+wireguardHandshakeInitiationMessageLength+paddingLen], nil)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgHsInitEncryptPaddingBlake3KeyedHashNonce(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	buf := make([]byte, maxPacketLength)
	writeWgHsInit(b, buf[nonceSize:])

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		paddingLen := mrand.Intn(maxPaddingLength + 1)
		nonce := buf[:nonceSize]

		// Generate nonce
		_, err := blake3xof.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, buf[nonceSize:nonceSize+wireguardHandshakeInitiationMessageLength+paddingLen], nil)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptPartialWgDataRandomNonce(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	overhead := xc20p1305.Overhead()
	buf := make([]byte, maxPacketLength)
	wg := make([]byte, maxPacketLength-nonceSize-overhead)
	writeWgData(b, wg)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nonce := buf[:nonceSize]

		// Generate nonce
		_, err := rand.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Copy wg packet after 16B
		copy(buf[nonceSize+16+overhead:], wg[16:])

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, wg[:16], buf[nonceSize+16+overhead:nonceSize+len(wg)+overhead])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptPartialWgDataBlake3KeyedHashNonce(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	overhead := xc20p1305.Overhead()
	buf := make([]byte, maxPacketLength)
	wg := make([]byte, maxPacketLength-nonceSize-overhead)
	writeWgData(b, wg)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nonce := buf[:nonceSize]

		// Generate nonce
		_, err := blake3xof.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Copy wg packet after 16B
		copy(buf[nonceSize+16+overhead:], wg[16:])

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, wg[:16], buf[nonceSize+16+overhead:nonceSize+len(wg)+overhead])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgDataRandomNonce(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	overhead := xc20p1305.Overhead()
	buf := make([]byte, maxPacketLength)
	plaintext := buf[nonceSize : maxPacketLength-nonceSize-overhead]
	writeWgData(b, plaintext)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nonce := buf[:nonceSize]

		// Generate nonce
		_, err := rand.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, plaintext, nil)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgDataBlake3KeyedHashNonce(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	overhead := xc20p1305.Overhead()
	buf := make([]byte, maxPacketLength)
	plaintext := buf[nonceSize : maxPacketLength-nonceSize-overhead]
	writeWgData(b, plaintext)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nonce := buf[:nonceSize]

		// Generate nonce
		_, err := blake3xof.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, plaintext, nil)
	}
}
