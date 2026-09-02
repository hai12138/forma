/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
	"strconv"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

type SecretProvider interface {
	StoreSecret(context.Context, string, string, int32, []byte) ([]byte, error)
	ResolveSecret(context.Context, string, string, int32, []byte) ([]byte, error)
	RotateSecret(context.Context, string, string, int32, []byte) ([]byte, error)
	DeleteSecret(context.Context, string, string, int32, []byte) error
}

type LocalSecretProvider struct {
	key  []byte
	rand io.Reader
}

func NewLocalSecretProviderFromEnv() (*LocalSecretProvider, error) {
	key, err := decodeMasterKey(os.Getenv("FORMA_SECRET_MASTER_KEY"))
	if err != nil {
		return nil, entity.ErrSecretProviderUnavailable
	}
	return &LocalSecretProvider{key: key, rand: rand.Reader}, nil
}

func decodeMasterKey(value string) ([]byte, error) {
	if value == "" {
		return nil, entity.ErrSecretProviderUnavailable
	}
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if decoded, err := hex.DecodeString(value); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	return nil, entity.ErrSecretInvalid
}

func secretAAD(tenantID, credentialRefID string, keyVersion int32) []byte {
	return []byte(tenantID + "|" + credentialRefID + "|" + strconv.FormatInt(int64(keyVersion), 10))
}

func (p *LocalSecretProvider) StoreSecret(_ context.Context, tenantID, credentialRefID string, keyVersion int32, plaintext []byte) ([]byte, error) {
	if p == nil || len(p.key) != 32 || len(plaintext) == 0 {
		return nil, entity.ErrSecretInvalid
	}
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, entity.ErrSecretProviderUnavailable
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, entity.ErrSecretProviderUnavailable
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(p.rand, nonce); err != nil {
		return nil, entity.ErrSecretProviderUnavailable
	}
	// Ciphertext column envelope: nonce (12 bytes) || AES-GCM ciphertext.
	// AAD is tenant_id|credential_ref_id|key_version.
	return gcm.Seal(nonce, nonce, plaintext, secretAAD(tenantID, credentialRefID, keyVersion)), nil
}

func (p *LocalSecretProvider) ResolveSecret(_ context.Context, tenantID, credentialRefID string, keyVersion int32, encrypted []byte) ([]byte, error) {
	if p == nil || len(p.key) != 32 {
		return nil, entity.ErrSecretProviderUnavailable
	}
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, entity.ErrSecretProviderUnavailable
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil || len(encrypted) < gcm.NonceSize() {
		return nil, entity.ErrSecretInvalid
	}
	nonce, ciphertext := encrypted[:gcm.NonceSize()], encrypted[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, secretAAD(tenantID, credentialRefID, keyVersion))
	if err != nil {
		return nil, entity.ErrSecretInvalid
	}
	return plain, nil
}

func (p *LocalSecretProvider) RotateSecret(ctx context.Context, tenantID, credentialRefID string, keyVersion int32, plaintext []byte) ([]byte, error) {
	return p.StoreSecret(ctx, tenantID, credentialRefID, keyVersion, plaintext)
}

func (p *LocalSecretProvider) DeleteSecret(_ context.Context, _ string, _ string, _ int32, encrypted []byte) error {
	for i := range encrypted {
		encrypted[i] = 0
	}
	return nil
}
