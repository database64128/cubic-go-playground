package swgp

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	mrand "math/rand/v2"
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
	key       = make([]byte, 32)
	aesecb    cipher.Block
	xc20p1305 cipher.AEAD
	blake3xof *blake3.OutputReader
)

func init() {
	rand.Read(key)

	var err error
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

func writeWgHsInit(buf []byte) {
	buf[0] = 1
	rand.Read(buf[1:wireguardHandshakeInitiationMessageLength])
}

func writeWgData(buf []byte) {
	buf[0] = 4
	rand.Read(buf[1:])
}

func BenchmarkZeroOverheadAesEncryptPartialWgHsInit(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	writeWgHsInit(buf)

	for b.Loop() {
		aesecb.Encrypt(buf[:16], buf[:16])
	}
}

func BenchmarkZeroOverheadAesEncryptPartialWgHsInitRandomPadding(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	writeWgHsInit(buf)

	for b.Loop() {
		aesecb.Encrypt(buf[:16], buf[:16])

		paddingLen := mrand.IntN(maxPaddingLength + 1)
		rand.Read(buf[wireguardHandshakeInitiationMessageLength : wireguardHandshakeInitiationMessageLength+paddingLen])
	}
}

func BenchmarkZeroOverheadAesEncryptPartialWgHsInitBlake3KeyedHashPadding(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	writeWgHsInit(buf)

	for b.Loop() {
		aesecb.Encrypt(buf[:16], buf[:16])

		paddingLen := mrand.IntN(maxPaddingLength + 1)
		_, err := blake3xof.Read(buf[wireguardHandshakeInitiationMessageLength : wireguardHandshakeInitiationMessageLength+paddingLen])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkZeroOverheadAesEncryptPartialWgData(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	writeWgData(buf)

	for b.Loop() {
		aesecb.Encrypt(buf[:16], buf[:16])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptPartialWgHsInitRandomNonce(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	wg := make([]byte, wireguardHandshakeInitiationMessageLength)
	writeWgHsInit(wg)

	for b.Loop() {
		validLen := chacha20poly1305.NonceSizeX + wireguardHandshakeInitiationMessageLength + chacha20poly1305.Overhead
		nonce := buf[:chacha20poly1305.NonceSizeX]

		// Generate nonce
		rand.Read(nonce)

		// Copy plaintext after 16 bytes
		copy(buf[chacha20poly1305.NonceSizeX+16+chacha20poly1305.Overhead:], wg[16:])

		// Seal first 16 bytes
		xc20p1305.Seal(nonce, nonce, wg[:16], buf[chacha20poly1305.NonceSizeX+16+chacha20poly1305.Overhead:validLen])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptPartialWgHsInitBlake3KeyedHashNonce(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	wg := make([]byte, wireguardHandshakeInitiationMessageLength)
	writeWgHsInit(wg)

	for b.Loop() {
		validLen := chacha20poly1305.NonceSizeX + wireguardHandshakeInitiationMessageLength + chacha20poly1305.Overhead
		nonce := buf[:chacha20poly1305.NonceSizeX]

		// Generate nonce
		_, err := blake3xof.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Copy plaintext after 16 bytes
		copy(buf[chacha20poly1305.NonceSizeX+16+chacha20poly1305.Overhead:], wg[16:])

		// Seal first 16 bytes
		xc20p1305.Seal(nonce, nonce, wg[:16], buf[chacha20poly1305.NonceSizeX+16+chacha20poly1305.Overhead:validLen])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptPartialWgHsInitRandomPadding(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	wg := make([]byte, wireguardHandshakeInitiationMessageLength)
	writeWgHsInit(wg)

	for b.Loop() {
		paddingLen := mrand.IntN(maxPaddingLength + 1)
		validLen := chacha20poly1305.NonceSizeX + wireguardHandshakeInitiationMessageLength + chacha20poly1305.Overhead
		totalLen := validLen + paddingLen
		nonce := buf[:chacha20poly1305.NonceSizeX]
		padding := buf[validLen:totalLen]

		// Generate nonce
		rand.Read(nonce)

		// Copy plaintext after 16 bytes
		copy(buf[chacha20poly1305.NonceSizeX+16+chacha20poly1305.Overhead:], wg[16:])

		// Add padding
		rand.Read(padding)

		// Seal first 16 bytes
		xc20p1305.Seal(nonce, nonce, wg[:16], buf[chacha20poly1305.NonceSizeX+16+chacha20poly1305.Overhead:totalLen])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptPartialWgHsInitBlake3KeyedHashPadding(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	wg := make([]byte, wireguardHandshakeInitiationMessageLength)
	writeWgHsInit(wg)

	for b.Loop() {
		paddingLen := mrand.IntN(maxPaddingLength + 1)
		validLen := chacha20poly1305.NonceSizeX + wireguardHandshakeInitiationMessageLength + chacha20poly1305.Overhead
		totalLen := validLen + paddingLen
		nonce := buf[:chacha20poly1305.NonceSizeX]
		padding := buf[validLen:totalLen]

		// Generate nonce
		_, err := blake3xof.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Copy plaintext after 16 bytes
		copy(buf[chacha20poly1305.NonceSizeX+16+chacha20poly1305.Overhead:], wg[16:])

		// Add padding
		_, err = blake3xof.Read(padding)
		if err != nil {
			b.Fatal(err)
		}

		// Seal first 16 bytes
		xc20p1305.Seal(nonce, nonce, wg[:16], buf[chacha20poly1305.NonceSizeX+16+chacha20poly1305.Overhead:totalLen])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgHsInitRandomNonce(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	writeWgHsInit(buf[chacha20poly1305.NonceSizeX:])

	for b.Loop() {
		nonce := buf[:chacha20poly1305.NonceSizeX]

		// Generate nonce
		rand.Read(nonce)

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, buf[chacha20poly1305.NonceSizeX:chacha20poly1305.NonceSizeX+wireguardHandshakeInitiationMessageLength], nil)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgHsInitBlake3KeyedHashNonce(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	writeWgHsInit(buf[chacha20poly1305.NonceSizeX:])

	for b.Loop() {
		nonce := buf[:chacha20poly1305.NonceSizeX]

		// Generate nonce
		_, err := blake3xof.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, buf[chacha20poly1305.NonceSizeX:chacha20poly1305.NonceSizeX+wireguardHandshakeInitiationMessageLength], nil)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgHsInitRandomPadding(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	writeWgHsInit(buf[chacha20poly1305.NonceSizeX:])

	for b.Loop() {
		paddingLen := mrand.IntN(maxPaddingLength + 1)
		validLen := chacha20poly1305.NonceSizeX + wireguardHandshakeInitiationMessageLength + chacha20poly1305.Overhead
		totalLen := validLen + paddingLen
		nonce := buf[:chacha20poly1305.NonceSizeX]
		padding := buf[validLen:totalLen]

		// Generate nonce
		rand.Read(nonce)

		// Add padding
		rand.Read(padding)

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, buf[chacha20poly1305.NonceSizeX:chacha20poly1305.NonceSizeX+wireguardHandshakeInitiationMessageLength], padding)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgHsInitBlake3KeyedHashPadding(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	writeWgHsInit(buf[chacha20poly1305.NonceSizeX:])

	for b.Loop() {
		paddingLen := mrand.IntN(maxPaddingLength + 1)
		validLen := chacha20poly1305.NonceSizeX + wireguardHandshakeInitiationMessageLength + chacha20poly1305.Overhead
		totalLen := validLen + paddingLen
		nonce := buf[:chacha20poly1305.NonceSizeX]
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
		xc20p1305.Seal(nonce, nonce, buf[chacha20poly1305.NonceSizeX:chacha20poly1305.NonceSizeX+wireguardHandshakeInitiationMessageLength], padding)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgHsInitEncryptPaddingRandomNonce(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	writeWgHsInit(buf[chacha20poly1305.NonceSizeX:])

	for b.Loop() {
		paddingLen := mrand.IntN(maxPaddingLength + 1)
		nonce := buf[:chacha20poly1305.NonceSizeX]

		// Generate nonce
		rand.Read(nonce)

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, buf[chacha20poly1305.NonceSizeX:chacha20poly1305.NonceSizeX+wireguardHandshakeInitiationMessageLength+paddingLen], nil)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgHsInitEncryptPaddingBlake3KeyedHashNonce(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	writeWgHsInit(buf[chacha20poly1305.NonceSizeX:])

	for b.Loop() {
		paddingLen := mrand.IntN(maxPaddingLength + 1)
		nonce := buf[:chacha20poly1305.NonceSizeX]

		// Generate nonce
		_, err := blake3xof.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, buf[chacha20poly1305.NonceSizeX:chacha20poly1305.NonceSizeX+wireguardHandshakeInitiationMessageLength+paddingLen], nil)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptPartialWgDataRandomNonce(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	wg := make([]byte, maxPacketLength-chacha20poly1305.NonceSizeX-chacha20poly1305.Overhead)
	writeWgData(wg)

	for b.Loop() {
		nonce := buf[:chacha20poly1305.NonceSizeX]

		// Generate nonce
		rand.Read(nonce)

		// Copy wg packet after 16B
		copy(buf[chacha20poly1305.NonceSizeX+16+chacha20poly1305.Overhead:], wg[16:])

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, wg[:16], buf[chacha20poly1305.NonceSizeX+16+chacha20poly1305.Overhead:chacha20poly1305.NonceSizeX+len(wg)+chacha20poly1305.Overhead])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptPartialWgDataBlake3KeyedHashNonce(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	wg := make([]byte, maxPacketLength-chacha20poly1305.NonceSizeX-chacha20poly1305.Overhead)
	writeWgData(wg)

	for b.Loop() {
		nonce := buf[:chacha20poly1305.NonceSizeX]

		// Generate nonce
		_, err := blake3xof.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Copy wg packet after 16B
		copy(buf[chacha20poly1305.NonceSizeX+16+chacha20poly1305.Overhead:], wg[16:])

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, wg[:16], buf[chacha20poly1305.NonceSizeX+16+chacha20poly1305.Overhead:chacha20poly1305.NonceSizeX+len(wg)+chacha20poly1305.Overhead])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgDataRandomNonce(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	plaintext := buf[chacha20poly1305.NonceSizeX : maxPacketLength-chacha20poly1305.NonceSizeX-chacha20poly1305.Overhead]
	writeWgData(plaintext)

	for b.Loop() {
		nonce := buf[:chacha20poly1305.NonceSizeX]

		// Generate nonce
		rand.Read(nonce)

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, plaintext, nil)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgDataBlake3KeyedHashNonce(b *testing.B) {
	b.SetBytes(maxPacketLength)

	buf := make([]byte, maxPacketLength)
	plaintext := buf[chacha20poly1305.NonceSizeX : maxPacketLength-chacha20poly1305.NonceSizeX-chacha20poly1305.Overhead]
	writeWgData(plaintext)

	for b.Loop() {
		nonce := buf[:chacha20poly1305.NonceSizeX]

		// Generate nonce
		_, err := blake3xof.Read(nonce)
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		xc20p1305.Seal(nonce, nonce, plaintext, nil)
	}
}
