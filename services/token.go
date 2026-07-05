package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func newAgentToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "nav_agent_" + hex.EncodeToString(buf), nil
}

func hashAgentToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}
