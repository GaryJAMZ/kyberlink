// Copyright 2026 José Antonio Garibay Marcelo
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"kyberlink-gateway/crypto"

	"github.com/gin-gonic/gin"
)

var (
	BackendURL = "http://localhost:34890"
)

func HandleGateway(c *gin.Context) {
	var req KyberLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	log.Printf("[PHASE 2] Secure Request Received")

	sessionID := req.SessionID

	if len(sessionID) > 8 {
		log.Printf("   Session: %s...", sessionID[:8])
	}
	log.Printf("   Kyber CT1: [%d bytes]", len(req.SecretCiphertext))
	log.Printf("   Encrypted Payload: [%d bytes]", len(req.EncryptedData))

	privKey, err := GlobalSessionStore.GetSessionPrivateKey(sessionID)
	if err != nil {
		log.Printf("   [PHASE 2] Session Error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
		return
	}

	defer GlobalSessionStore.DeleteSession(sessionID)

	secretCiphertext, err := crypto.Base64Decode(req.SecretCiphertext)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid secret ciphertext"})
		return
	}

	sharedSecret, err := crypto.Decapsulate(privKey, secretCiphertext)
	if err != nil {
		log.Printf("   [PHASE 2] Decapsulation failure")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Decapsulation failed"})
		return
	}

	log.Printf("   SS1 derived")

	encryptedPayload, err := crypto.Base64Decode(req.EncryptedData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data encoding"})
		return
	}

	decryptedJSON, err := crypto.DecryptAESGCMWithSession(encryptedPayload, sharedSecret, sessionID)
	if err != nil {
		log.Printf("   [PHASE 2] Decryption with SS1 failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Decryption failed"})
		return
	}

	log.Printf("   Payload decrypted (%d bytes)", len(decryptedJSON))

	var payload GatewayRequestPayload
	if err := json.Unmarshal(decryptedJSON, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if GlobalReplayStore.IsReplay(payload.Nonce, payload.Timestamp) {
		log.Printf("   [SECURITY] Replay detected")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Replay Detected"})
		return
	}

	log.Printf("   Destination: %s %s", payload.Method, payload.FinalAPI)

	// SSRF prevention: only relative paths allowed
	targetURL := payload.FinalAPI
	if strings.HasPrefix(targetURL, "http://") || strings.HasPrefix(targetURL, "https://") {
		log.Printf("   [SECURITY] Absolute URL rejected: %s", targetURL)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Absolute URLs not allowed"})
		return
	}
	if !strings.HasPrefix(targetURL, "/") {
		log.Printf("   [SECURITY] Invalid path: %s", targetURL)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API path"})
		return
	}
	targetURL = BackendURL + targetURL

	log.Printf("[PHASE 3] Forwarding to backend: %s", targetURL)

	var bodyReader io.Reader
	if payload.Payload != nil {
		jsonBytes, _ := json.Marshal(payload.Payload)
		bodyReader = bytes.NewBuffer(jsonBytes)
	}

	proxyReq, err := http.NewRequest(payload.Method, targetURL, bodyReader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Proxy Error"})
		return
	}

	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("X-Kyber-Proxy", "true")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("   [PHASE 3] Backend unreachable: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Backend unreachable"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10MB limit
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Read Error"})
		return
	}

	log.Printf("[PHASE 3] Backend response: %d bytes", len(respBody))

	log.Printf("[PHASE 3] Encrypting response with Pub2...")
	clientPubBytes, err := crypto.Base64Decode(req.ClientPublicKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid public_2 key"})
		return
	}

	clientPub, err := crypto.UnmarshalPublicKey(clientPubBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unmarshal public_2 failed"})
		return
	}

	kyberCiphertextResp, sharedSecretResponse, err := crypto.Encapsulate(clientPub)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Response Encap failure"})
		return
	}
	log.Printf("   CT2 generated")

	encData, encSaltIv, err := crypto.EncryptAESGCMWithSession(respBody, sharedSecretResponse, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Response Encryption failure"})
		return
	}

	combinedSecret := make([]byte, len(kyberCiphertextResp)+len(encSaltIv))
	copy(combinedSecret[:len(kyberCiphertextResp)], kyberCiphertextResp)
	copy(combinedSecret[len(kyberCiphertextResp):], encSaltIv)

	log.Printf("[PHASE 4] Sending encrypted response")

	c.JSON(http.StatusOK, KyberLinkResponse{
		Version:          1,
		SecretCiphertext: crypto.Base64Encode(combinedSecret),
		EncryptedData:    crypto.Base64Encode(encData),
	})
}

func HandleInitSession(c *gin.Context) {
	log.Printf("[PHASE 1] New handshake request")
	sessionID, publicKey, err := GlobalSessionStore.GenerateSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Session Generation failed"})
		return
	}

	pubKeyBytes, err := publicKey.MarshalBinary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Key Marshal failed"})
		return
	}

	log.Printf("[PHASE 1] Session %s... created", sessionID[:8])

	c.JSON(http.StatusOK, gin.H{
		"sessionID": sessionID,
		"publicKey": crypto.Base64Encode(pubKeyBytes),
	})
}
