package swgp

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	mrand "math/rand"
	"testing"

	"golang.org/x/crypto/chacha20poly1305"
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
}

func makeWgHsInit(b *testing.B) []byte {
	wgHsInit := make([]byte, wireguardHandshakeInitiationMessageLength)
	wgHsInit[0] = 1
	_, err := rand.Read(wgHsInit[1:])
	if err != nil {
		b.Fatal(err)
	}
	return wgHsInit
}

func makeWgData(b *testing.B) []byte {
	wgHsInit := make([]byte, maxPacketLength-40)
	wgHsInit[0] = 4
	_, err := rand.Read(wgHsInit[1:])
	if err != nil {
		b.Fatal(err)
	}
	return wgHsInit
}

func BenchmarkZeroOverheadAesEncryptPartialWgHsInit(b *testing.B) {
	wg := makeWgHsInit(b)
	buf := make([]byte, maxPacketLength)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		aesecb.Encrypt(buf, wg[:16])
		copy(buf[16:], wg[16:])

		paddingLen := mrand.Intn(maxPaddingLength + 1)
		_, err := rand.Read(buf[wireguardHandshakeInitiationMessageLength : wireguardHandshakeInitiationMessageLength+paddingLen])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkZeroOverheadAesEncryptPartialWgData(b *testing.B) {
	wg := makeWgData(b)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		aesecb.Encrypt(wg[:16], wg[:16])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptPartialWgHsInit(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	overhead := xc20p1305.Overhead()
	wg := makeWgHsInit(b)
	buf := make([]byte, maxPacketLength)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		paddingLen := mrand.Intn(maxPaddingLength + 1)
		validLen := nonceSize + wireguardHandshakeInitiationMessageLength + overhead
		totalLen := validLen + paddingLen
		padding := buf[validLen:totalLen]

		// Add padding
		_, err := rand.Read(padding)
		if err != nil {
			b.Fatal(err)
		}

		// Generate nonce
		_, err = rand.Read(buf[:nonceSize])
		if err != nil {
			b.Fatal(err)
		}

		// Copy wg packet after 16B
		copy(buf[nonceSize+16+overhead:], wg[16:])

		// Seal AEAD
		xc20p1305.Seal(buf[nonceSize:nonceSize], buf[:nonceSize], wg[:16], buf[nonceSize+16+overhead:totalLen])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgHsInit(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	overhead := xc20p1305.Overhead()
	wg := makeWgHsInit(b)
	buf := make([]byte, maxPacketLength)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		paddingLen := mrand.Intn(maxPaddingLength + 1)
		validLen := nonceSize + wireguardHandshakeInitiationMessageLength + overhead
		totalLen := validLen + paddingLen
		padding := buf[validLen:totalLen]

		// Add padding
		_, err := rand.Read(padding)
		if err != nil {
			b.Fatal(err)
		}

		// Generate nonce
		_, err = rand.Read(buf[:nonceSize])
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		xc20p1305.Seal(buf[nonceSize:nonceSize], buf[:nonceSize], wg, padding)
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptPartialWgData(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	overhead := xc20p1305.Overhead()
	wg := makeWgData(b)
	buf := make([]byte, maxPacketLength)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generate nonce
		_, err := rand.Read(buf[:nonceSize])
		if err != nil {
			b.Fatal(err)
		}

		// Copy wg packet after 16B
		copy(buf[nonceSize+16+overhead:], wg[16:])

		// Seal AEAD
		xc20p1305.Seal(buf[nonceSize:nonceSize], buf[:nonceSize], wg[:16], buf[nonceSize+16+overhead:nonceSize+len(wg)+overhead])
	}
}

func BenchmarkParanoidXChaCha20Poly1305EncryptFullWgData(b *testing.B) {
	nonceSize := xc20p1305.NonceSize()
	wg := makeWgData(b)
	buf := make([]byte, maxPacketLength)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generate nonce
		_, err := rand.Read(buf[:nonceSize])
		if err != nil {
			b.Fatal(err)
		}

		// Seal AEAD
		xc20p1305.Seal(buf[nonceSize:nonceSize], buf[:nonceSize], wg, nil)
	}
}
