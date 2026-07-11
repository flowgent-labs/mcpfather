// Standalone OIDC provider for integration testing.
// Supports client_credentials and device_code grants with configurable
// clients, real RSA key signing, JWKS endpoint, and flexible JWT claims.
// Prints its listen address to stdout so tests can discover it.
//
// Endpoints:
//   GET  /.well-known/openid-configuration  — OIDC discovery
//   GET  /keys                               — JWKS (public key)
//   POST /token                              — client_credentials or device_code → JWT
//   POST /device/code                        — initiate device authorization flow
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
	"sync"
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

	// Derive issuer from actual listen address so JWTs match what
	// discovery advertises (real OIDC providers do this too).
	issuerURL := "http://" + ln.Addr().String()

	srv := &oidcServer{
		clients:     clientMap,
		privKey:     privKey,
		issuerVal:   issuerURL,
		audience:    *audience,
		callCount:   atomic.Int64{},
		deviceCodes: make(map[string]*deviceCodeEntry),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", srv.handleDiscovery)
	mux.HandleFunc("/token", srv.handleToken)
	mux.HandleFunc("/device/code", srv.handleDeviceCode)
	mux.HandleFunc("/sign", srv.handleSign)
	mux.HandleFunc("/keys", srv.handleJWKS)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.SetOutput(os.Stderr)
	log.Printf("Test OIDC provider listening on %s (issuer=%s, audience=%s)", ln.Addr().String(), issuerURL, *audience)
	if err := http.Serve(ln, mux); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

type deviceCodeEntry struct {
	clientID  string
	scope     string
	expiresAt time.Time
}

type oidcServer struct {
	clients     map[string]string
	privKey     *rsa.PrivateKey
	issuerVal   string
	audience    string
	callCount   atomic.Int64
	deviceCodes map[string]*deviceCodeEntry
	dcMu        sync.Mutex
}

func (s *oidcServer) issuerURL(r *http.Request) string {
	return "http://" + r.Host
}

func (s *oidcServer) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	iss := s.issuerURL(r)
	disc := map[string]interface{}{
		"issuer":                        iss,
		"token_endpoint":                iss + "/token",
		"authorization_endpoint":        iss + "/auth",
		"device_authorization_endpoint": iss + "/device/code",
		"jwks_uri":                      iss + "/keys",
		"grant_types_supported": []string{
			"client_credentials",
			"urn:ietf:params:oauth:grant-type:device_code",
		},
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post"},
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

	grantType := r.FormValue("grant_type")

	switch grantType {
	case "client_credentials":
		s.handleClientCredentials(w, r)
	case "urn:ietf:params:oauth:grant-type:device_code":
		s.handleDeviceToken(w, r)
	default:
		http.Error(w, `{"error":"unsupported_grant_type"}`, http.StatusBadRequest)
	}
}

func (s *oidcServer) handleClientCredentials(w http.ResponseWriter, r *http.Request) {
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	expectedSecret, ok := s.clients[clientID]
	if !ok || clientSecret != expectedSecret {
		http.Error(w, `{"error":"invalid_client"}`, http.StatusUnauthorized)
		return
	}

	s.issueToken(w, clientID, r.FormValue("scope"))
}

func (s *oidcServer) handleDeviceToken(w http.ResponseWriter, r *http.Request) {
	deviceCode := r.FormValue("device_code")
	clientID := r.FormValue("client_id")

	s.dcMu.Lock()
	entry, ok := s.deviceCodes[deviceCode]
	if ok && entry.clientID == clientID && time.Now().Before(entry.expiresAt) {
		delete(s.deviceCodes, deviceCode)
		s.dcMu.Unlock()
		s.issueToken(w, clientID, entry.scope)
		return
	}
	s.dcMu.Unlock()

	if !ok {
		http.Error(w, `{"error":"invalid_grant","error_description":"device code not found or already used"}`, http.StatusBadRequest)
	} else {
		http.Error(w, `{"error":"authorization_pending"}`, http.StatusUnauthorized)
	}
}

func (s *oidcServer) issueToken(w http.ResponseWriter, subject, scope string) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": s.issuerVal,
		"sub": subject,
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
		"scope":        scope,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleDeviceCode initiates a device authorization flow (RFC 8628).
// In this test provider approval is automatic — the /token endpoint
// immediately returns a token when polled with the device_code.
func (s *oidcServer) handleDeviceCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}

	clientID := r.FormValue("client_id")
	if clientID == "" {
		http.Error(w, `{"error":"invalid_request","error_description":"client_id is required"}`, http.StatusBadRequest)
		return
	}

	deviceCode := randomString(32)
	userCode := randomString(8)

	s.dcMu.Lock()
	s.deviceCodes[deviceCode] = &deviceCodeEntry{
		clientID:  clientID,
		scope:     r.FormValue("scope"),
		expiresAt: time.Now().Add(300 * time.Second),
	}
	s.dcMu.Unlock()

	resp := map[string]interface{}{
		"device_code":               deviceCode,
		"user_code":                 strings.ToUpper(userCode[:4]) + "-" + strings.ToUpper(userCode[4:]),
		"verification_uri":          s.issuerVal + "/device",
		"verification_uri_complete": s.issuerVal + "/device?user_code=" + userCode,
		"expires_in":                300,
		"interval":                  1,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		bi, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[bi.Int64()]
	}
	return string(b)
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
