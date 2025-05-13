/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package dnscache

import (
	"bufio"
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

// type (
// 	// `tDefaultResolver` is the resolver used by the package-level
// 	// Lookup functions.

// 	tDefaultResolver struct {
// 		*net.Resolver
// 	}
// )

// func (r *tDefaultResolver) Dial(ctx context.Context, aNetType, aHostname string) (net.Conn, error) {

// 	dialer := net.Dialer{
// 		Timeout: time.Second * 3,
// 	}

// 	if "udp" == aNetType || "tcp" == aNetType {
// 		return dialer.DialContext(ctx, aNetType, aServer+":53")
// 	}

// 	return nil, fmt.Errorf("unsupported network: %s", aNetType)
// } // Dial()

// func (r *tDefaultResolver) preferGo() bool { return r != nil && r.PreferGo }

// func (r *tDefaultResolver) strictErrors() bool { return r != nil && r.StrictErrors }

// // DefaultResolver is the resolver used by the package-level Lookup functions.
// var DefaultResolver = &net.Resolver{}

// `getDNSServers()` reads the DNS servers from `/etc/resolv.conf`.
//
// Returns:
//   - `[]string`: List of DNS servers.
//   - `error`: `nil` if the DNS servers were read successfully, the error otherwise.
func getDNSServers() ([]string, error) {
	file, err := os.Open("/etc/resolv.conf")
	if nil != err {
		return nil, err
	}
	defer file.Close()

	var (
		fields []string
		line   string
		result []string
	)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if line = strings.TrimSpace(scanner.Text()); 0 == len(line) {
			continue
		}
		// Ignore comment lines
		if "#" == string(line[0]) || ";" == string(line[0]) {
			continue
		}

		// Add lines that are valid nameserver entries
		if fields = strings.Fields(line); (1 < len(fields)) &&
			("nameserver" == fields[0]) {
			// Check if the IP is a valid address
			if nil != net.ParseIP(fields[1]) {
				result = append(result, fields[1])
			}
		}
	} // for scanner.Scan()

	if err = scanner.Err(); nil != err {
		return nil, err
	}

	// Check if we found any DNS servers
	if 0 == len(result) {
		return nil, errors.New("no DNS servers found")
	}

	// Since the order of entries is significant
	// we must not sort the list!

	return result, nil
} // getDNSServers()

// `lookupDNS()` resolves a hostname using a specific DNS server.
//
// Parameters:
//   - `aCtx`: Context for the lookup operation.
//   - `aServer`: DNS server to use.
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `[]net.IP`: List of IP addresses for the given hostname.
//   - `error`: `nil` if the hostname was resolved successfully, the error otherwise.
func lookupDNS(aCtx context.Context, aServer, aHostname string) ([]net.IP, error) {
	var (
		addrs []string
		err   error
		ip    net.IP
	)

	resolver := &net.Resolver{
		PreferGo: true, // Use Go's built-in DNS resolver

		// `Dial` is used to connect to the DNS server
		Dial: func(aCtx context.Context, aNetType, _ string) (net.Conn, error) {
			dialer := net.Dialer{
				Timeout: time.Second * 3,
			}
			// switch aNetType {
			// case "tcp", "tcp4", "tcp6":
			// case "udp", "udp4", "udp6":
			// // case "ip", "ip4", "ip6":
			// default:
			// 	return nil, net.UnknownNetworkError(aNetType)
			// }

			return dialer.DialContext(aCtx, aNetType, aServer+":53")
		}, // Dial
	} // resolver

	// Do the DNS lookup
	if addrs, err = resolver.LookupHost(aCtx, aHostname); nil != err {
		return nil, err
	}

	ips := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		if ip = net.ParseIP(addr); nil != ip {
			ips = append(ips, ip)
		}
	}

	return ips, err
} // lookupDNS()

/* _EoF_ */
