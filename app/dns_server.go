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
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mwat56/dnscache"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

// DNS message constants
const (
	// DNS header flags
	dnsQR uint16 = 1 << 15 // Query Response bit
	dnsAA uint16 = 1 << 10 // Authoritative Answer
	dnsTC uint16 = 1 << 9  // Truncated
	dnsRD uint16 = 1 << 8  // Recursion Desired
	dnsRA uint16 = 1 << 7  // Recursion Available

	// DNS response codes
	dnsRcodeNoError  uint16 = 0 // No error
	dnsRcodeFormErr  uint16 = 1 // Format error
	dnsRcodeServFail uint16 = 2 // Server failure
	dnsRcodeNXDomain uint16 = 3 // Non-existent domain
	dnsRcodeNotImp   uint16 = 4 // Not implemented
	dnsRcodeRefused  uint16 = 5 // Query refused

	// DNS record types
	dnsTypeA    uint16 = 1  // A record (IPv4)
	dnsTypeAAAA uint16 = 28 // AAAA record (IPv6)
	dnsClassIN  uint16 = 1  // Internet class
)

type (
	// `iForwarderClient` defines an interface for forwarding DNS requests.
	// It is used to decouple the DNS server from the forwarding mechanism
	// and allows for easier testing and potential future enhancements.
	iForwarderClient interface {
		ForwardDNSRequest(aCtx context.Context, aForwarder string, aRequest []byte) ([]byte, error)
	}

	// tStdForwarder implements the DNSForwarderClient interface using UDP.
	tStdForwarder struct{}
)

// `addAnswersToResponse()` adds DNS answers to a response.
//
// Parameters:
//   - `aResponse`: The DNS response being built.
//   - `aOffset`: The current offset in the response.
//   - `aAnswerCount`: The current answer count.
//   - `aIPs`: The IP addresses to add.
//   - `aQType`: The query type (A or AAAA).
//   - `aNameStart`: The start of the name in the request.
//
// Returns:
//   - `int`: The new offset in the response.
//   - `uint16`: The new answer count.
func addAnswersToResponse(aResponse []byte, aOffset int, aAnswerCount uint16,
	aIPs []net.IP, aQType uint16, aNameStart int) (int, uint16) {
	offset := aOffset
	answerCount := aAnswerCount

	for _, ip := range aIPs {
		if dnsTypeA == aQType {
			// For A records, we need IPv4 addresses
			ip4 := ip.To4()
			if nil == ip4 {
				// Skip if not IPv4
				continue
			}

			// Check if we have enough space in the response
			if offset+16 > len(aResponse) {
				// Response too large, stop adding answers
				break
			}

			// Add name pointer to question
			binary.BigEndian.PutUint16(aResponse[offset:offset+2], 0xC000|uint16(aNameStart)) //#nosec G115
			offset += 2

			// Add type, class, TTL, and data length
			binary.BigEndian.PutUint16(aResponse[offset:offset+2], dnsTypeA)
			offset += 2
			binary.BigEndian.PutUint16(aResponse[offset:offset+2], dnsClassIN)
			offset += 2
			binary.BigEndian.PutUint32(aResponse[offset:offset+4], 300) // 5 minutes TTL
			offset += 4
			binary.BigEndian.PutUint16(aResponse[offset:offset+2], 4) // IPv4 address length
			offset += 2

			// Add IPv4 address
			copy(aResponse[offset:offset+4], ip4)
			offset += 4

			answerCount++
		} else if dnsTypeAAAA == aQType {
			// For AAAA records, we need IPv6 addresses
			ip6 := ip.To16()
			if nil == ip6 || nil != ip.To4() {
				// Skip if not IPv6 or if it's an IPv4 mapped to IPv6
				continue
			}

			// Check if we have enough space in the response
			if offset+28 > len(aResponse) {
				// Response too large, stop adding answers
				break
			}

			// Add name pointer to question
			binary.BigEndian.PutUint16(aResponse[offset:offset+2], 0xC000|uint16(aNameStart)) //#nosec G115
			offset += 2

			// Add type, class, TTL, and data length
			binary.BigEndian.PutUint16(aResponse[offset:offset+2], dnsTypeAAAA)
			offset += 2
			binary.BigEndian.PutUint16(aResponse[offset:offset+2], dnsClassIN)
			offset += 2
			binary.BigEndian.PutUint32(aResponse[offset:offset+4], 300) // 5 minutes TTL
			offset += 4
			binary.BigEndian.PutUint16(aResponse[offset:offset+2], 16) // IPv6 address length
			offset += 2

			// Add IPv6 address
			copy(aResponse[offset:offset+16], ip6)
			offset += 16

			answerCount++
		} // if dnsTypeA == aQType
	} // for _, ip := range aIPs

	return offset, answerCount
} // addAnswersToResponse()

