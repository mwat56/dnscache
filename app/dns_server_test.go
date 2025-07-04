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
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/mwat56/dnscache"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

// Helper types and functions for testing DNS forwarding

type (
	tMockForwarder struct {
		responses map[string][]byte
	}

	tMockForwarderClient struct {
		mockForwarder *tMockForwarder
		forwardCalled bool
	}
)

func (c *tMockForwarderClient) ForwardDNSRequest(ctx context.Context, forwarder string, request []byte) ([]byte, error) {
	c.forwardCalled = true

	// Extract hostname from request for testing purposes
	if len(request) < 12 {
		return nil, fmt.Errorf("request too short")
	}

	// Very simplified hostname extraction for testing
	var hostname string
	pos := 12
	for pos < len(request) {
		labelLen := int(request[pos])
		if 0 == labelLen {
			break
		}
		pos++
		if pos+labelLen > len(request) {
			return nil, fmt.Errorf("invalid format")
		}

		if 0 < len(hostname) {
			hostname += "."
		}
		hostname += string(request[pos : pos+labelLen])
		pos += labelLen
	}

	// Look up the response in our mock data
	if response, ok := c.mockForwarder.responses[hostname]; ok {
		// Copy the ID from the request to the response
		copy(response[0:2], request[0:2])
		return response, nil
	}

	// Create a generic response
	response := make([]byte, 512)
	// Copy ID from request
	copy(response[0:2], request[0:2])
	// Set QR bit and other flags
	binary.BigEndian.PutUint16(response[2:4], dnsQR|dnsRA|(binary.BigEndian.Uint16(request[2:4])&dnsRD))
	// Copy question count
	copy(response[4:6], request[4:6])
	// Set answer count to 0
	binary.BigEndian.PutUint16(response[6:8], 0)
	// Copy the question section
	questionLen := 0
	for i := 12; i < len(request); i++ {
		response[i] = request[i]
		questionLen++
	}

	return response[:12+questionLen], nil
} // ForwardDNSRequest()

// createDNSQuery creates a simple DNS query for testing
func createDNSQuery(hostname string, qType uint16) []byte {
	query := make([]byte, 512)

	// Set ID to 1234
	binary.BigEndian.PutUint16(query[0:2], 1234)
	// Set flags (RD bit)
	binary.BigEndian.PutUint16(query[2:4], dnsRD)
	// Set question count to 1
	binary.BigEndian.PutUint16(query[4:6], 1)
	// Set other counts to 0
	binary.BigEndian.PutUint16(query[6:8], 0)
	binary.BigEndian.PutUint16(query[8:10], 0)
	binary.BigEndian.PutUint16(query[10:12], 0)

	// Add the question section
	offset := 12

	// Add hostname labels
	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		query[offset] = byte(len(label))
		offset++
		copy(query[offset:], label)
		offset += len(label)
	}

	// Add terminating zero
	query[offset] = 0
	offset++

	// Add type and class
	binary.BigEndian.PutUint16(query[offset:], qType)
	offset += 2
	binary.BigEndian.PutUint16(query[offset:], dnsClassIN)
	offset += 2

	return query[:offset]
} // createDNSQuery()

// `createMalformedDNSRequest()` creates a DNS request with a malformed question section
func createMalformedDNSRequest() []byte {
	request := make([]byte, 512)

	// Set header fields
	binary.BigEndian.PutUint16(request[0:2], 1234)  // ID
	binary.BigEndian.PutUint16(request[2:4], dnsRD) // Flags (RD bit set)
	binary.BigEndian.PutUint16(request[4:6], 1)     // QDCOUNT = 1
	binary.BigEndian.PutUint16(request[6:8], 0)     // ANCOUNT = 0
	binary.BigEndian.PutUint16(request[8:10], 0)    // NSCOUNT = 0
	binary.BigEndian.PutUint16(request[10:12], 0)   // ARCOUNT = 0

	// Add a malformed question section that will cause questionLen to be 0
	// Just add the terminating zero immediately
	request[12] = 0

	// Add type and class
	binary.BigEndian.PutUint16(request[13:15], dnsTypeA)
	binary.BigEndian.PutUint16(request[15:17], dnsClassIN)

	return request[:17]
} // createMalformedDNSRequest()

