/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package impl

import (
	crossforma "github.com/coze-dev/coze-studio/backend/crossdomain/forma"
	"github.com/coze-dev/coze-studio/backend/crossdomain/forma/integration"
)

type impl struct {
	integration integration.FormaCozeIntegration
}

func InitDomainService(integrationSvc integration.FormaCozeIntegration) crossforma.Forma {
	svc := &impl{integration: integrationSvc}
	crossforma.SetDefaultSVC(svc)
	return svc
}

func (i *impl) Integration() integration.FormaCozeIntegration {
	return i.integration
}