// extractFirstHostname extracts the first hostname from a DNS request
func extractFirstHostname(aRequest []byte) string {
	if len(aRequest) <= 12 {
		return ""
	}

	offset := 12
	var hostname strings.Builder

	for {
		if offset >= len(aRequest) {
			return ""
		}

		labelLen := int(aRequest[offset])
		if 0 == labelLen {
			break
		}

		offset++
		if offset+labelLen > len(aRequest) {
			return ""
		}

		if hostname.Len() > 0 {
			hostname.WriteByte('.')
		}

		hostname.Write(aRequest[offset : offset+labelLen])
		offset += labelLen
	}

	return hostname.String()
} // extractFirstHostname()

// `extractHostname()` extracts a hostname from a DNS question section.
//
// Parameters:
//   - `aQuestion`: The DNS question.
//
// Returns:
//   - `string`: The extracted hostname.
func extractHostname(aQuestion []byte) string {
	if 0 == len(aQuestion) {
		return ""
	}

	var hostname strings.Builder
	pos := 0
	for pos < len(aQuestion) {
		labelLen := int(aQuestion[pos])
		if 0 == labelLen {
			break
		}
		pos++
		if pos+labelLen > len(aQuestion) {
			/*
				Returning an empty string when encountering an
				invalid format is based on security and
				robustness principles:

				Security First: In DNS processing, which is
				security-sensitive, it's safer to reject
				malformed data entirely rather than try to
				salvage partial information. Returning a
				partial hostname could lead to incorrect
				DNS resolution or potentially security
				vulnerabilities.

				Fail Fast and Explicitly: When processing
				network protocols like DNS, the "fail fast"
				principle is important. If the data doesn't
				conform to the expected format, it's better
				to reject it completely rather than try to
				interpret it, which could lead to unpredictable
				behaviour.

				Protocol Compliance: DNS has strict formatting
				requirements. A malformed label length is a
				protocol violation, and the proper handling is
				to reject the entire query rather than try to
				extract partial information.

				Preventing Ambiguity: Returning a partial
				hostname could lead to ambiguous situations
				where the system might resolve a different
				domain than intended, potentially leading to
				DNS spoofing or cache poisoning attacks.

				Consistent Error Handling: Having a clear,
				consistent approach to error handling (return
				an empty string on any format error) makes the
				code more maintainable and predictable than
				trying to handle different error cases
				differently.
			*/
			return "" // Invalid format
		}

		if 0 < hostname.Len() {
			hostname.WriteByte('.')
		}
		hostname.Write(aQuestion[pos : pos+labelLen])
		pos += labelLen
	}

	return hostname.String()
} // extractHostname()

