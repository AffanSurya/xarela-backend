package tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/AffanSurya/xarela-backend/internal/service"
)

func TestLocalKMSEncryptDecryptRoundTrip(t *testing.T) {
	kms, err := service.NewLocalKMS("super-secret-key")
	if err != nil {
		t.Fatalf("new local kms: %v", err)
	}

	plaintext := []byte(`{"provider":"binance","api_key":"abc123"}`)
	sealed, err := kms.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if string(sealed) == string(plaintext) {
		t.Fatal("expected ciphertext to differ from plaintext")
	}

	opened, err := kms.Decrypt(sealed)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(opened) != string(plaintext) {
		t.Fatalf("unexpected plaintext %q", string(opened))
	}
	if kms.KeyID() == "" {
		t.Fatal("expected non-empty key id")
	}
}

func TestCredentialCipherSealOpen(t *testing.T) {
	kms, err := service.NewLocalKMS("super-secret-key")
	if err != nil {
		t.Fatalf("new local kms: %v", err)
	}

	cipher := service.NewCredentialCipher(kms)
	record, err := cipher.Seal(map[string]any{
		"api_key": "abc123",
		"secret":  "s3cr3t",
	})
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	if record.CredentialsKeyID != kms.KeyID() {
		t.Fatalf("expected key id %q, got %q", kms.KeyID(), record.CredentialsKeyID)
	}
	if record.EncryptedCredentials["ciphertext"] == "" {
		t.Fatal("expected encrypted payload to be present")
	}

	opened, err := cipher.Open(record)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if !reflect.DeepEqual(opened["api_key"], "abc123") {
		t.Fatalf("unexpected api key %#v", opened["api_key"])
	}
	if !reflect.DeepEqual(opened["secret"], "s3cr3t") {
		t.Fatalf("unexpected secret %#v", opened["secret"])
	}
}

func TestCredentialCipherRotate(t *testing.T) {
	oldKMS, err := service.NewLocalKMS("old-secret")
	if err != nil {
		t.Fatalf("new old kms: %v", err)
	}
	newKMS, err := service.NewLocalKMS("new-secret")
	if err != nil {
		t.Fatalf("new new kms: %v", err)
	}

	oldCipher := service.NewCredentialCipher(oldKMS)
	currentCipher := service.NewCredentialCipher(service.NewKeyRing(newKMS, oldKMS))

	seededAt := time.Date(2026, time.April, 30, 12, 0, 0, 0, time.UTC)
	rotatedAt := seededAt.Add(2 * time.Hour)

	record, err := oldCipher.Seal(map[string]any{
		"provider": "kraken",
		"api_key":  "legacy-key",
	})
	if err != nil {
		t.Fatalf("seal old record: %v", err)
	}

	rotated, changed, err := currentCipher.Rotate(record, rotatedAt)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if !changed {
		t.Fatal("expected rotation to occur")
	}
	if rotated.CredentialsKeyID != newKMS.KeyID() {
		t.Fatalf("expected new key id %q, got %q", newKMS.KeyID(), rotated.CredentialsKeyID)
	}
	if rotated.CredentialsRotatedAt == nil || !rotated.CredentialsRotatedAt.Equal(rotatedAt) {
		t.Fatalf("expected rotation timestamp %v, got %#v", rotatedAt, rotated.CredentialsRotatedAt)
	}

	opened, err := currentCipher.Open(rotated)
	if err != nil {
		t.Fatalf("open rotated record: %v", err)
	}
	if opened["provider"] != "kraken" {
		t.Fatalf("unexpected provider %#v", opened["provider"])
	}
	if opened["api_key"] != "legacy-key" {
		t.Fatalf("unexpected api key %#v", opened["api_key"])
	}
}
