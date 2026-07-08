// Standalone OIDC provider for integration testing.
// Supports client_credentials grant with configurable clients,
// real RSA key signing, JWKS endpoint, and flexible JWT claims.
// Prints its listen address to stdout so tests can discover it.
//
// Endpoints:
//   GET  /.well-known/openid-configuration  — OIDC discovery with jwks_uri
//   GET  /keys                               — JWKS (public key)
//   POST /token                              — client_credentials → signed JWT
//   POST /sign                               — custom claims → signed JWT
//   GET  /health                             — readiness probe
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	port     = flag.Int("port", 0, "listen port (0 = random)")
	clients  = flag.String("clients", "mcpfather-client:mcpfather-secret", "comma-separated client_id:client_secret pairs")
	issuer   = flag.String("issuer", "test-oidc", "issuer claim in signed JWTs")
	audience = flag.String("audience", "mcpfather", "audience claim in signed JWTs")
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

	// Generate RSA 2048-bit keypair for real JWT signing.
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("generate RSA key: %v", err)
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	// Print address to stdout for test discovery
	fmt.Println(ln.Addr().String())

	srv := &oidcServer{
		clients:   clientMap,
		privKey:   privKey,
		issuerVal: *issuer,
		audience:  *audience,
		callCount: atomic.Int64{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", srv.handleDiscovery)
	mux.HandleFunc("/token", srv.handleToken)
	mux.HandleFunc("/sign", srv.handleSign)
	mux.HandleFunc("/keys", srv.handleJWKS)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.SetOutput(os.Stderr)
	log.Printf("Test OIDC provider listening on %s (issuer=%s, audience=%s)", ln.Addr().String(), *issuer, *audience)
	if err := http.Serve(ln, mux); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

type oidcServer struct {
	clients   map[string]string
	privKey   *rsa.PrivateKey
	issuerVal string
	audience  string
	callCount atomic.Int64
}

func (s *oidcServer) issuerURL(r *http.Request) string {
	return "http://" + r.Host
}

func (s *oidcServer) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	iss := s.issuerURL(r)
	disc := map[string]interface{}{
		"issuer":                 iss,
		"token_endpoint":         iss + "/token",
		"authorization_endpoint": iss + "/auth",
		"jwks_uri":               iss + "/keys",
		"grant_types_supported":  []string{"client_credentials"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(disc)
}

func (s *oidcServer) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		return
	}
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

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": s.issuerVal,
		"sub": clientID,
		"aud": s.audience,
		"iat": now.Unix(),
		"exp": now.Add(3600 * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-1"
	signed, err := token.SignedString(s.privKey)
	if err != nil {
		http.Error(w, `{"error":"token_signing_failed"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"access_token": signed,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        r.FormValue("scope"),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleSign accepts a JSON body of JWT claims and returns a signed JWT.
// This allows tests to craft tokens with specific aud, iss, exp, iat, etc.
func (s *oidcServer) handleSign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var claims jwt.MapClaims
	if err := json.NewDecoder(r.Body).Decode(&claims); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid_claims_json: %v"}`, err), http.StatusBadRequest)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-1"
	signed, err := token.SignedString(s.privKey)
	if err != nil {
		http.Error(w, `{"error":"token_signing_failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"access_token": signed})
}

// handleJWKS serves the RSA public key as a JWKS document (RFC 7517).
// The keyfunc/v3 library used by the generated MCP server consumes this format.
func (s *oidcServer) handleJWKS(w http.ResponseWriter, r *http.Request) {
	pub := &s.privKey.PublicKey

	// Encode modulus (N) as base64url.
	nBytes := pub.N.Bytes()
	n := base64.RawURLEncoding.EncodeToString(nBytes)

	// Encode exponent (E). Standard E=65537 → "AQAB" in base64url.
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes())

	keys := map[string]interface{}{
		"keys": []map[string]string{
			{
				"kty": "RSA",
				"kid": "test-key-1",
				"alg": "RS256",
				"use": "sig",
				"n":   n,
				"e":   e,
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}