// `ForwardDNSRequest()` forwards a DNS request to the specified forwarder
// and returns the response.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aForwarder`: The DNS forwarder to use.
//   - `aRequest`: The DNS request to forward.
//
// Returns:
//   - `[]byte`: The DNS response.
//   - `error`: `nil` if the request was forwarded successfully, the error otherwise.
func (f *tStdForwarder) ForwardDNSRequest(aCtx context.Context, aForwarder string, aRequest []byte) ([]byte, error) {
	// Create a UDP connection to the forwarder
	conn, err := net.Dial("udp", aForwarder)
	if nil != err {
		return nil, fmt.Errorf("failed to connect to forwarder: %w", err)
	}
	defer conn.Close()

	// Set deadline based on context
	deadline, ok := aCtx.Deadline()
	if ok {
		if err := conn.SetDeadline(deadline); nil != err {
			return nil, fmt.Errorf("failed to set connection deadline: %w", err)
		}
	} else {
		// Default timeout of 8 seconds if no deadline in context
		if err := conn.SetDeadline(time.Now().Add(time.Second << 3)); nil != err {
			return nil, fmt.Errorf("failed to set connection deadline: %w", err)
		}
	}

	// Send the request
	if _, err := conn.Write(aRequest); nil != err {
		return nil, fmt.Errorf("failed to send request to forwarder: %w", err)
	}

	// Read the response
	response := make([]byte, 512)
	n, err := conn.Read(response)
	if nil != err {
		return nil, fmt.Errorf("failed to read response from forwarder: %w", err)
	}

	return response[:n], nil
} // ForwardDNSRequest()

// `forwardRequest()` forwards a DNS request to the specified forwarder.
//
// Parameters:
//   - `aConn`: The UDP connection to write response to.
//   - `aAddr`: The address to send response to.
//   - `aRequest`: The DNS request message.
//   - `aID`: The DNS request ID.
//   - `aFlags`: The DNS request flags.
//   - `aQDCount`: The number of questions in the request.
//   - `aForwarder`: The DNS forwarder to use.
//   - `aForwarderClient`: The client to use for forwarding requests.
func forwardRequest(aConn net.PacketConn, aAddr net.Addr, aRequest []byte,
	aID, aFlags, aQDCount uint16, aForwarder string, aForwarderClient iForwarderClient) {
	// Forward the request
	ctx, cancel := context.WithTimeout(context.Background(), time.Second<<3)
	defer cancel()

	// Forward the request
	response, err := aForwarderClient.ForwardDNSRequest(ctx, aForwarder, aRequest)
	if nil != err {
		// Send NXDOMAIN response
		sendNXDOMAINResponse(aConn, aAddr, aID, aFlags, aQDCount, aRequest[12:])
		return
	}

	// Send the response from the forwarder
	_, _ = aConn.WriteTo(response, aAddr)
	// Error sending response is not critical, hence we ignore it.
} // forwardRequest()

// `handleDNSRequest()` processes a DNS request and sends a response.
//
// Parameters:
//   - `aConn`: The UDP connection to write response to.
//   - `aAddr`: The address to send response to.
//   - `aRequest`: The DNS request message.
//   - `aResolver`: The DNS resolver to use for lookups.
func handleDNSRequest(aConn net.PacketConn, aAddr net.Addr, aRequest []byte, aResolver *dnscache.TResolver) {
	// Use the new function with an empty forwarder string
	handleDNSRequestWithForwarder(aConn, aAddr, aRequest, aResolver, "", &tStdForwarder{})
} // handleDNSRequest()

