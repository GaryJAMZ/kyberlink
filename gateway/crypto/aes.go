// Copyright 2026 José Antonio Garibay Marcelo
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// EncryptAESGCMWithSession encrypts data with AES-256-GCM.
// ss1 (shared secret) encrypts the payload via HKDF-derived key.
// ss2 (reversed ss) encrypts the salt+IV.
func EncryptAESGCMWithSession(data, sharedSecret []byte, sessionID string) ([]byte, []byte, error) {
	if len(sharedSecret) != 32 {
		return nil, nil, fmt.Errorf("shared secret must be 32 bytes, got %d", len(sharedSecret))
	}

	ss1 := sharedSecret
	ss2 := make([]byte, 32)
	for i := 0; i < 32; i++ {
		ss2[i] = sharedSecret[31-i]
	}

	salt := make([]byte, 16)
	iv := make([]byte, 12)
	if _, err := rand.Read(salt); err != nil {
		return nil, nil, err
	}
	if _, err := rand.Read(iv); err != nil {
		return nil, nil, err
	}

	info := fmt.Sprintf("kyberlink:v1|session=%s", sessionID)
	aesKey, err := deriveAESKeyHKDF(ss1, salt, info)
	if err != nil {
		return nil, nil, fmt.Errorf("HKDF derivation failed: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	encryptedData := gcm.Seal(nil, iv, data, []byte(info))

	saltIvCombined := make([]byte, len(salt)+len(iv))
	copy(saltIvCombined[:16], salt)
	copy(saltIvCombined[16:], iv)

	encryptedSaltIv, err := EncryptWithDirectKey(saltIvCombined, ss2)
	if err != nil {
		return nil, nil, fmt.Errorf("salt/IV encryption failed: %w", err)
	}

	return encryptedData, encryptedSaltIv, nil
}

// DecryptAESGCMWithSession reverses EncryptAESGCMWithSession.
// Format: [2-byte length prefix] + encrypted_salt_iv + encrypted_payload
func DecryptAESGCMWithSession(encryptedData, sharedSecret []byte, sessionID string) ([]byte, error) {
	if len(sharedSecret) != 32 {
		return nil, fmt.Errorf("shared secret must be 32 bytes, got %d", len(sharedSecret))
	}

	ss1 := sharedSecret
	ss2 := make([]byte, 32)
	for i := 0; i < 32; i++ {
		ss2[i] = sharedSecret[31-i]
	}

	if len(encryptedData) < 2 {
		return nil, errors.New("data too short")
	}

	encSaltIvLen := int(encryptedData[0])<<8 | int(encryptedData[1])
	if len(encryptedData) < 2+encSaltIvLen {
		return nil, errors.New("data shorter than declared salt/IV length")
	}

	encryptedSaltIv := encryptedData[2 : 2+encSaltIvLen]
	encryptedPayload := encryptedData[2+encSaltIvLen:]

	saltIvBytes, err := DecryptWithDirectKey(encryptedSaltIv, ss2)
	if err != nil {
		return nil, fmt.Errorf("salt/IV decryption failed: %w", err)
	}

	if len(saltIvBytes) != 28 {
		return nil, errors.New("invalid salt/IV length")
	}

	salt := saltIvBytes[:16]
	iv := saltIvBytes[16:28]

	info := fmt.Sprintf("kyberlink:v1|session=%s", sessionID)
	aesKey, err := deriveAESKeyHKDF(ss1, salt, info)
	if err != nil {
		return nil, fmt.Errorf("HKDF derivation failed: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, iv, encryptedPayload, []byte(info))
	if err != nil {
		return nil, fmt.Errorf("AES-GCM decryption failed: %w", err)
	}

	return plaintext, nil
}

func deriveAESKeyHKDF(sharedSecret, salt []byte, info string) ([]byte, error) {
	r := hkdf.New(sha256.New, sharedSecret, salt, []byte(info))
	key := make([]byte, 32)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, err
	}
	return key, nil
}

// EncryptWithDirectKey encrypts with AES-256-GCM using the key directly. Output: IV + ciphertext.
func EncryptWithDirectKey(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, 12)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, iv, data, nil)

	result := make([]byte, len(iv)+len(ciphertext))
	copy(result[:len(iv)], iv)
	copy(result[len(iv):], ciphertext)

	return result, nil
}

// DecryptWithDirectKey reverses EncryptWithDirectKey. Input: IV + ciphertext.
func DecryptWithDirectKey(encryptedData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(encryptedData) < 12 {
		return nil, errors.New("data too short for IV")
	}

	plaintext, err := gcm.Open(nil, encryptedData[:12], encryptedData[12:], nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}
