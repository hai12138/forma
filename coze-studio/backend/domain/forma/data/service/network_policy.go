/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"net"
	"net/url"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

type HostResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

type OutboundNetworkPolicy interface {
	ValidateURL(context.Context, *url.URL) error
}

type DefaultOutboundNetworkPolicy struct {
	Resolver HostResolver
}

func NewDefaultOutboundNetworkPolicy(resolver HostResolver) *DefaultOutboundNetworkPolicy {
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	return &DefaultOutboundNetworkPolicy{Resolver: resolver}
}

func (p *DefaultOutboundNetworkPolicy) ValidateURL(ctx context.Context, u *url.URL) error {
	if u == nil || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return entity.ErrPublicConfigInvalid
	}
	ips, err := p.Resolver.LookupIPAddr(ctx, u.Hostname())
	if err != nil || len(ips) == 0 {
		return entity.ErrDataConnectionFailed
	}
	for _, resolved := range ips {
		if blockedOutboundIP(resolved.IP) {
			return entity.ErrPublicConfigInvalid
		}
	}
	return nil
}

func blockedOutboundIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if v4 := ip.To4(); v4 != nil {
		return v4[0] == 169 && v4[1] == 254
	}
	return ip.IsLinkLocalUnicast()
}