// `handleDNSRequestWithForwarder()` processes a DNS request and sends a response,
// forwarding non-A/AAAA requests to the specified forwarder if provided.
//
// Parameters:
//   - `aConn`: The UDP connection to write response to.
//   - `aAddr`: The address to send response to.
//   - `aRequest`: The DNS request message.
//   - `aResolver`: The DNS resolver to use for lookups.
//   - `aForwarder`: The DNS forwarder to use for non-A/AAAA requests (empty string means no forwarding).
//   - `aForwarderClient`: The client to use for forwarding requests.
func handleDNSRequestWithForwarder(aConn net.PacketConn, aAddr net.Addr, aRequest []byte,
	aResolver *dnscache.TResolver, aForwarder string, aForwarderClient iForwarderClient) {
	log.Printf("DEBUG: handleDNSRequestWithForwarder called with request length %d", len(aRequest))

	// Check if request is too short
	if 12 > len(aRequest) {
		log.Printf("DEBUG: Request too short (%d bytes), minimum 12 required", len(aRequest))
		return
	}

	// Parse DNS request header
	requestID := binary.BigEndian.Uint16(aRequest[0:2])
	requestFlags := binary.BigEndian.Uint16(aRequest[2:4])
	requestQDCount := binary.BigEndian.Uint16(aRequest[4:6])

	log.Printf("DEBUG: Request ID=%d, Flags=0x%04x, QDCount=%d", requestID, requestFlags, requestQDCount)

	// First pass: check if we need to forward any questions
	if shouldForwardRequest(aRequest, requestQDCount, aForwarder) {
		log.Printf("DEBUG: Forwarding request to %s", aForwarder)
		forwardRequest(aConn, aAddr, aRequest, requestID, requestFlags, requestQDCount, aForwarder, aForwarderClient)
		return
	}

	// Second pass: handle A/AAAA records locally
	log.Printf("DEBUG: Handling request locally")
	handleLocalRequest(aConn, aAddr, aRequest, requestID, requestFlags, requestQDCount, aResolver)
} // handleDNSRequestWithForwarder()

