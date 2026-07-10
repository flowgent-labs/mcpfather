package tests

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

// mockLDAPServer is a minimal LDAP server that handles simple bind requests.
// It listens on a random TCP port and authenticates against a fixed set of
// credentials (dn + password).  Only LDAPv3 simple bind is supported — enough
// for the client_credentials-style service-account bind used by the generated
// MCP server.
type mockLDAPServer struct {
	listener net.Listener
	url      string // ldap://host:port

	mu           sync.Mutex
	bindCallCount int
	// validBindDN and validPassword are the only credentials the server accepts.
	validBindDN  string
	validPassword string
}

// startMockLDAPServer creates and starts the mock LDAP server.
// The server will accept binds for validDN/validPassword only.
func startMockLDAPServer(validDN, validPassword string) *mockLDAPServer {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("mockLDAPServer: listen: %v", err))
	}

	s := &mockLDAPServer{
		listener:      l,
		url:           fmt.Sprintf("ldap://%s", l.Addr().String()),
		validBindDN:   validDN,
		validPassword: validPassword,
	}
	go s.serve()
	return s
}

// Close shuts down the mock server.
func (s *mockLDAPServer) Close() {
	s.listener.Close()
}

// URL returns the ldap:// address of the server.
func (s *mockLDAPServer) URL() string { return s.url }

// BindCallCount returns how many binds were attempted.
func (s *mockLDAPServer) BindCallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.bindCallCount
}

// ExpectedBasicAuth returns the Basic auth header that the generated server
// should produce from a successful bind with this mock.
func (s *mockLDAPServer) ExpectedBasicAuth() string {
	return "Basic " + base64.StdEncoding.EncodeToString(
		[]byte(s.validBindDN+":"+s.validPassword))
}

func (s *mockLDAPServer) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

func (s *mockLDAPServer) handleConn(conn net.Conn) {
	defer conn.Close()

	// Read the full LDAP message (assume it fits in one TCP segment for the
	// simple-bind case — this is a test mock, not production code).
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}
	buf = buf[:n]

	s.mu.Lock()
	s.bindCallCount++
	s.mu.Unlock()

	// Extract the password from the simple-authentication field (tag 0x80).
	// BER structure: ... 80 <len> <password_bytes>
	dn, password := extractBindCredentials(buf)

	var resp []byte
	if dn == s.validBindDN && password == s.validPassword {
		resp = buildBindResponse(buf, 0) // success
	} else {
		resp = buildBindResponse(buf, 49) // invalidCredentials
	}
	conn.Write(resp)
}

// extractBindCredentials does minimal BER parsing to recover the DN and
// password from a simple-bind LDAP BindRequest.
func extractBindCredentials(data []byte) (dn, password string) {
	// Skip outer SEQUENCE (tag 30) and messageID (tag 02)
	// Look for BindRequest tag 60 (APPLICATION 0).
	pos := 2                   // skip 30 <len>
	pos += int(data[1])        // outer SEQUENCE body (for small packets)
	if pos < 2 || pos >= len(data) {
		pos = 2
	}
	// The bind request starts at tag 60.
	for pos < len(data) {
		if data[pos] == 0x60 {
			pos++ // tag
			if pos >= len(data) { break }
			blen, n := readBERLen(data[pos:])
			pos += n
			_ = blen
			break
		}
		pos++
	}
	if pos >= len(data) {
		return "", ""
	}

	// Now we're inside the BindRequest SEQUENCE.
	// Fields: 02 01 <version>, 04 <dnlen> <dn>, 80 <pwlen> <password>
	for pos < len(data) {
		tag := data[pos]
		pos++
		if pos >= len(data) { break }
		flen, n := readBERLen(data[pos:])
		pos += n

		switch {
		case tag == 0x04:
			// DN (first OCTET STRING in the sequence)
			if dn == "" && pos+flen <= len(data) {
				dn = string(data[pos : pos+flen])
			}
		case tag == 0x80:
			if pos+flen <= len(data) {
				password = string(data[pos : pos+flen])
				return dn, password
			}
		}
		pos += flen
	}
	return dn, password
}

// readBERLen reads a BER length field. Returns (length, bytes_consumed).
func readBERLen(data []byte) (int, int) {
	if len(data) == 0 {
		return 0, 0
	}
	if data[0]&0x80 == 0 {
		return int(data[0]), 1
	}
	numBytes := int(data[0] & 0x7f)
	if numBytes == 0 || numBytes > 4 || 1+numBytes > len(data) {
		return 0, 1
	}
	val := uint32(0)
	for i := 0; i < numBytes; i++ {
		val = (val << 8) | uint32(data[1+i])
	}
	return int(val), 1 + numBytes
}

// buildBindResponse constructs a BindResponse BER message with the given
// resultCode (0 = success, 49 = invalidCredentials).  It copies the messageID
// from the request so the client can correlate the response.
func buildBindResponse(req []byte, resultCode int) []byte {
	// Extract messageID from the request (tag 02, first INTEGER).
	msgID := 1 // default
	if len(req) > 4 && req[2] == 0x02 {
		mlen := int(req[3])
		if mlen <= 4 && 4+mlen <= len(req) {
			idBytes := req[4 : 4+mlen]
			// Pad to 4 bytes for binary.Read
			for len(idBytes) < 4 {
				idBytes = append([]byte{0}, idBytes...)
			}
			msgID = int(binary.BigEndian.Uint32(idBytes))
		}
	}

	// BER-encode: BindResponse
	// 0a 01 <resultCode>
	// 04 00              (matchedDN = "")
	// 04 00              (diagnosticMessage = "")
	msg := ""
	if resultCode != 0 {
		msg = "invalid credentials"
	}
	inner := []byte{0x0a, 0x01, byte(resultCode),
		0x04, 0x00,
		0x04, byte(len(msg))}
	inner = append(inner, []byte(msg)...)

	// Wrap in APPLICATION 1 (tag 61)
	app1 := []byte{0x61, byte(len(inner))}
	app1 = append(app1, inner...)

	// Encode messageID
	idEnc := []byte{0x02, 0x01, byte(msgID)}

	// Outer SEQUENCE
	body := append(idEnc, app1...)
	out := []byte{0x30, byte(len(body))}
	out = append(out, body...)
	return out
}
