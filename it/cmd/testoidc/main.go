// Standalone OIDC provider for integration testing.
// Supports client_credentials grant with configurable clients.
// Prints its listen address to stdout so tests can discover it.
package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

var (
	port    = flag.Int("port", 0, "listen port (0 = random)")
	clients = flag.String("clients", "mcpgen-client:mcpgen-secret", "comma-separated client_id:client_secret pairs")
)

func main() {
	flag.Parse()

	clientMap := make(map[string]string)
	for _, pair := range strings.Split(*clients, ",") {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			clientMap[parts[0]] = parts[1]
		}
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	// Print address to stdout for test discovery
	fmt.Println(ln.Addr().String())

	srv := &oidcServer{
		clients:  clientMap,
		callCount: atomic.Int64{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", srv.handleDiscovery)
	mux.HandleFunc("/token", srv.handleToken)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.SetOutput(os.Stderr)
	log.Printf("Test OIDC provider listening on %s", ln.Addr().String())
	if err := http.Serve(ln, mux); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

type oidcServer struct {
	clients   map[string]string
	callCount atomic.Int64
}

func (s *oidcServer) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	issuer := "http://" + host
	disc := map[string]interface{}{
		"issuer":                 issuer,
		"token_endpoint":         issuer + "/token",
		"authorization_endpoint": issuer + "/auth",
		"jwks_uri":               issuer + "/keys",
		"grant_types_supported":  []string{"client_credentials"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(disc)
}

func (s *oidcServer) handleToken(w http.ResponseWriter, r *http.Request) {
	s.callCount.Add(1)

	if err := r.ParseForm(); err != nil {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}

	if r.FormValue("grant_type") != "client_credentials" {
		http.Error(w, `{"error":"unsupported_grant_type"}`, http.StatusBadRequest)
		return
	}

	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	expectedSecret, ok := s.clients[clientID]
	if !ok || clientSecret != expectedSecret {
		http.Error(w, `{"error":"invalid_client"}`, http.StatusUnauthorized)
		return
	}

	accessToken := makeJWT(clientID)

	resp := map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        r.FormValue("scope"),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func makeJWT(clientID string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))

	payload := map[string]interface{}{
		"iss": "test-oidc",
		"sub": clientID,
		"aud": "mcpgen",
		"iat": 0,
		"exp": 9999999999,
	}
	payloadJSON, _ := json.Marshal(payload)
	payloadEnc := base64.RawURLEncoding.EncodeToString(payloadJSON)

	sig := make([]byte, 32)
	rand.Read(sig)
	sigEnc := base64.RawURLEncoding.EncodeToString(sig)

	return fmt.Sprintf("%s.%s.%s", header, payloadEnc, sigEnc)
}
