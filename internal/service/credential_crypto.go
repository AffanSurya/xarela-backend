package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	credentialEnvelopeVersion = 1
	encryptedNonceSize        = 12
	defaultKMSSecretEnv       = "KMS_SECRET"
)

var (
	ErrEmptyKMSSecret             = errors.New("kms secret is required")
	ErrEncryptedCredentialsAbsent = errors.New("encrypted credentials are missing")
	ErrUnknownCredentialKey       = errors.New("unknown encrypted credential key")
)

type KMS interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
	KeyID() string
}

type LocalKMS struct {
	keyID string
	aead  cipher.AEAD
}

type KeyRing struct {
	current  KMS
	previous map[string]KMS
}

type EncryptedCredentialRecord struct {
	EncryptedCredentials map[string]any
	CredentialsKeyID     string
	CredentialsRotatedAt *time.Time
}

type CredentialCipher struct {
	kms KMS
}

func NewLocalKMS(secret string) (*LocalKMS, error) {
	trimmed := strings.TrimSpace(secret)
	if trimmed == "" {
		return nil, ErrEmptyKMSSecret
	}

	key := sha256.Sum256([]byte(trimmed))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("create block cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	return &LocalKMS{
		keyID: keyIDFromSecret(trimmed),
		aead:  aead,
	}, nil
}

func NewLocalKMSFromEnv() (*LocalKMS, error) {
	return NewLocalKMS(os.Getenv(defaultKMSSecretEnv))
}

func NewKeyRing(current KMS, previous ...KMS) *KeyRing {
	entries := make(map[string]KMS, len(previous))
	for _, kms := range previous {
		if kms == nil {
			continue
		}
		entries[kms.KeyID()] = kms
	}

	return &KeyRing{
		current:  current,
		previous: entries,
	}
}

func NewCredentialCipher(kms KMS) *CredentialCipher {
	return &CredentialCipher{kms: kms}
}

func (k *LocalKMS) Encrypt(plaintext []byte) ([]byte, error) {
	if k == nil || k.aead == nil {
		return nil, ErrUnknownCredentialKey
	}

	nonce := make([]byte, encryptedNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	sealed := k.aead.Seal(nil, nonce, plaintext, nil)
	envelope := encryptedCredentialEnvelope{
		Version:    credentialEnvelopeVersion,
		KeyID:      k.keyID,
		Nonce:      base64.RawStdEncoding.EncodeToString(nonce),
		Ciphertext: base64.RawStdEncoding.EncodeToString(sealed),
	}

	encoded, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("marshal encrypted payload: %w", err)
	}

	return encoded, nil
}

func (k *LocalKMS) Decrypt(ciphertext []byte) ([]byte, error) {
	if k == nil || k.aead == nil {
		return nil, ErrUnknownCredentialKey
	}

	envelope, err := decodeEncryptedCredentialEnvelope(ciphertext)
	if err != nil {
		return nil, err
	}
	if envelope.KeyID != k.keyID {
		return nil, ErrUnknownCredentialKey
	}

	nonce, err := base64.RawStdEncoding.DecodeString(envelope.Nonce)
	if err != nil {
		return nil, fmt.Errorf("decode nonce: %w", err)
	}
	sealed, err := base64.RawStdEncoding.DecodeString(envelope.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}

	plaintext, err := k.aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt payload: %w", err)
	}

	return plaintext, nil
}

func (k *LocalKMS) KeyID() string {
	if k == nil {
		return ""
	}
	return k.keyID
}

func (k *KeyRing) Encrypt(plaintext []byte) ([]byte, error) {
	if k == nil || k.current == nil {
		return nil, ErrUnknownCredentialKey
	}
	return k.current.Encrypt(plaintext)
}

func (k *KeyRing) Decrypt(ciphertext []byte) ([]byte, error) {
	envelope, err := decodeEncryptedCredentialEnvelope(ciphertext)
	if err != nil {
		return nil, err
	}

	if k == nil || k.current == nil {
		return nil, ErrUnknownCredentialKey
	}
	if envelope.KeyID == k.current.KeyID() {
		return k.current.Decrypt(ciphertext)
	}
	if previous, ok := k.previous[envelope.KeyID]; ok {
		return previous.Decrypt(ciphertext)
	}

	return nil, ErrUnknownCredentialKey
}

