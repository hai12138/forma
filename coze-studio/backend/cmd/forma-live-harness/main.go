/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 *
 * Minimal live harness: real Coze passport + SessionAuthMW + Forma tenancy APIs.
 */

package main

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/api/handler/coze"
	"github.com/coze-dev/coze-studio/backend/api/middleware"
	formaRouter "github.com/coze-dev/coze-studio/backend/api/router/forma"
	formaApp "github.com/coze-dev/coze-studio/backend/application/forma"
	userApp "github.com/coze-dev/coze-studio/backend/application/user"
	bizconfig "github.com/coze-dev/coze-studio/backend/bizpkg/config"
	crossforma "github.com/coze-dev/coze-studio/backend/crossdomain/forma"
	formaImpl "github.com/coze-dev/coze-studio/backend/crossdomain/forma/impl"
	formaIntegration "github.com/coze-dev/coze-studio/backend/crossdomain/forma/integration"
	crossuser "github.com/coze-dev/coze-studio/backend/crossdomain/user"
	userImpl "github.com/coze-dev/coze-studio/backend/crossdomain/user/impl"
	redisimpl "github.com/coze-dev/coze-studio/backend/infra/cache/impl/redis"
	"github.com/coze-dev/coze-studio/backend/infra/idgen/impl/idgen"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
)

type memStorage struct{}

func (m *memStorage) PutObject(ctx context.Context, objectKey string, content []byte, opts ...storage.PutOptFn) error {
	return nil
}
func (m *memStorage) PutObjectWithReader(ctx context.Context, objectKey string, content io.Reader, opts ...storage.PutOptFn) error {
	return nil
}
func (m *memStorage) GetObject(ctx context.Context, objectKey string) ([]byte, error) {
	return nil, storage.ErrObjectNotFound
}
func (m *memStorage) DeleteObject(ctx context.Context, objectKey string) error { return nil }
func (m *memStorage) GetObjectUrl(ctx context.Context, objectKey string, opts ...storage.GetOptFn) (string, error) {
	return "", nil
}
func (m *memStorage) HeadObject(ctx context.Context, objectKey string, opts ...storage.GetOptFn) (*storage.FileInfo, error) {
	return nil, storage.ErrObjectNotFound
}
func (m *memStorage) ListAllObjects(ctx context.Context, prefix string, opts ...storage.GetOptFn) ([]*storage.FileInfo, error) {
	return nil, nil
}
func (m *memStorage) ListObjectsPaginated(ctx context.Context, input *storage.ListObjectsPaginatedInput, opts ...storage.GetOptFn) (*storage.ListObjectsPaginatedOutput, error) {
	return &storage.ListObjectsPaginatedOutput{}, nil
}

func main() {
	ctx := context.Background()
	dsn := envOr("MYSQL_DSN", "coze:coze123@tcp(forma-live-mysql:3306)/opencoze?charset=utf8mb4&parseTime=True")
	redisAddr := envOr("REDIS_ADDR", "forma-live-redis:6379")
	addr := envOr("LISTEN_ADDR", ":8888")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logs.Fatalf("mysql: %v", err)
	}
	rdb := redisimpl.NewWithAddrAndPassword(redisAddr, "")
	idGen, err := idgen.New(rdb)
	if err != nil {
		logs.Fatalf("idgen: %v", err)
	}

	oss := &memStorage{}
	if err := bizconfig.Init(ctx, db, oss); err != nil {
		logs.Fatalf("bizconfig init: %v", err)
	}
	userSVC := userApp.InitService(ctx, db, oss, idGen)
	crossuser.SetDefaultSVC(userImpl.InitDomainService(userSVC.DomainSVC))

	formaApp.InitService(ctx, &formaApp.ServiceComponents{DB: db, IDGen: idGen})
	crossforma.SetDefaultSVC(formaImpl.InitDomainService(
		formaIntegration.NewFormaCozeIntegration(
			formaIntegration.NewCozeAgentAdapter(),
			formaIntegration.NewCozeSpaceAdapter(),
		),
	))

	h := server.Default(server.WithHostPorts(addr), server.WithExitWaitTime(time.Second))
	h.Use(middleware.ContextCacheMW())
	h.Use(func(c context.Context, ctx *app.RequestContext) {
		ctx.Set(middleware.RequestAuthTypeStr, int32(middleware.RequestAuthTypeWebAPI))
		ctx.Next(c)
	})
	h.Use(middleware.SessionAuthMW())

	h.POST("/api/passport/web/email/register/v2/", coze.PassportWebEmailRegisterV2Post)
	h.POST("/api/passport/web/email/login/", coze.PassportWebEmailLoginPost)
	h.GET("/api/passport/web/logout/", coze.PassportWebLogoutGet)

	formaRouter.Register(h)

	logs.Infof("forma live harness listening on %s", addr)
	h.Spin()
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
