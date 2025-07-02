/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package main

import (
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
	dnsRD uint16 = 1 << 8  // Recursion Desired
	dnsRA uint16 = 1 << 7  // Recursion Available
	// dnsRcodeNoError  uint16 = 0       // No error
	dnsRcodeNXDomain uint16 = 3 // Non-existent domain

	// DNS record types
	dnsTypeA   uint16 = 1 // A record (IPv4)
	dnsClassIN uint16 = 1 // Internet class
)

// type (
// 	// DNS message header structure
// 	tDNSheader struct {
// 		ID      uint16 // Query ID
// 		Flags   uint16 // Query/Response flags
// 		QDCount uint16 // Question count
// 		ANCount uint16 // Answer count
// 		NSCount uint16 // Authority count
// 		ARCount uint16 // Additional record count
// 	}
// )

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

// `handleDNSRequest()` processes a DNS request and sends a response.
//
// Parameters:
//   - `aConn`: The UDP connection to write response to.
//   - `aAddr`: The address to send response to.
//   - `aRequest`: The DNS request message.
//   - `aResolver`: The DNS resolver to use for lookups.
func handleDNSRequest(aConn net.PacketConn, aAddr net.Addr, aRequest []byte, aResolver *dnscache.TResolver) {
	// log.Printf("Handling DNS request of %d bytes", len(aRequest))

	// Check if request is too short
	if 12 > len(aRequest) {
		// log.Printf("Request too short: %d bytes", len(aRequest))
		return
	}

	// Parse DNS request header
	requestID := binary.BigEndian.Uint16(aRequest[0:2])
	requestFlags := binary.BigEndian.Uint16(aRequest[2:4])
	requestQDCount := binary.BigEndian.Uint16(aRequest[4:6])

	// log.Printf("Request ID: %d, Flags: 0x%04x, Question count: %d",
	// 	requestID, requestFlags, requestQDCount)

	// Prepare response
	response := make([]byte, 512)

	// Set response header
	binary.BigEndian.PutUint16(response[0:2], requestID)
	binary.BigEndian.PutUint16(response[2:4], dnsQR|dnsAA|dnsRA|(requestFlags&dnsRD))
	binary.BigEndian.PutUint16(response[4:6], requestQDCount)
	binary.BigEndian.PutUint16(response[6:8], 0)   // ANCount - will update later
	binary.BigEndian.PutUint16(response[8:10], 0)  // NSCount
	binary.BigEndian.PutUint16(response[10:12], 0) // ARCount

	responseOffset := 12
	answerCount := uint16(0)

	// Copy questions from request to response
	questionStart := 12
	currentOffset := 12

	// Process each question
	for range requestQDCount {
		// log.Printf("Processing question %d", i+1)

		// Find the end of the domain name
		nameStart := currentOffset
		for {
			if currentOffset >= len(aRequest) {
				// log.Printf("Invalid request: currentOffset (%d) >= request length (%d)", currentOffset, len(aRequest))
				return
			}

			labelLen := aRequest[currentOffset]
			if 0 == labelLen {
				currentOffset++ // Skip the terminating zero
				break
			}

			// Check for malformed labels
			if 63 < labelLen {
				// log.Printf("Malformed query: label length %d > 63", labelLen)
				return
			}

			// Make sure we don't go out of bounds
			if currentOffset+int(labelLen)+1 > len(aRequest) {
				// log.Printf("Malformed query: label would exceed request bounds")
				return
			}

			currentOffset += int(labelLen) + 1
		}

		// Make sure we have enough space for type and class
		if currentOffset+4 > len(aRequest) {
			// log.Printf("Invalid request: not enough space for type and class")
			return
		}

		// Extract question type and class
		qType := binary.BigEndian.Uint16(aRequest[currentOffset : currentOffset+2])
		qClass := binary.BigEndian.Uint16(aRequest[currentOffset+2 : currentOffset+4])
		currentOffset += 4

		// log.Printf("Question type: %d, class: %d", qType, qClass)

		// Copy question to response
		questionLen := currentOffset - questionStart
		if responseOffset+questionLen > len(response) {
			// log.Printf("Response would be too large")
			return
		}
		copy(response[responseOffset:responseOffset+questionLen], aRequest[questionStart:currentOffset])
		responseOffset += questionLen

		// Process A record queries
		if dnsTypeA == qType && dnsClassIN == qClass {
			// Extract hostname
			hostname := extractHostname(aRequest[nameStart : nameStart+questionLen-4])
			if "" == hostname {
				// log.Printf("Invalid hostname")
				return
			}

			// log.Printf("Looking up hostname: %s", hostname)

			// Lookup IP addresses
			ips, err := aResolver.Fetch(hostname)
			if nil != err || 0 == len(ips) {
				// log.Printf("Lookup failed or no IPs found: %v", err)

				// Set NXDOMAIN if lookup fails
				binary.BigEndian.PutUint16(response[2:4], dnsQR|dnsAA|dnsRA|(requestFlags&dnsRD)|dnsRcodeNXDomain)

				// log.Printf("Sending NXDOMAIN response of %d bytes", responseOffset)
				if _, err := aConn.WriteTo(response[:responseOffset], aAddr); nil != err {
					// log.Printf("Error sending DNS response: %v", err)
				}
				return
			}

			// log.Printf("Found %d IPs for %s", len(ips), hostname)

			// Add A records to response
			for _, ip := range ips {
				ip4 := ip.To4()
				if nil == ip4 {
					// log.Printf("Skipping non-IPv4 address: %s", ip.String())
					continue
				}

				// Check if we have enough space in the response
				if responseOffset+16 > len(response) {
					// log.Printf("Response too large, stopping at %d bytes", responseOffset)
					break
				}

				// Add name pointer to question
				binary.BigEndian.PutUint16(response[responseOffset:responseOffset+2], 0xC000|uint16(nameStart)) //#nosec G115
				responseOffset += 2

				// Add type, class, TTL, and data length
				binary.BigEndian.PutUint16(response[responseOffset:responseOffset+2], dnsTypeA)
				responseOffset += 2
				binary.BigEndian.PutUint16(response[responseOffset:responseOffset+2], dnsClassIN)
				responseOffset += 2
				binary.BigEndian.PutUint32(response[responseOffset:responseOffset+4], 300) // 5 minutes TTL
				responseOffset += 4
				binary.BigEndian.PutUint16(response[responseOffset:responseOffset+2], 4) // IPv4 address length
				responseOffset += 2

				// Add IPv4 address
				copy(response[responseOffset:responseOffset+4], ip4)
				responseOffset += 4

				answerCount++
				// log.Printf("Added A record for %s: %s", hostname, ip.String())
			}
			// } else {
			// log.Printf("Unsupported query type: %d", qType)
		}

		questionStart = currentOffset
	}

	// Update answer count in header
	binary.BigEndian.PutUint16(response[6:8], answerCount)

	// log.Printf("Sending response of %d bytes with %d answers", responseOffset, answerCount)

	// Send response
	_, _ = aConn.WriteTo(response[:responseOffset], aAddr)
	// n, err := aConn.WriteTo(response[:responseOffset], aAddr)
	// if nil != err {
	// 	log.Printf("Error sending DNS response: %v", err)
	// } else {
	// 	log.Printf("Sent %d bytes to %s", n, aAddr.String())
	// }
} // handleDNSRequest()

// `startDNSserver()` starts a DNS server on the specified address and port.
//
// Parameters:
//   - `aResolver`: The DNS resolver to use.
//   - `aAddress`: The IP address to bind to (empty string means all addresses).
//   - `aPort`: The port to listen on.
//
// Returns:
//   - `error`: `nil` if the server started successfully, otherwise the error that occurred.
func startDNSserver(aResolver *dnscache.TResolver, aAddress string, aPort int) error {
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

	// Start handler in a goroutine
	go func() {
		log.Printf("Starting DNS server on %s", listenAddr)

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
				go handleDNSRequest(conn, addr, buffer[:n], aResolver)
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
