package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// Encrypt encrypts plaintext with AES-256-GCM using a key derived from the shared secret via HKDF-SHA256.
// Returns (ciphertext_with_auth_tag, nonce). The auth tag (16 bytes) is appended to ciphertext.
func Encrypt(sharedSecret []byte, plaintext []byte) (ciphertext []byte, nonce []byte, err error) {
	hk := hkdf.New(sha256.New, sharedSecret, nil, nil)
	key := make([]byte, 32)
	if _, err := io.ReadFull(hk, key); err != nil {
		return nil, nil, fmt.Errorf("HKDF: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("AES: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("GCM: %w", err)
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("nonce: %w", err)
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}
