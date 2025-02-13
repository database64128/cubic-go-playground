package ecdh

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"testing"
	"unsafe"

	"lukechampine.com/blake3"
)

type publicKeyHeader struct {
	curve     ecdh.Curve
	publicKey []byte
	boring    unsafe.Pointer
}

func BenchmarkX25519(b *testing.B) {
	key, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	pubkey := key.PublicKey()
	pubkeyHeader := (*publicKeyHeader)(unsafe.Pointer(pubkey))

	for b.Loop() {
		pubkeyHeader.publicKey, err = key.ECDH(pubkey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func handshakeSetup() (h *blake3.Hasher, clientKey, serverKey *ecdh.PrivateKey, clientPubkey, serverPubkey *ecdh.PublicKey, err error) {
	psk := make([]byte, 32)
	rand.Read(psk)
	h = blake3.New(64, psk)

	clientKey, err = ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return
	}
	clientPubkey = clientKey.PublicKey()

	serverKey, err = ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return
	}
	serverPubkey = serverKey.PublicKey()

	return
}

func clientInitiate(h *blake3.Hasher, clientKey *ecdh.PrivateKey, serverPubkey *ecdh.PublicKey) (clientEphemeralKey *ecdh.PrivateKey, clientEphemeralPubkey *ecdh.PublicKey, sum0 []byte, err error) {
	clientEphemeralKey, err = ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return
	}
	clientEphemeralPubkey = clientEphemeralKey.PublicKey()

	// es
	result, err := clientEphemeralKey.ECDH(serverPubkey)
	if err != nil {
		return
	}
	h.Write(result)

	// ss
	result, err = clientKey.ECDH(serverPubkey)
	if err != nil {
		return
	}
	h.Write(result)

	sum0 = h.Sum(nil)
	h.Reset()
	return
}

func serverRespond(h *blake3.Hasher, serverKey *ecdh.PrivateKey, clientPubkey, clientEphemeralPubkey *ecdh.PublicKey) (serverEphemeralPubkey *ecdh.PublicKey, sum1, sum2 []byte, err error) {
	// es
	result, err := serverKey.ECDH(clientEphemeralPubkey)
	if err != nil {
		return
	}
	h.Write(result)

	// ss
	result, err = serverKey.ECDH(clientPubkey)
	if err != nil {
		return
	}
	h.Write(result)

	sum1 = h.Sum(nil)
	h.Reset()

	serverEphemeralKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return
	}
	serverEphemeralPubkey = serverEphemeralKey.PublicKey()

	// ee
	result, err = serverEphemeralKey.ECDH(clientEphemeralPubkey)
	if err != nil {
		return
	}
	h.Write(result)

	// se
	result, err = serverEphemeralKey.ECDH(clientPubkey)
	if err != nil {
		return
	}
	h.Write(result)

	sum2 = h.Sum(nil)
	h.Reset()
	return
}

func clientRespond(h *blake3.Hasher, clientKey, clientEphemeralKey *ecdh.PrivateKey, serverEphemeralPubkey *ecdh.PublicKey) (sum3 []byte, err error) {
	// ee
	result, err := clientEphemeralKey.ECDH(serverEphemeralPubkey)
	if err != nil {
		return
	}
	h.Write(result)

	// se
	result, err = clientKey.ECDH(serverEphemeralPubkey)
	if err != nil {
		return
	}
	h.Write(result)

	sum3 = h.Sum(nil)
	h.Reset()
	return
}

func TestHandshake(t *testing.T) {
	h, clientKey, serverKey, clientPubkey, serverPubkey, err := handshakeSetup()
	if err != nil {
		t.Fatal(err)
	}

	clientEphemeralKey, clientEphemeralPubkey, sum0, err := clientInitiate(h, clientKey, serverPubkey)
	if err != nil {
		t.Fatal(err)
	}

	serverEphemeralPubkey, sum1, sum2, err := serverRespond(h, serverKey, clientPubkey, clientEphemeralPubkey)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(sum0, sum1) {
		t.Fatal("sum0 != sum1")
	}

	sum3, err := clientRespond(h, clientKey, clientEphemeralKey, serverEphemeralPubkey)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(sum2, sum3) {
		t.Fatal("sum2 != sum3")
	}
}

func BenchmarkHandshake(b *testing.B) {
	h, clientKey, serverKey, clientPubkey, serverPubkey, err := handshakeSetup()
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		clientEphemeralKey, clientEphemeralPubkey, sum0, err := clientInitiate(h, clientKey, serverPubkey)
		if err != nil {
			b.Fatal(err)
		}

		serverEphemeralPubkey, sum1, sum2, err := serverRespond(h, serverKey, clientPubkey, clientEphemeralPubkey)
		if err != nil {
			b.Fatal(err)
		}
		if !bytes.Equal(sum0, sum1) {
			b.Fatal("sum0 != sum1")
		}

		sum3, err := clientRespond(h, clientKey, clientEphemeralKey, serverEphemeralPubkey)
		if err != nil {
			b.Fatal(err)
		}
		if !bytes.Equal(sum2, sum3) {
			b.Fatal("sum2 != sum3")
		}
	}
}
