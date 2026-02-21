// Copyright 2026 José Antonio Garibay Marcelo
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"fmt"
	"log"
	"sync"
	"time"

	"kyberlink-gateway/crypto"

	kcq "github.com/cloudflare/circl/kem/mlkem/mlkem1024"
)

type SessionKey struct {
	SessionID  string
	PublicKey  *kcq.PublicKey
	PrivateKey *kcq.PrivateKey
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*SessionKey
	ttl      time.Duration
}

var (
	GlobalSessionStore *SessionStore
	once               sync.Once
)

func InitSessionStore() {
	once.Do(func() {
		GlobalSessionStore = &SessionStore{
			sessions: make(map[string]*SessionKey),
			ttl:      5 * time.Minute,
		}
		go GlobalSessionStore.cleanupLoop()
	})
}

func (s *SessionStore) GenerateSession() (string, *kcq.PublicKey, error) {
	pub, priv, err := crypto.GenerateKemKey()
	if err != nil {
		return "", nil, err
	}

	sessionID, err := generateRandomID()
	if err != nil {
		return "", nil, err
	}

	key := &SessionKey{
		SessionID:  sessionID,
		PublicKey:  pub,
		PrivateKey: priv,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(s.ttl),
	}

	s.mu.Lock()
	s.sessions[sessionID] = key
	s.mu.Unlock()

	log.Printf("[SESSION] Created: %s...", sessionID[:8])

	return sessionID, pub, nil
}

func (s *SessionStore) GetSessionPrivateKey(sessionID string) (*kcq.PrivateKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(key.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	return key.PrivateKey, nil
}

func (s *SessionStore) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		s.cleanup()
	}
}

func (s *SessionStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, key := range s.sessions {
		if now.After(key.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}

func (s *SessionStore) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.sessions[sessionID]; exists {
		delete(s.sessions, sessionID)
		log.Printf("[SESSION] Burned: %s...", sessionID[:8])
	}
}

// 256-bit random hex string (32 bytes)
func generateRandomID() (string, error) {
	b, err := crypto.GenerateRandomBytes(32)
	if err != nil {
		return "", err
	}
	return crypto.EncodeHex(b), nil
}
