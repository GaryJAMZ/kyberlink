// Copyright 2026 José Antonio Garibay Marcelo
// SPDX-License-Identifier: Apache-2.0

package gateway

type KyberLinkRequest struct {
	Version          int    `json:"v"`
	SessionID        string `json:"sessionID"`
	ClientPublicKey  string `json:"clientPublicKey"`
	SecretCiphertext string `json:"secretCiphertext"`
	EncryptedData    string `json:"encryptedData"`
}

type KyberLinkResponse struct {
	Version          int    `json:"v"`
	SecretCiphertext string `json:"secretCiphertext"`
	EncryptedData    string `json:"encryptedData"`
}

type GatewayRequestPayload struct {
	FinalAPI  string      `json:"finalApi"`
	Method    string      `json:"method"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
	Nonce     string      `json:"nonce"`
}
