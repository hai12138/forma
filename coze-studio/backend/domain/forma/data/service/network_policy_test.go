package service

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

type fixedResolver map[string][]net.IPAddr

func (r fixedResolver) LookupIPAddr(_ context.Context, host string) ([]net.IPAddr, error) {
	return r[host], nil
}

func TestDefaultOutboundNetworkPolicyBlocksLinkLocalAndAllowsPrivate(t *testing.T) {
	policy := NewDefaultOutboundNetworkPolicy(fixedResolver{
		"metadata": {{IP: net.ParseIP("169.254.169.254")}},
		"ipv6":     {{IP: net.ParseIP("fe80::1")}},
		"internal": {{IP: net.ParseIP("10.1.2.3")}},
	})
	for _, raw := range []string{"http://metadata/latest", "http://ipv6/"} {
		target, _ := url.Parse(raw)
		require.ErrorIs(t, policy.ValidateURL(context.Background(), target), entity.ErrPublicConfigInvalid)
	}
	target, _ := url.Parse("http://internal/")
	require.NoError(t, policy.ValidateURL(context.Background(), target))
}

func TestHTTPAdapterRejectsRedirectToLinkLocal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://169.254.169.254/latest/meta-data", http.StatusFound)
	}))
	defer server.Close()
	adapter := NewHTTPAdapter(nil)
	err := adapter.TestConnection(context.Background(), &AdapterRequest{PublicConfigJSON: `{"base_url":"` + server.URL + `"}`})
	require.ErrorIs(t, err, entity.ErrDataConnectionFailed)
}