// createMockMXResponse creates a mock MX record response
func createMockMXResponse(hostname, mailServer string) []byte {
	response := make([]byte, 512)

	// Set ID to 0 (will be copied from request)
	binary.BigEndian.PutUint16(response[0:2], 0)
	// Set flags (QR, AA, RA bits)
	binary.BigEndian.PutUint16(response[2:4], dnsQR|dnsAA|dnsRA)
	// Set question count to 1
	binary.BigEndian.PutUint16(response[4:6], 1)
	// Set answer count to 1
	binary.BigEndian.PutUint16(response[6:8], 1)
	// Set other counts to 0
	binary.BigEndian.PutUint16(response[8:10], 0)
	binary.BigEndian.PutUint16(response[10:12], 0)

	// Add the question section (simplified)
	offset := 12

	// Add hostname labels
	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		response[offset] = byte(len(label))
		offset++
		copy(response[offset:], label)
		offset += len(label)
	}

	// Add terminating zero
	response[offset] = 0
	offset++

	// Add type (MX = 15) and class
	binary.BigEndian.PutUint16(response[offset:], 15)
	offset += 2
	binary.BigEndian.PutUint16(response[offset:], dnsClassIN)
	offset += 2

	// Add the answer section (simplified)
	// Pointer to the question name
	binary.BigEndian.PutUint16(response[offset:], 0xC00C)
	offset += 2

	// Type (MX = 15) and class
	binary.BigEndian.PutUint16(response[offset:], 15)
	offset += 2
	binary.BigEndian.PutUint16(response[offset:], dnsClassIN)
	offset += 2

	// TTL (300 seconds)
	binary.BigEndian.PutUint32(response[offset:], 300)
	offset += 4

	// Data length placeholder
	dataLenPos := offset
	offset += 2

	// MX preference (10)
	binary.BigEndian.PutUint16(response[offset:], 10)
	offset += 2

	// Mail server name
	dataStartPos := offset
	labels = strings.Split(mailServer, ".")
	for _, label := range labels {
		response[offset] = byte(len(label))
		offset++
		copy(response[offset:], label)
		offset += len(label)
	}
	response[offset] = 0
	offset++

	// Update data length
	binary.BigEndian.PutUint16(response[dataLenPos:dataLenPos+2], uint16(offset-dataStartPos))

	return response[:offset]
} // createMockMXResponse()

