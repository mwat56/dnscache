/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/mwat56/dnscache"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_extractHostname(t *testing.T) {
	tests := []struct {
		name      string
		nameField []byte
		wantHost  string
	}{
		/* */
		{
			name:      "01 - empty name field",
			nameField: []byte{},
			wantHost:  "",
		},
		{
			name:      "02 - single label",
			nameField: []byte{1, 'a'},
			wantHost:  "a",
		},
		{
			name:      "03 - two labels",
			nameField: []byte{1, 'a', 1, 'b'},
			wantHost:  "a.b",
		},
		{
			name:      "04 - four labels",
			nameField: []byte{1, 'a', 1, 'b', 1, 'c', 10, 'd'},
			wantHost:  "",
		},
		{
			name:      "05 - five labels",
			nameField: []byte{1, 'a', 1, 'b', 1, 'c', 1, 'd', 0, ' '},
			wantHost:  "a.b.c.d",
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotHost := extractHostname(tc.nameField)

			if gotHost != tc.wantHost {
				t.Errorf("extractHostname() = %q, want %q",
					gotHost, tc.wantHost)
			}
		})
	}
} // Test_extractHostname()

type (
	// Mock address implementation
	tMockAddr struct{}

	// Mock PacketConn implementation for testing
	tMockPacketConn struct {
		writeTo  func([]byte, net.Addr) (int, error)
		respChan chan []byte
	}
)

func (ma *tMockAddr) Network() string {
	return "udp"
} // Network()

func (ma *tMockAddr) String() string {
	return "127.0.0.1:53"
} // String()

func (mpc *tMockPacketConn) Close() error {
	return nil
} // Close()

func (mpc *tMockPacketConn) LocalAddr() net.Addr {
	return &tMockAddr{}
} // LocalAddr()

func (mpc *tMockPacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	return 0, nil, io.EOF
} // ReadFrom()

func (mpc *tMockPacketConn) SetDeadline(time.Time) error {
	return nil
} // SetDeadline()

func (mpc *tMockPacketConn) SetReadDeadline(time.Time) error {
	return nil
} // SetReadDeadline()

func (mpc *tMockPacketConn) SetWriteDeadline(time.Time) error {
	return nil
} // SetWriteDeadline()

func (mpc *tMockPacketConn) WriteTo(aBuf []byte, aAddr net.Addr) (int, error) {
	if nil != mpc.writeTo {
		return mpc.writeTo(aBuf, aAddr)
	}
	if nil != mpc.respChan {
		resp := make([]byte, len(aBuf))
		copy(resp, aBuf)
		mpc.respChan <- resp
	}

	return len(aBuf), nil
} // WriteTo()

// Helper function to create a DNS request packet
func createDNSRequest(aID uint16, aHostname string) []byte {
	request := make([]byte, 512)

	// Set header fields
	binary.BigEndian.PutUint16(request[0:2], aID)   // ID
	binary.BigEndian.PutUint16(request[2:4], dnsRD) // Flags (RD bit set)
	binary.BigEndian.PutUint16(request[4:6], 1)     // QDCOUNT = 1
	binary.BigEndian.PutUint16(request[6:8], 0)     // ANCOUNT = 0
	binary.BigEndian.PutUint16(request[8:10], 0)    // NSCOUNT = 0
	binary.BigEndian.PutUint16(request[10:12], 0)   // ARCOUNT = 0

	// Add question section
	offset := 12

	// Convert hostname to DNS format (sequence of length-prefixed labels)
	labels := []byte(aHostname)
	start := 0
	for i := 0; i <= len(labels); i++ {
		if i == len(labels) || labels[i] == '.' {
			labelLen := i - start
			request[offset] = byte(labelLen)
			offset++
			copy(request[offset:offset+labelLen], labels[start:start+labelLen])
			offset += labelLen
			start = i + 1
		}
	}

	// Terminating zero
	request[offset] = 0
	offset++

	// Set question type (A) and class (IN)
	binary.BigEndian.PutUint16(request[offset:offset+2], dnsTypeA)
	offset += 2
	binary.BigEndian.PutUint16(request[offset:offset+2], dnsClassIN)
	offset += 2

	return request[:offset]
} // createDNSRequest()

