package dockerops

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"docker-go/domains"
)

const defaultMaxUpgradeBytes int64 = 200 * 1024 * 1024

func (e *Executor) UpgradeAgent(ctx context.Context, rawPayload json.RawMessage) (domains.AgentUpgradeResult, error) {
	payload, err := parseAgentUpgradePayload(rawPayload)
	if err != nil {
		return domains.AgentUpgradeResult{}, err
	}
	if e.agentVersion != "" && payload.Version == e.agentVersion && !payload.Force {
		return domains.AgentUpgradeResult{}, errors.New("agent is already at target version")
	}
	executablePath, err := os.Executable()
	if err != nil {
		return domains.AgentUpgradeResult{}, err
	}
	executablePath, _ = filepath.EvalSymlinks(executablePath)
	dir := filepath.Dir(executablePath)
	tempFile, downloadedBytes, actualSHA, err := downloadUpgradeBinary(ctx, payload.DownloadURL, dir, e.maxUpgradeBytes)
	if err != nil {
		return domains.AgentUpgradeResult{}, err
	}
	defer os.Remove(tempFile)
	if actualSHA != payload.SHA256 {
		return domains.AgentUpgradeResult{}, fmt.Errorf("upgrade sha256 mismatch: expected %s got %s", payload.SHA256, actualSHA)
	}
	signatureVerified := false
	if e.upgradePublicKey != "" || e.requireUpgradeSignature {
		if payload.Signature == "" {
			return domains.AgentUpgradeResult{}, errors.New("upgrade signature is required")
		}
		if err = verifyUpgradeSignature(tempFile, e.upgradePublicKey, payload.Signature); err != nil {
			return domains.AgentUpgradeResult{}, err
		}
		signatureVerified = true
	}
	if err = makeExecutableLike(tempFile, executablePath); err != nil {
		return domains.AgentUpgradeResult{}, err
	}
	backupPath, err := replaceExecutable(executablePath, tempFile)
	if err != nil {
		return domains.AgentUpgradeResult{}, err
	}
	result := domains.AgentUpgradeResult{
		AgentGuid:           payload.AgentGuid,
		PackageGuid:         payload.PackageGuid,
		PreviousVersion:     e.agentVersion,
		TargetVersion:       payload.Version,
		ExecutablePath:      executablePath,
		BackupPath:          backupPath,
		DownloadedBytes:     downloadedBytes,
		SHA256:              actualSHA,
		SignatureVerified:   signatureVerified,
		RestartScheduled:    true,
		RestartDelaySeconds: payload.RestartDelaySeconds,
	}
	go func() {
		time.Sleep(time.Duration(payload.RestartDelaySeconds) * time.Second)
		os.Exit(0)
	}()
	return result, nil
}

func parseAgentUpgradePayload(raw json.RawMessage) (domains.AgentUpgradePayload, error) {
	var payload domains.AgentUpgradePayload
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return payload, err
		}
	}
	payload.AgentGuid = strings.TrimSpace(payload.AgentGuid)
	payload.PackageGuid = strings.TrimSpace(payload.PackageGuid)
	payload.Version = strings.TrimSpace(payload.Version)
	payload.DownloadURL = strings.TrimSpace(payload.DownloadURL)
	payload.SHA256 = strings.ToLower(strings.TrimSpace(payload.SHA256))
	payload.Signature = strings.TrimSpace(payload.Signature)
	if payload.Version == "" {
		return payload, errors.New("missing target agent version")
	}
	if payload.DownloadURL == "" {
		return payload, errors.New("missing upgrade download url")
	}
	parsed, err := url.Parse(payload.DownloadURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return payload, errors.New("invalid upgrade download url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return payload, errors.New("upgrade download url must use http or https")
	}
	if len(payload.SHA256) != 64 {
		return payload, errors.New("sha256 must be 64 hex characters")
	}
	if _, err = hex.DecodeString(payload.SHA256); err != nil {
		return payload, errors.New("sha256 must be hex encoded")
	}
	if payload.RestartDelaySeconds <= 0 {
		payload.RestartDelaySeconds = 3
	}
	if payload.RestartDelaySeconds > 60 {
		payload.RestartDelaySeconds = 60
	}
	return payload, nil
}

func downloadUpgradeBinary(ctx context.Context, downloadURL string, dir string, maxBytes int64) (string, int64, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return "", 0, "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", 0, "", fmt.Errorf("download upgrade binary failed: %s", resp.Status)
	}
	if resp.ContentLength > maxBytes {
		return "", 0, "", errors.New("upgrade binary is too large")
	}
	tmp, err := os.CreateTemp(dir, ".nav-docker-agent-upgrade-*")
	if err != nil {
		return "", 0, "", err
	}
	defer tmp.Close()
	hash := sha256.New()
	limited := io.LimitReader(resp.Body, maxBytes+1)
	written, err := io.Copy(io.MultiWriter(tmp, hash), limited)
	if err != nil {
		_ = os.Remove(tmp.Name())
		return "", 0, "", err
	}
	if written > maxBytes {
		_ = os.Remove(tmp.Name())
		return "", 0, "", errors.New("upgrade binary is too large")
	}
	if err = tmp.Sync(); err != nil {
		_ = os.Remove(tmp.Name())
		return "", 0, "", err
	}
	return tmp.Name(), written, hex.EncodeToString(hash.Sum(nil)), nil
}

func verifyUpgradeSignature(filePath string, publicKeyText string, signatureText string) error {
	publicKey, err := parseEd25519PublicKey(publicKeyText)
	if err != nil {
		return err
	}
	signature, err := decodeBase64Text(signatureText)
	if err != nil {
		return errors.New("invalid upgrade signature")
	}
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if !ed25519.Verify(publicKey, raw, signature) {
		return errors.New("upgrade signature verification failed")
	}
	return nil
}

func parseEd25519PublicKey(value string) (ed25519.PublicKey, error) {
	value = strings.TrimSpace(strings.TrimPrefix(value, "ed25519:"))
	if value == "" {
		return nil, errors.New("missing upgrade public key")
	}
	if block, _ := pem.Decode([]byte(value)); block != nil {
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		publicKey, ok := key.(ed25519.PublicKey)
		if !ok {
			return nil, errors.New("upgrade public key is not ed25519")
		}
		return publicKey, nil
	}
	raw, err := decodeBase64Text(value)
	if err != nil {
		return nil, errors.New("invalid upgrade public key")
	}
	if len(raw) != ed25519.PublicKeySize {
		return nil, errors.New("upgrade public key must be 32 bytes")
	}
	return ed25519.PublicKey(raw), nil
}

func decodeBase64Text(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	var lastErr error
	for _, encoding := range encodings {
		raw, err := encoding.DecodeString(value)
		if err == nil {
			return raw, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func makeExecutableLike(tempFile string, executablePath string) error {
	mode := os.FileMode(0755)
	if info, err := os.Stat(executablePath); err == nil {
		mode = info.Mode().Perm()
	}
	return os.Chmod(tempFile, mode|0700)
}

func replaceExecutable(executablePath string, tempFile string) (string, error) {
	backupPath := fmt.Sprintf("%s.bak.%d", executablePath, time.Now().Unix())
	if err := os.Rename(executablePath, backupPath); err != nil {
		return "", err
	}
	if err := os.Rename(tempFile, executablePath); err != nil {
		_ = os.Rename(backupPath, executablePath)
		return "", err
	}
	return backupPath, nil
}

func normalizeMaxUpgradeBytes(value int64) int64 {
	if value <= 0 {
		return defaultMaxUpgradeBytes
	}
	if value < 1024*1024 {
		return 1024 * 1024
	}
	return value
}