// `createMockTXTResponse()` creates a mock TXT record response
func createMockTXTResponse(hostname, txtData string) []byte {
	response := make([]byte, 512)

	// Set ID to 0 (will be copied from request)
	binary.BigEndian.PutUint16(response[0:2], 0)
	// Set flags (QR, AA, RA bits)
	binary.BigEndian.PutUint16(response[2:4], dnsQR|dnsAA|dnsRA)
	// Set question count to 1
	binary.BigEndian.PutUint16(response[4:6], 1)
	// Set answer count to 1
	binary.BigEndian.PutUint16(response[6:8], 1)
	// Set other counts to 0
	binary.BigEndian.PutUint16(response[8:10], 0)
	binary.BigEndian.PutUint16(response[10:12], 0)

	// Add the question section (simplified)
	offset := 12

	// Add hostname labels
	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		response[offset] = byte(len(label))
		offset++
		copy(response[offset:], label)
		offset += len(label)
	}

	// Add terminating zero
	response[offset] = 0
	offset++

	// Add type (TXT = 16) and class
	binary.BigEndian.PutUint16(response[offset:], 16)
	offset += 2
	binary.BigEndian.PutUint16(response[offset:], dnsClassIN)
	offset += 2

	// Add the answer section
	// Pointer to the question name
	binary.BigEndian.PutUint16(response[offset:], 0xC00C)
	offset += 2

	// Type (TXT = 16) and class
	binary.BigEndian.PutUint16(response[offset:], 16)
	offset += 2
	binary.BigEndian.PutUint16(response[offset:], dnsClassIN)
	offset += 2

	// TTL (300 seconds)
	binary.BigEndian.PutUint32(response[offset:], 300)
	offset += 4

	// Data length placeholder
	dataLenPos := offset
	offset += 2

	// TXT data
	dataStartPos := offset
	response[offset] = byte(len(txtData))
	offset++
	copy(response[offset:], txtData)
	offset += len(txtData)

	// Update data length
	binary.BigEndian.PutUint16(response[dataLenPos:dataLenPos+2], uint16(offset-dataStartPos))

	return response[:offset]
} // createMockTXTResponse()

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
		if (len(labels) == i) || ('.' == labels[i]) {
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

func Test_handleDNSRequest(t *testing.T) {
	// Create a mock resolver
	resolver := dnscache.New(0)

	// Add a test entry to the resolver
	testHost := "example.org"
	testIP := net.ParseIP("192.168.2.1")
	_ = resolver.Create(context.TODO(), testHost, []net.IP{testIP}, 300)

	// Mock the lookup for non-existent domain to avoid actual DNS queries
	nonExistentDomain := "nonexistent.example"
	// Ensure this domain doesn't exist in the cache but will return NXDOMAIN immediately
	resolver.Create(context.TODO(), nonExistentDomain, []net.IP{}, 300)

	// Add debug logging
	t.Logf("Test setup complete. nonExistentDomain=%s", nonExistentDomain)

	// Verify the resolver setup
	ips, err := resolver.Fetch(nonExistentDomain)
	t.Logf("Resolver.Fetch(%q) returned: ips=%v, err=%v", nonExistentDomain, ips, err)

	// Create a test request for debugging
	testReq := createDNSRequest(5678, nonExistentDomain)
	t.Logf("Test request created: len=%d", len(testReq))

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

				return true
			},
			wantResp: true,
		},
		/* */
		{
			name:    "03 - non-existent domain",
			request: createDNSRequest(5678, nonExistentDomain),
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

				// Check response code (should be NXDOMAIN = 3 or NOERROR = 0 with empty answer)
				if (dnsRcodeNXDomain != rcode) && (dnsRcodeNoError != rcode) {
					t.Logf("Expected NXDOMAIN (3) or NOERROR (0), got rcode %d", rcode)
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
			// Set a timeout for this specific test case
			ctx, cancel := context.WithTimeout(context.Background(), time.Second<<3)
			defer cancel()

			// Create a mock connection with a channel for responses
			responseCh := make(chan []byte, 1)
			mockConn := &tMockPacketConn{
				respChan: responseCh,
			}

			// Add debug logging
			t.Logf("Starting test case: %s", tc.name)

			// Handle the request in a goroutine
			done := make(chan struct{})
			go func() {
				t.Logf("Calling handleDNSRequest")
				handleDNSRequest(mockConn, &tMockAddr{}, tc.request, resolver)
				t.Logf("handleDNSRequest returned")
				close(done)
			}()

			// Wait for response, completion, or timeout
			var resp []byte
			select {
			case resp = <-responseCh:
				// Got response
				t.Logf("Received response of %d bytes", len(resp))
			case <-done:
				// Handler completed without sending response
				if tc.wantResp {
					t.Logf("Handler completed without sending response")
				}
				resp = []byte{} // Empty response for validation
			case <-ctx.Done():
				t.Logf("Test timed out after 5 seconds")
				t.Fail()
				return
			case <-time.After(time.Second << 1):
				// Short timeout for tests that don't expect a response
				if !tc.wantResp {
					resp = []byte{} // Empty response for validation
				} else {
					t.Logf("Timed out waiting for response")
					t.Fail()
					return
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

func Test_handleDNSRequestWithForwarding(t *testing.T) {
	// Create a mock resolver
	resolver := dnscache.New(0)

	// Add a test entry to the resolver
	testHost := "example.org"
	testIP := net.ParseIP("192.168.2.1")
	_ = resolver.Create(context.TODO(), testHost, []net.IP{testIP}, 300)

	// Create a mock forwarder server
	mockForwarder := &tMockForwarder{
		responses: map[string][]byte{
			"mx.example.com":  createMockMXResponse("mx.example.com", "mail.example.com"),
			"txt.example.com": createMockTXTResponse("txt.example.com", "v=spf1 include:_spf.example.com ~all"),
		},
	}

	tests := []struct {
		name        string
		request     []byte
		forwarder   string
		checkResp   func([]byte) bool
		wantResp    bool
		wantForward bool
	}{
		/* */
		{
			name:      "01 - A record request (handled locally)",
			request:   createDNSQuery("example.org", dnsTypeA),
			forwarder: "",
			checkResp: func(resp []byte) bool {
				if 0 == len(resp) {
					return false
				}
				// Check if it's a response (QR bit set)
				flags := binary.BigEndian.Uint16(resp[2:4])
				if 0 == (flags & dnsQR) {
					return false
				}
				// Check if we have an answer
				answerCount := binary.BigEndian.Uint16(resp[6:8])
				return 0 < answerCount
			},
			wantResp:    true,
			wantForward: false,
		},
		{
			name:      "02 - MX record request (forwarded)",
			request:   createDNSQuery("mx.example.com", 15), // 15 = MX record
			forwarder: "8.8.8.8:53",
			checkResp: func(resp []byte) bool {
				if 0 == len(resp) {
					return false
				}
				// Check if it's a response (QR bit set)
				flags := binary.BigEndian.Uint16(resp[2:4])

				return (0 != (flags & dnsQR))
			},
			wantResp:    true,
			wantForward: true,
		},
		{
			name:      "03 - TXT record request (forwarded)",
			request:   createDNSQuery("txt.example.com", 16), // 16 = TXT record
			forwarder: "8.8.8.8:53",
			checkResp: func(resp []byte) bool {
				if 0 == len(resp) {
					return false
				}
				// Check if it's a response (QR bit set)
				flags := binary.BigEndian.Uint16(resp[2:4])

				return (0 != (flags & dnsQR))
			},
			wantResp:    true,
			wantForward: true,
		},
		{
			name:      "04 - A record with forwarder configured (still handled locally)",
			request:   createDNSQuery("example.org", dnsTypeA),
			forwarder: "8.8.8.8:53",
			checkResp: func(resp []byte) bool {
				if 0 == len(resp) {
					return false
				}
				// Check if it's a response (QR bit set)
				flags := binary.BigEndian.Uint16(resp[2:4])
				if 0 == (flags & dnsQR) {
					return false
				}
				// Check if we have an answer
				answerCount := binary.BigEndian.Uint16(resp[6:8])
				return 0 < answerCount
			},
			wantResp:    true,
			wantForward: false,
		},
		/* */
		{
			name:      "05 - malformed question with zero length",
			request:   createMalformedDNSRequest(),
			forwarder: "",
			checkResp: func(resp []byte) bool {
				return true // No validation needed as we don't expect a response
			},
			wantResp:    false,
			wantForward: false,
		},
		/* */
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock connection with a channel for responses
			responseCh := make(chan []byte, 1)
			mockConn := &tMockPacketConn{
				respChan: responseCh,
			}

			// Create a mock forwarder client
			mockClient := &tMockForwarderClient{
				mockForwarder: mockForwarder,
				forwardCalled: false,
			}

			// Handle the request directly with the forwarder
			handleDNSRequestWithForwarder(mockConn, &tMockAddr{}, tc.request, resolver, tc.forwarder, mockClient)

			// Wait for response or timeout
			var resp []byte
			select {
			case resp = <-responseCh:
				// Got response
				t.Logf("Received response of %d bytes", len(resp))
			case <-time.After(100 * time.Millisecond):
				resp = []byte{} // Empty response for validation
				t.Logf("Timed out waiting for response")
			}

			// Check if forwarding was called as expected
			if tc.wantForward != mockClient.forwardCalled {
				t.Errorf("handleDNSRequestWithForwarder() forwarding = %v, want %v",
					mockClient.forwardCalled, tc.wantForward)
			}

			// Check the response
			if !tc.checkResp(resp) {
				t.Errorf("handleDNSRequestWithForwarder() response validation failed")
			}
		})
	}
} // Test_handleDNSRequestWithForwarding()

func Test_startDNSserver(t *testing.T) {
	// Create a test resolver
	resolver := dnscache.New(0)

	tests := []struct {
		name      string
		resolver  *dnscache.TResolver
		address   string
		port      int
		forwarder string
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
				err = startDNSserver(tc.resolver, tc.address, tc.port, tc.forwarder)
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