// `handleLocalRequest()` handles a DNS request locally.
//
// Parameters:
//   - `aConn`: The UDP connection to write response to.
//   - `aAddr`: The address to send response to.
//   - `aRequest`: The DNS request message.
//   - `aID`: The DNS request ID.
//   - `aFlags`: The DNS request flags.
//   - `aQDCount`: The number of questions in the request.
//   - `aResolver`: The DNS resolver to use for lookups.
func handleLocalRequest(aConn net.PacketConn, aAddr net.Addr, aRequest []byte,
	aID, aFlags, aQDCount uint16, aResolver *dnscache.TResolver) {
	log.Printf("DEBUG: handleLocalRequest called with ID=%d, QDCount=%d", aID, aQDCount)

	// For non-existent domains, send NXDOMAIN response immediately
	if 0 < aQDCount {
		// Extract the first hostname
		hostname := extractFirstHostname(aRequest)
		log.Printf("DEBUG: First hostname in request: %s", hostname)

		if "" != hostname {
			// Try to lookup the hostname
			ips, err := aResolver.Fetch(hostname)
			log.Printf("DEBUG: Resolver.Fetch(%s) returned: ips=%v, err=%v", hostname, ips, err)

			// If lookup fails, send NXDOMAIN immediately
			if (nil != err) || (0 == len(ips)) {
				log.Printf("DEBUG: Sending NXDOMAIN response for %s", hostname)
				sendNXDOMAINResponse(aConn, aAddr, aID, aFlags, aQDCount, aRequest[12:])
				return
			}
		}
	}

	// Prepare response for A/AAAA records
	response := make([]byte, 512)

	// Set response header
	binary.BigEndian.PutUint16(response[0:2], aID)
	binary.BigEndian.PutUint16(response[2:4], dnsQR|dnsAA|dnsRA|(aFlags&dnsRD))
	binary.BigEndian.PutUint16(response[4:6], aQDCount)
	binary.BigEndian.PutUint16(response[6:8], 0)   // ANCount - will update later
	binary.BigEndian.PutUint16(response[8:10], 0)  // NSCount
	binary.BigEndian.PutUint16(response[10:12], 0) // ARCount

	responseOffset := 12
	answerCount := uint16(0)
	questionProcessed := false

	// Process each question with safety checks
	currentOffset := 12
	questionStart := 12

	// Ensure we don't exceed the request length
	if 12 >= len(aRequest) {
		// Malformed request, send FORMERR
		binary.BigEndian.PutUint16(response[2:4], dnsQR|dnsAA|dnsRA|(aFlags&dnsRD)|dnsRcodeFormErr)
		_, _ = aConn.WriteTo(response[:12], aAddr)
		return
	}

	// Process questions with a limit to prevent infinite loops
	for i := uint16(0); i < aQDCount && 10 > i; i++ { // Limit to 10 questions max
		if currentOffset >= len(aRequest) {
			break // Exit if we've reached the end of the request
		}

		nameStart := currentOffset

		// Skip over domain name labels with safety checks
		labelCount := 0
		for labelCount < 64 { // Limit to 64 labels max to prevent infinite loops
			if currentOffset >= len(aRequest) {
				break // Exit if we've reached the end of the request
			}

			labelLen := int(aRequest[currentOffset])
			if 0 == labelLen {
				// End of domain name
				currentOffset++
				break
			}

			// Move to the next label
			currentOffset += labelLen + 1
			labelCount++

			if currentOffset >= len(aRequest) {
				break // Exit if we've reached the end of the request
			}
		}

		// Check if we have enough bytes for type and class
		if currentOffset+4 > len(aRequest) {
			break // Exit if we don't have enough bytes
		}

		// Extract type and class
		qType := binary.BigEndian.Uint16(aRequest[currentOffset : currentOffset+2])
		qClass := binary.BigEndian.Uint16(aRequest[currentOffset+2 : currentOffset+4])
		currentOffset += 4

		questionLen := currentOffset - questionStart
		questionProcessed = true

		// Copy question to response
		if len(response) >= responseOffset+questionLen {
			copy(response[responseOffset:responseOffset+questionLen],
				aRequest[questionStart:currentOffset])
			responseOffset += questionLen
		} else {
			// Response would be too large, set TC bit
			binary.BigEndian.PutUint16(response[2:4], binary.BigEndian.Uint16(response[2:4])|dnsTC) // Set truncated bit
			break
		}

		// Only process IN class A/AAAA records
		if (dnsClassIN == qClass) && ((dnsTypeA == qType) || (dnsTypeAAAA == qType)) {
			// Extract hostname
			var hostname string
			if questionLen > 4 {
				hostname = extractHostname(aRequest[nameStart : nameStart+questionLen-4])
			}

			if "" != hostname {
				// Lookup IP addresses
				ips, err := aResolver.Fetch(hostname)
				if (nil != err) || (0 == len(ips)) {
					// Set NXDOMAIN if lookup fails
					binary.BigEndian.PutUint16(response[2:4], dnsQR|dnsAA|dnsRA|(aFlags&dnsRD)|dnsRcodeNXDomain)
				} else {
					// Add answers to response
					newOffset, newAnswerCount := addAnswersToResponse(response, responseOffset, answerCount, ips, qType, nameStart)
					responseOffset = newOffset
					answerCount = newAnswerCount
				}
			}
		}

		questionStart = currentOffset
	}

	// Update answer count in header
	binary.BigEndian.PutUint16(response[6:8], answerCount)

	// Always send a response
	if questionProcessed {
		_, _ = aConn.WriteTo(response[:responseOffset], aAddr)
	} else if 0 < aQDCount {
		// If we have questions but couldn't process any, send FORMERR
		binary.BigEndian.PutUint16(response[2:4], dnsQR|dnsAA|dnsRA|(aFlags&dnsRD)|dnsRcodeFormErr)
		_, _ = aConn.WriteTo(response[:12], aAddr)
	}
	// Error sending response is not critical, hence we ignore it.
} // handleLocalRequest()

