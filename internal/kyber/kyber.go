package kyber

import (
	"fmt"
)

// encapsulator encapsulates a shared secret using a public key.
type encapsulator interface {
	Encapsulate(publicKey []byte) (ciphertext []byte, sharedSecret []byte, err error)
}

// KyberProvider provides Kyber768 post-quantum key encapsulation.
// Uses PQClean ML-KEM 768 (same as Node.js WASM and Rust pqcrypto-kyber) for interoperability.
type KyberProvider struct {
	impl encapsulator
}

// NewKyberProvider creates a new Kyber768 provider.
func NewKyberProvider() *KyberProvider {
	return &KyberProvider{
		impl: newEncapsulator(),
	}
}

// Encapsulate encapsulates a shared secret using the device's post-quantum public key.
// Returns (kyber_ciphertext, shared_secret). Ciphertext is 1088 bytes, shared secret is 32 bytes.
func (k *KyberProvider) Encapsulate(publicKey []byte) (ciphertext []byte, sharedSecret []byte, err error) {
	if k.impl == nil {
		return nil, nil, fmt.Errorf("Kyber provider not available: build with CGO_ENABLED=1 and run 'make pqclean-lib'")
	}
	return k.impl.Encapsulate(publicKey)
}
