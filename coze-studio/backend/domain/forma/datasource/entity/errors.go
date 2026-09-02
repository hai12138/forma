/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import "errors"

var (
	ErrDataSourceNotFound         = errors.New("data source not found")
	ErrDataConnectionNotFound     = errors.New("data connection not found")
	ErrDataCredentialNotFound     = errors.New("data credential not found")
	ErrDataAdapterNotSupported    = errors.New("data adapter not supported")
	ErrDataConnectionFailed       = errors.New("data connection failed")
	ErrDataDiscoveryFailed        = errors.New("data discovery failed")
	ErrDataAssetNotFound          = errors.New("data asset not found")
	ErrDataSchemaSnapshotNotFound = errors.New("data schema snapshot not found")
	ErrSecretProviderUnavailable  = errors.New("data secret provider unavailable")
	ErrSecretInvalid              = errors.New("data secret invalid")
	ErrPublicConfigInvalid        = errors.New("data public config invalid")
)
