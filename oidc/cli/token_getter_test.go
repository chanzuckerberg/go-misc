package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDiscoverIntrospectionEndpoint(t *testing.T) {
	r := require.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.Equal("/.well-known/openid-configuration", req.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		r.NoError(json.NewEncoder(w).Encode(map[string]string{
			"introspection_endpoint": "https://idp.example.com/oauth2/v1/introspect",
		}))
	}))
	defer srv.Close()

	ep, err := discoverIntrospectionEndpoint(context.Background(), srv.URL)
	r.NoError(err)
	r.Equal("https://idp.example.com/oauth2/v1/introspect", ep)
}

func TestDiscoverIntrospectionEndpointMissing(t *testing.T) {
	r := require.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		r.NoError(json.NewEncoder(w).Encode(map[string]string{
			"authorization_endpoint": "https://idp.example.com/oauth2/v1/authorize",
		}))
	}))
	defer srv.Close()

	_, err := discoverIntrospectionEndpoint(context.Background(), srv.URL)
	r.Error(err)
	r.Contains(err.Error(), "no introspection_endpoint")
}

func TestIntrospectTokenExpiry(t *testing.T) {
	r := require.New(t)

	expiry := time.Now().Add(7 * 24 * time.Hour).Truncate(time.Second)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.Equal(http.MethodPost, req.Method)
		r.Equal("application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
		r.NoError(req.ParseForm())
		r.Equal("my-refresh-token", req.FormValue("token"))
		r.Equal("refresh_token", req.FormValue("token_type_hint"))
		r.Equal("my-client-id", req.FormValue("client_id"))

		w.Header().Set("Content-Type", "application/json")
		r.NoError(json.NewEncoder(w).Encode(map[string]interface{}{
			"active": true,
			"exp":    expiry.Unix(),
		}))
	}))
	defer srv.Close()

	got, err := introspectTokenExpiry(context.Background(), srv.URL, "my-client-id", "my-refresh-token")
	r.NoError(err)
	r.Equal(expiry, got)
}

func TestIntrospectTokenExpiryInactive(t *testing.T) {
	r := require.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		r.NoError(json.NewEncoder(w).Encode(map[string]interface{}{
			"active": false,
		}))
	}))
	defer srv.Close()

	_, err := introspectTokenExpiry(context.Background(), srv.URL, "client", "token")
	r.Error(err)
	r.Contains(err.Error(), "no longer active")
}

func TestIntrospectTokenExpiryNoExp(t *testing.T) {
	r := require.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		r.NoError(json.NewEncoder(w).Encode(map[string]interface{}{
			"active": true,
		}))
	}))
	defer srv.Close()

	_, err := introspectTokenExpiry(context.Background(), srv.URL, "client", "token")
	r.Error(err)
	r.Contains(err.Error(), "no exp in introspection response")
}

func TestIntrospectTokenExpiryServerError(t *testing.T) {
	r := require.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := introspectTokenExpiry(context.Background(), srv.URL, "client", "token")
	r.Error(err)
	r.Contains(err.Error(), "500")
}
