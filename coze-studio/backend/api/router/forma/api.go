/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	formahandler "github.com/coze-dev/coze-studio/backend/api/handler/forma"
)

func Register(r *server.Hertz) {
	root := r.Group("/api")
	v1 := root.Group("/forma/v1")
	{
		v1.GET("/health", formahandler.Health)
		v1.GET("/version", formahandler.Version)
		v1.GET("/meta/baseline", formahandler.Baseline)
	}
}