/* * /
// `processARecord()` processes an A or AAAA record query.
//
// Parameters:
//   - `aRequest`: The DNS request message.
//   - `aResponse`: The DNS response being built.
//   - `aOffset`: The current offset in the response.
//   - `aAnswerCount`: The current answer count.
//   - `aNameStart`: The start of the name in the request.
//   - `aQuestionLen`: The length of the question.
//   - `aQType`: The query type (A or AAAA).
//   - `aResolver`: The DNS resolver to use for lookups.
//   - `aConn`: The UDP connection to write response to.
//   - `aAddr`: The address to send response to.
//   - `aFlags`: The DNS request flags.
//
// Returns:
//   - `int`: The new offset in the response.
//   - `uint16`: The new answer count.
func processARecord(aRequest, aResponse []byte, aOffset int, aAnswerCount uint16,
	aNameStart, aQuestionLen int, aQType uint16, aResolver *dnscache.TResolver,
	aConn net.PacketConn, aAddr net.Addr, aFlags uint16) (int, uint16) {
	// Extract hostname
	var hostname string
	if aQuestionLen > 4 {
		hostname = extractHostname(aRequest[aNameStart : aNameStart+aQuestionLen-4])
	}

	if "" == hostname {
		return aOffset, aAnswerCount
	}

	// Lookup IP addresses
	ips, err := aResolver.Fetch(hostname)
	if (nil != err) || (0 == len(ips)) {
		// Set NXDOMAIN if lookup fails, but don't send response here
		binary.BigEndian.PutUint16(aResponse[2:4], dnsQR|dnsAA|dnsRA|(aFlags&dnsRD)|dnsRcodeNXDomain)

		// Don't send response here, let handleLocalRequest do it
		// Just return the current offset and answer count
		return aOffset, aAnswerCount
	}

	// Add answers to response
	return addAnswersToResponse(aResponse, aOffset, aAnswerCount, ips, aQType, aNameStart)
} // processARecord()
/* */

// `sendNXDOMAINResponse()` sends a DNS response with NXDOMAIN status.
//
// Parameters:
//   - `aConn`: The UDP connection to write response to.
//   - `aAddr`: The address to send response to.
//   - `aID`: The DNS request ID.
//   - `aFlags`: The DNS request flags.
//   - `aQDCount`: The DNS request question count.
//   - `aQuestion`: The DNS question section.
func sendNXDOMAINResponse(aConn net.PacketConn, aAddr net.Addr, aID, aFlags, aQDCount uint16, aQuestion []byte) {
	log.Printf("DEBUG: Sending NXDOMAIN response for ID=%d", aID)

	// Prepare response
	response := make([]byte, 512)

	// Set response header
	binary.BigEndian.PutUint16(response[0:2], aID)
	binary.BigEndian.PutUint16(response[2:4], dnsQR|dnsAA|dnsRA|(aFlags&dnsRD)|dnsRcodeNXDomain)
	binary.BigEndian.PutUint16(response[4:6], aQDCount)
	binary.BigEndian.PutUint16(response[6:8], 0)   // ANCount
	binary.BigEndian.PutUint16(response[8:10], 0)  // NSCount
	binary.BigEndian.PutUint16(response[10:12], 0) // ARCount

	// Copy question to response
	questionLen := min(len(aQuestion), 500) // Ensure we don't exceed response buffer
	copy(response[12:12+questionLen], aQuestion[:questionLen])

	// Send response
	n, err := aConn.WriteTo(response[:12+questionLen], aAddr)
	log.Printf("DEBUG: NXDOMAIN response sent: %d bytes, err=%v", n, err)
} // sendNXDOMAINResponse()