func (k *KeyRing) KeyID() string {
	if k == nil || k.current == nil {
		return ""
	}
	return k.current.KeyID()
}

func (c *CredentialCipher) Seal(credentials map[string]any) (EncryptedCredentialRecord, error) {
	if c == nil || c.kms == nil {
		return EncryptedCredentialRecord{}, ErrUnknownCredentialKey
	}

	plaintext, err := json.Marshal(credentials)
	if err != nil {
		return EncryptedCredentialRecord{}, fmt.Errorf("marshal credentials: %w", err)
	}

	sealed, err := c.kms.Encrypt(plaintext)
	if err != nil {
		return EncryptedCredentialRecord{}, err
	}

	envelope, err := decodeEncryptedCredentialEnvelope(sealed)
	if err != nil {
		return EncryptedCredentialRecord{}, err
	}

	return EncryptedCredentialRecord{
		EncryptedCredentials: map[string]any{
			"v":          envelope.Version,
			"key_id":     envelope.KeyID,
			"nonce":      envelope.Nonce,
			"ciphertext": envelope.Ciphertext,
		},
		CredentialsKeyID: envelope.KeyID,
	}, nil
}

func (c *CredentialCipher) Open(record EncryptedCredentialRecord) (map[string]any, error) {
	if c == nil || c.kms == nil {
		return nil, ErrUnknownCredentialKey
	}

	sealed, err := encodeEncryptedCredentialRecord(record.EncryptedCredentials)
	if err != nil {
		return nil, err
	}

	plaintext, err := c.kms.Decrypt(sealed)
	if err != nil {
		return nil, err
	}

	credentials := make(map[string]any)
	if err := json.Unmarshal(plaintext, &credentials); err != nil {
		return nil, fmt.Errorf("unmarshal credentials: %w", err)
	}

	return credentials, nil
}

func (c *CredentialCipher) Rotate(record EncryptedCredentialRecord, rotatedAt time.Time) (EncryptedCredentialRecord, bool, error) {
	if c == nil || c.kms == nil {
		return EncryptedCredentialRecord{}, false, ErrUnknownCredentialKey
	}
	if record.CredentialsKeyID == c.kms.KeyID() {
		return record, false, nil
	}

	credentials, err := c.Open(record)
	if err != nil {
		return EncryptedCredentialRecord{}, false, err
	}

	rotated, err := c.Seal(credentials)
	if err != nil {
		return EncryptedCredentialRecord{}, false, err
	}
	rotated.CredentialsRotatedAt = &rotatedAt

	return rotated, true, nil
}

type encryptedCredentialEnvelope struct {
	Version    int    `json:"v"`
	KeyID      string `json:"key_id"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

func decodeEncryptedCredentialEnvelope(ciphertext []byte) (encryptedCredentialEnvelope, error) {
	var envelope encryptedCredentialEnvelope
	if len(strings.TrimSpace(string(ciphertext))) == 0 {
		return envelope, ErrEncryptedCredentialsAbsent
	}
	if err := json.Unmarshal(ciphertext, &envelope); err != nil {
		return envelope, fmt.Errorf("decode encrypted payload: %w", err)
	}
	if envelope.KeyID == "" || envelope.Nonce == "" || envelope.Ciphertext == "" {
		return envelope, ErrEncryptedCredentialsAbsent
	}
	if envelope.Version != credentialEnvelopeVersion {
		return envelope, fmt.Errorf("unsupported encrypted payload version %d", envelope.Version)
	}

	return envelope, nil
}

func encodeEncryptedCredentialRecord(record map[string]any) ([]byte, error) {
	if len(record) == 0 {
		return nil, ErrEncryptedCredentialsAbsent
	}

	normalized := map[string]any{
		"v":          asInt(record["v"]),
		"key_id":     fmt.Sprint(record["key_id"]),
		"nonce":      fmt.Sprint(record["nonce"]),
		"ciphertext": fmt.Sprint(record["ciphertext"]),
	}
	if normalized["v"] == 0 {
		normalized["v"] = credentialEnvelopeVersion
	}

	encoded, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal encrypted payload: %w", err)
	}

	return encoded, nil
}

func asInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int8:
		return int(typed)
	case int16:
		return int(typed)
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}

func keyIDFromSecret(secret string) string {
	checksum := sha256.Sum256([]byte(strings.TrimSpace(secret)))
	return hex.EncodeToString(checksum[:8])
}
