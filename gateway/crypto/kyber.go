// Copyright 2026 José Antonio Garibay Marcelo
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"crypto/rand"

	kcq "github.com/cloudflare/circl/kem/mlkem/mlkem1024"
)

func GenerateKemKey() (*kcq.PublicKey, *kcq.PrivateKey, error) {
	return kcq.GenerateKeyPair(rand.Reader)
}

func UnmarshalPrivateKey(data []byte) (*kcq.PrivateKey, error) {
	priv, err := kcq.Scheme().UnmarshalBinaryPrivateKey(data)
	if err != nil {
		return nil, err
	}
	return priv.(*kcq.PrivateKey), nil
}

func UnmarshalPublicKey(data []byte) (*kcq.PublicKey, error) {
	pub, err := kcq.Scheme().UnmarshalBinaryPublicKey(data)
	if err != nil {
		return nil, err
	}
	return pub.(*kcq.PublicKey), nil
}

func Decapsulate(privateKey *kcq.PrivateKey, ciphertext []byte) ([]byte, error) {
	return kcq.Scheme().Decapsulate(privateKey, ciphertext)
}

func Encapsulate(publicKey *kcq.PublicKey) ([]byte, []byte, error) {
	return kcq.Scheme().Encapsulate(publicKey)
}