// `shouldForwardRequest()` determines if a DNS request should be forwarded.
//
// Parameters:
//   - `aRequest`: The DNS request message.
//   - `aQDCount`: The number of questions in the request.
//   - `aForwarder`: The DNS forwarder to use (empty string means no forwarding).
//
// Returns:
//   - `bool`: `true` if the request should be forwarded, `false` otherwise.
func shouldForwardRequest(aRequest []byte, aQDCount uint16, aForwarder string) bool {
	// If no forwarder is configured, we don't forward
	if "" == aForwarder {
		return false
	}

	currentOffset := 12
	// Process each question to determine if we need to forward
	for range aQDCount {
		// Skip over the domain name labels
		for {
			if currentOffset >= len(aRequest) {
				return false // Malformed request
			}

			labelLen := int(aRequest[currentOffset])
			if 0 == labelLen {
				// End of domain name
				currentOffset++
				break
			}

			// Move to the next label
			currentOffset += labelLen + 1
			if currentOffset >= len(aRequest) {
				return false // Malformed request
			}
		}

		// Check if we have enough bytes for type and class
		if currentOffset+4 > len(aRequest) {
			return false // Malformed request
		}

		// Extract type and class
		qType := binary.BigEndian.Uint16(aRequest[currentOffset : currentOffset+2])
		qClass := binary.BigEndian.Uint16(aRequest[currentOffset+2 : currentOffset+4])
		currentOffset += 4

		// If this is not an A or AAAA record query and we have a forwarder,
		// we should forward the request
		if dnsTypeA != qType && dnsTypeAAAA != qType && dnsClassIN == qClass {
			return true
		}
	}

	return false
} // shouldForwardRequest()

// `startDNSserver()` starts a DNS server on the specified address and port.
//
// Parameters:
//   - `aResolver`: The DNS resolver to use.
//   - `aAddress`: The IP address to bind to (empty string means all addresses).
//   - `aPort`: The port to listen on.
//   - `aForwarder`: The DNS forwarder to use for non-A/AAAA requests (empty string means no forwarding).
//
// Returns:
//   - `error`: `nil` if the server started successfully, otherwise the error that occurred.
func startDNSserver(aResolver *dnscache.TResolver, aAddress string, aPort int, aForwarder string) error {
	if nil == aResolver {
		return fmt.Errorf("nil resolver provided")
	}

	if 0 >= aPort || 65535 < aPort {
		return fmt.Errorf("invalid port number: %d", aPort)
	}

	// Create UDP listener
	listenAddr := fmt.Sprintf("%s:%d", aAddress, aPort)
	conn, err := net.ListenPacket("udp", listenAddr)
	if nil != err {
		//TODO: implement retry logic
		return fmt.Errorf("failed to start DNS server: %w", err)
	}

	// Setup signal handling for graceful shutdown
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Create a channel to signal shutdown to the handler goroutine
	done := make(chan struct{})

	// Create a forwarder client
	forwarderClient := &tStdForwarder{}

	// Start handler in a goroutine
	go func() {
		log.Printf("Starting DNS server on %s", listenAddr)
		if "" != aForwarder {
			log.Printf("Using DNS forwarder: %s", aForwarder)
		}

		buffer := make([]byte, 512) // Standard DNS message size
		for {
			select {
			case <-done:
				return

			default:
				// Set read deadline to allow checking for shutdown signal
				if err := conn.SetReadDeadline(time.Now().Add(time.Second)); nil != err {
					log.Printf("Error setting read deadline: %v", err)
				}

				// Read incoming DNS request
				n, addr, err := conn.ReadFrom(buffer)
				if nil != err {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						// This is just a timeout, continue to check for shutdown
						continue
					}
					log.Printf("Error reading DNS request: %v", err)
					continue
				}

				// Handle the DNS request in a separate goroutine
				go handleDNSRequestWithForwarder(conn, addr, buffer[:n], aResolver, aForwarder, forwarderClient)
			} // select
		} // for
	}() // go func()

	// Wait for termination signal
	<-sig
	log.Println("Shutting down DNS server ...")
	// Signal handler goroutine to stop
	close(done)

	// Stop background refresh and expire
	aResolver.StopRefresh().StopExpire()

	// Close the connection
	if err := conn.Close(); nil != err {
		return fmt.Errorf("error closing connection: %w", err)
	}

	log.Println("DNS server shutdown complete")
	return nil
} // startDNSserver()

/* _EoF_ */