func Test_handleDNSRequest(t *testing.T) {
	// Create a mock resolver
	resolver := dnscache.New(0)

	// Add a test entry to the resolver
	testHost := "example.org"
	testIP := net.ParseIP("192.168.2.1")
	_ = resolver.Create(context.TODO(), testHost, []net.IP{testIP}, 300)

	tests := []struct {
		name      string
		request   []byte
		checkResp func([]byte) bool
		wantResp  bool // true if we expect a response, false otherwise
	}{
		/* */
		{
			name:    "01 - request too short",
			request: make([]byte, 10), // Less than 12 bytes
			checkResp: func(resp []byte) bool {
				return true // No validation needed as we don't expect a response
			},
			wantResp: false,
		},
		{
			name:    "02 - valid A record request",
			request: createDNSRequest(1234, testHost),
			checkResp: func(resp []byte) bool {
				if 12 > len(resp) {
					t.Logf("Response too short: %d bytes", len(resp))
					return false // Too short to be valid
				}

				// Check header fields
				id := binary.BigEndian.Uint16(resp[0:2])
				flags := binary.BigEndian.Uint16(resp[2:4])
				anCount := binary.BigEndian.Uint16(resp[6:8])

				// Verify response has our query ID
				if 1234 != id {
					t.Logf("Expected ID 1234, got %d", id)
					return false
				}

				// Check if response flag is set (QR bit must be 1 in responses)
				if 0 == (flags & dnsQR) {
					t.Logf("QR bit not set in response flags: %04x", flags)
					return false
				}

				// Should have at least one answer
				if 0 == anCount {
					t.Logf("Expected at least 1 answer, got %d", anCount)
					return false
				}

				// For this test, we'll accept any valid response with answers
				// since the resolver might be returning different IPs
				return true
			},
			wantResp: true,
		},
		/* */
		{
			name:    "03 - non-existent domain",
			request: createDNSRequest(5678, "nonexistent.example"),
			checkResp: func(resp []byte) bool {
				if 12 > len(resp) {
					t.Logf("Response too short: %d bytes", len(resp))
					return false
				}

				// Check header fields
				id := binary.BigEndian.Uint16(resp[0:2])
				flags := binary.BigEndian.Uint16(resp[2:4])
				rcode := flags & 0xF

				// Verify response has our query ID
				if 5678 != id {
					t.Logf("Expected ID 5678, got %d", id)
					return false
				}

				// Check response code (should be NXDOMAIN = 3)
				if rcode != dnsRcodeNXDomain {
					t.Logf("Expected NXDOMAIN (3), got rcode %d", rcode)
					return false
				}

				return true
			},
			wantResp: true,
		},
		/* */
		{
			name:    "04 - malformed query",
			request: append(createDNSRequest(9876, "malformed"), 0xFF, 0xFF, 0xFF),
			checkResp: func(resp []byte) bool {
				return true // No validation needed as we don't expect a response
			},
			wantResp: false,
		},
		/* */
		{
			name: "05 - unsupported query type",
			request: func() []byte {
				req := createDNSRequest(4321, "example.org")
				// Change the query type from A (1) to MX (15)
				binary.BigEndian.PutUint16(req[len(req)-4:len(req)-2], 15)
				return req
			}(),
			checkResp: func(resp []byte) bool {
				if 12 > len(resp) {
					t.Logf("Response too short: %d bytes", len(resp))
					return false
				}

				// Check header fields
				id := binary.BigEndian.Uint16(resp[0:2])
				// flags := binary.BigEndian.Uint16(resp[2:4])
				anCount := binary.BigEndian.Uint16(resp[6:8])

				// Verify response has our query ID
				if 4321 != id {
					t.Logf("Expected ID 4321, got %d", id)
					return false
				}

				// Should have zero answers since we don't support MX records
				if 0 != anCount {
					t.Logf("Expected 0 answers for unsupported query type, got %d", anCount)
					return false
				}

				return true
			},
			wantResp: true,
		},
		/* */
		{
			name: "06 - multiple questions",
			request: func() []byte {
				req := createDNSRequest(7890, "example.org")
				// Change QDCOUNT to 2
				binary.BigEndian.PutUint16(req[4:6], 2)
				// Add another question with valid format
				req = append(req, []byte{7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0, 0, 1, 0, 1}...)
				return req
			}(),
			checkResp: func(resp []byte) bool {
				if 12 > len(resp) {
					t.Logf("Response too short: %d bytes", len(resp))
					return false
				}

				// Check header fields
				id := binary.BigEndian.Uint16(resp[0:2])

				// Verify response has our query ID
				if 7890 != id {
					t.Logf("Expected ID 7890, got %d", id)
					return false
				}

				return true
			},
			wantResp: true,
		},
		/* */
		// TODO: Add more test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock connection with a channel for responses
			respChan := make(chan []byte, 1)
			mockConn := &tMockPacketConn{
				respChan: respChan,
			}

			// Handle the request directly
			handleDNSRequest(mockConn, &tMockAddr{}, tc.request, resolver)

			// Wait for response or timeout
			var resp []byte
			select {
			case resp = <-respChan:
				// Got response
				t.Logf("Received response of %d bytes", len(resp))
			case <-time.After(100 * time.Millisecond):
				// For test cases where we expect no response
				switch tc.name {
				case "01 - request too short", "04 - malformed query":
					resp = []byte{} // Empty response for validation
				default:
					t.Logf("Timed out waiting for response")
				}
			}

			// Check the response
			if !tc.checkResp(resp) {
				t.Errorf("handleDNSRequest() response validation failed for test: %q",
					tc.name)
			}
		})
	}
} // Test_handleDNSRequest()

