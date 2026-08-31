/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"github.com/coze-dev/coze-studio/backend/crossdomain/forma/integration"
)

type Forma interface {
	Integration() integration.FormaCozeIntegration
}

var defaultSVC Forma

func DefaultSVC() Forma {
	return defaultSVC
}

func SetDefaultSVC(s Forma) {
	defaultSVC = s
}
