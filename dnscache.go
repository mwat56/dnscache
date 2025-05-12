/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package dnscache

// Package dnscache caches DNS lookups

import (
	"errors"
	"maps"
	"net"
	"sync"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `tCacheList` is a map of DNS cache entries.
	tCacheList map[string][]net.IP

	// `TResolver` is a DNS resolver with an optional background refresh.
	//
	// It embeds a map of DNS cache entries to store the DNS cache entries
	// and uses a `sync.RWMutex` to synchronise access to the cache.
	TResolver struct {
		mtx   sync.RWMutex
		abort chan struct{} // signal to abort `autoRefresh()`
		tCacheList
	}
)

// `ErrNoIps` is returned when no IP addresses are found for a hostname.
var ErrNoIps = errors.New("no IPs found")

// ---------------------------------------------------------------------------
// constructor function:

// `New()` returns a new DNS resolver with an optional background refresh.
//
// If `aRefreshRate` is greater than zero, cached DNS entries will be
// automatically refreshed at the specified interval.
//
// Parameters:
//   - `aRefreshRate`: Optional interval in minutes to refresh the cache.
//
// Returns:
//   - `*Resolver`: A new `Resolver` instance.
func New(aRefreshRate uint8) *TResolver {
	result := &TResolver{
		abort:      make(chan struct{}), // signal to abort `autoRefresh()`
		tCacheList: make(tCacheList, 64),
	}

	if 0 < aRefreshRate {
		go result.autoRefresh(time.Duration(aRefreshRate) * time.Minute)
	}

	return result
} // New()

// ---------------------------------------------------------------------------
// `Resolver` methods:

// `autoRefresh()` refreshes the cache at a given interval.
//
// Parameters:
//   - `aRate`: Time interval to refresh the cache.
func (r *TResolver) autoRefresh(aRate time.Duration) {
	ticker := time.NewTicker(aRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.Refresh()

		case <-r.abort:
			return
		}
	}
} // autoRefresh()

// `Fetch()` returns the IP addresses for a given hostname.
//
// Parameters:
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `[]net.IP`: List of IP addresses for the given hostname.
//   - `error`: `nil` if the hostname was resolved successfully, the error otherwise.
func (r *TResolver) Fetch(aHostname string) ([]net.IP, error) {
	r.mtx.RLock()
	ips, ok := r.tCacheList[aHostname]
	r.mtx.RUnlock()

	if ok && (0 < len(ips)) {
		// fast path: we've already resolved this hostname
		return ips, nil
	}

	// slow path: we need to resolve this hostname
	return r.Lookup(aHostname)
} // Fetch()

// `FetchOne()` returns the first IP address for a given hostname.
//
// If the hostname has multiple IP addresses, the first one is returned.
//
// Parameters:
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `net.IP`: First IP address for the given hostname.
//   - `error`: `nil` if the hostname was resolved successfully, the error otherwise.
func (r *TResolver) FetchOne(aHostname string) (net.IP, error) {
	ips, err := r.Fetch(aHostname)
	if nil != err {
		return nil, err
	}

	return ips[0], nil
} // FetchOne()

// `FetchOneString()` returns the first IP address for a given hostname
// as a string.
//
// Parameters:
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `string`: First IP address for the given hostname as a string.
//   - `error`: `nil` if the hostname was resolved successfully, the error otherwise.
func (r *TResolver) FetchOneString(aHostname string) (string, error) {
	ip, err := r.FetchOne(aHostname)
	if nil != err {
		return "", err
	}

	return ip.String(), nil
} // FetchOneString()

// `Lookup()` resolves a hostname and caches the result.
//
// Parameters:
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `[]net.IP`: List of IP addresses for the given hostname.
//   - `error`: `nil` if the hostname was resolved successfully, the error otherwise.
func (r *TResolver) Lookup(aHostname string) ([]net.IP, error) {
	ips, err := net.LookupIP(aHostname)
	if nil != err {
		return nil, err
	}
	if 0 == len(ips) {
		return nil, ErrNoIps
	}

	r.mtx.Lock()
	r.tCacheList[aHostname] = ips
	r.mtx.Unlock()

	return ips, nil
} // Lookup()

// `Refresh()` resolves all cached hostnames and updates the cache.
//
// This method is called by automatically if a refresh rate was
// specified in the `New()` constructor.
func (r *TResolver) Refresh() {
	r.mtx.RLock()
	// This is a shallow clone, the new keys and values
	// are set using ordinary assignment:
	hosts := maps.Clone(r.tCacheList)
	r.mtx.RUnlock()

	for host := range hosts {
		_, _ = r.Lookup(host) //#nosec G104
		time.Sleep(time.Second * 2)
	}
} // Refresh()

/* _EoF_ */