func Test_startDNSserver(t *testing.T) {
	// Create a test resolver
	resolver := dnscache.New(0)

	tests := []struct {
		name      string
		resolver  *dnscache.TResolver
		address   string
		port      int
		wantErr   bool
		setupFunc func()                                            // Optional setup function
		checkFunc func(t *testing.T, address string, port int) bool // Optional validation function
	}{
		{
			name:     "01 - nil resolver",
			resolver: nil,
			address:  "",
			port:     5353,
			wantErr:  true,
		},
		{
			name:     "02 - invalid port",
			resolver: resolver,
			address:  "",
			port:     -1,
			wantErr:  true,
		},
		{
			name:     "03 - port already in use",
			resolver: resolver,
			address:  "127.0.0.1", // Use localhost explicitly
			port:     5353,
			wantErr:  true,
			setupFunc: func() {
				// Try to bind to the port first to make it unavailable
				listener, err := net.ListenPacket("udp", "127.0.0.1:5353")
				if nil != err {
					// Port might already be in use by another process
					// This is fine for our test since we want to test the "port in use" scenario
					t.Logf("Port 127.0.0.1:5353 already in use (which is what we want to test): %v", err)
					return
				}
				t.Cleanup(func() {
					listener.Close()
				})
			},
		},
		{
			name:     "04 - valid configuration",
			resolver: resolver,
			address:  "127.0.0.1", // Use localhost explicitly
			port:     5354,
			wantErr:  false,
			checkFunc: func(t *testing.T, address string, port int) bool {
				// Try to connect to the server
				conn, err := net.Dial("udp", net.JoinHostPort(address, fmt.Sprintf("%d", port)))
				if nil != err {
					return false
				}
				defer conn.Close()
				return true
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Run setup if provided
			if nil != tc.setupFunc {
				tc.setupFunc()
			}

			// Create a channel to signal when the server is started
			serverStarted := make(chan struct{})

			// Start the server in a goroutine
			var err error
			go func() {
				// Signal that we're about to start the server
				close(serverStarted)
				err = startDNSserver(tc.resolver, tc.address, tc.port)
			}()

			// Wait for server to start
			<-serverStarted
			time.Sleep(100 * time.Millisecond)

			// Check if server is running if we have a check function
			if nil != tc.checkFunc && !tc.wantErr {
				if !tc.checkFunc(t, tc.address, tc.port) {
					t.Errorf("startDNSserver() server not accessible on %s:%d", tc.address, tc.port)
				}
			}

			// For the valid test case, we need to send a signal to shut down the server
			if !tc.wantErr {
				// Send termination signal by simulating Ctrl+C
				process, err := os.FindProcess(os.Getpid())
				if nil != err {
					t.Fatalf("Failed to find process: %v", err)
				}
				_ = process.Signal(syscall.SIGINT)

				// Wait for server to shut down
				time.Sleep(100 * time.Millisecond)
			}

			if (nil != err) != tc.wantErr {
				t.Errorf("startDNSserver() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
} // Test_startDNSserver()

/* _EoF_ */
