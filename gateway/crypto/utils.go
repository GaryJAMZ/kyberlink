// Copyright 2026 José Antonio Garibay Marcelo
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func Base64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func EncodeHex(b []byte) string {
	return hex.EncodeToString(b)
}

func DecodeHex(hexStr string) ([]byte, error) {
	if len(hexStr)%2 != 0 {
		return nil, fmt.Errorf("invalid hex length: %d", len(hexStr))
	}

	result := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		high := hexCharToByte(hexStr[i])
		low := hexCharToByte(hexStr[i+1])
		if high == 255 || low == 255 {
			return nil, fmt.Errorf("invalid hex char at position %d", i)
		}
		result[i/2] = high<<4 | low
	}
	return result, nil
}

func hexCharToByte(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	default:
		return 255
	}
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
