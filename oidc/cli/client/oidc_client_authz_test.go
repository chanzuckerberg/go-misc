package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const startPort = 49152
const endPort = 49152 + 63

func TestNewOIDCClientShouldBindFirstAvailablePort(t *testing.T) {
	listeners, err := bindPortsInRange(startPort, startPort)
	if err != nil {
		t.Fatalf("Failed to start test: cannot bind ports in range %d-%d: %v", startPort, startPort, err)
	}
	defer func() {
		for _, l := range listeners {
			l.Close()
		}
	}()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"issuer":"http://`+r.Host+`","authorization_endpoint":"...","token_endpoint":"..."}`)
	}))
	defer testServer.Close()

	ctx := context.Background()

	client, err := NewOIDCClient(ctx, "test-client", testServer.URL)
	if err != nil {
		t.Fatalf("Expected bind to find an available port, but failed: %v", err)
	}

	portStr := strings.Split(client.RedirectURL, "http://localhost:")[1]
	var boundPort int
	fmt.Sscanf(portStr, "%d", &boundPort)
	checkIfBoundPortInRange(t, boundPort)
}

func TestNewOIDCClientPortConflictErrorIsPropagated(t *testing.T) {
	listeners, err := bindPortsInRange(startPort, endPort)
	if err != nil {
		t.Fatalf("Failed to start test: cannot bind ports in range %d-%d: %v", startPort, endPort, err)
	}
	defer func() {
		for _, l := range listeners {
			l.Close()
		}
	}()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"issuer":"http://`+r.Host+`","authorization_endpoint":"...","token_endpoint":"..."}`)
	}))
	defer testServer.Close()

	ctx := context.Background()

	_, err = NewOIDCClient(ctx, "test-client", testServer.URL)
	if err == nil {
		t.Fatal("Expected error due to port conflict, but got nil")
	}

	expectedMsg := fmt.Sprintf("creating default authenticator: could not open a port: binding to any port in range %d-%d", startPort, endPort)
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error was not propagated correctly. Got: %v", err)
	}
}

func TestServerBindPortConflictsShouldThrowError(t *testing.T) {
	listeners, err := bindPortsInRange(startPort, endPort)
	if err != nil {
		t.Fatalf("Failed to start test: cannot bind ports in range %d-%d: %v", startPort, endPort, err)
	}
	defer func() {
		for _, l := range listeners {
			l.Close()
		}
	}()

	config := &ServerConfig{
		FromPort: startPort,
		ToPort:   endPort,
	}

	s := &server{config: config}
	err = s.bind(config)

	if err == nil {
		t.Fatal("Expected bind error when all ports are occupied, but got nil")
	}

	expectedMsg := fmt.Sprintf("binding to any port in range %d-%d", startPort, endPort)
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain %q, but got %q", expectedMsg, err.Error())
	}
}

func TestServerBindShouldFindsAvailablePort(t *testing.T) {
	listeners, err := bindPortsInRange(startPort, startPort)
	if err != nil {
		t.Fatalf("Failed to start test: cannot bind ports in range %d-%d: %v", startPort, startPort, err)
	}
	defer func() {
		for _, l := range listeners {
			l.Close()
		}
	}()

	config := &ServerConfig{
		FromPort: startPort,
		ToPort:   endPort,
	}

	s := &server{config: config}
	err = s.bind(config)

	if err != nil {
		t.Fatalf("Expected bind to find an available port, but failed: %v", err)
	}
	checkIfBoundPortInRange(t, s.port)
}

func bindPortsInRange(start, end int) ([]net.Listener, error) {
	var listeners []net.Listener
	for p := start; p <= end; p++ {
		l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", p))
		if err != nil {
			continue
		}
		listeners = append(listeners, l)
	}

	if len(listeners) == 0 {
		return nil, fmt.Errorf("no ports could be bound in range %d-%d", start, end)
	}

	return listeners, nil
}

func checkIfBoundPortInRange(t *testing.T, boundPort int) {
	if boundPort < startPort || boundPort > startPort+64 {
		t.Errorf("Bound port %d is outside of allowed range [%d-%d]", boundPort, startPort, startPort+64)
	}
	if boundPort == startPort {
		t.Errorf("Client bound to port %d, which should have been occupied by the test", boundPort)
	}
}
