/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package dnscache

import (
	"context"
	"errors"
	"maps"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `tCacheList` is a map of DNS cache entries.
	tCacheList map[string][]net.IP

	// `TResolverOptions` contains options for creating a resolver.
	//
	// This are the public fields to configure a new `TResolver` instance:
	//
	//   - `RefreshInterval`: Optional interval in minutes to refresh the cache.
	//   - `Resolver`: Custom resolver, nil means use default.
	//   - `CacheSize`: Initial cache size, 0 means use default (64).
	TResolverOptions struct {
		RefreshInterval uint8
		Resolver        *net.Resolver
		CacheSize       int
	}

	// `TResolver` is a DNS resolver with an optional background refresh.
	//
	// It embeds a map of DNS cache entries to store the DNS cache entries
	// and uses a `sync.RWMutex` to synchronise access to the cache.
	TResolver struct {
		mtx      sync.RWMutex
		abort    chan struct{} // signal to abort `autoRefresh()`
		resolver *net.Resolver
		tCacheList
	}
)

// ---------------------------------------------------------------------------
// Constructor functions:

// `New()` returns a new DNS resolver with an optional background refresh.
//
// If `aRefreshInterval` is greater than zero, cached DNS entries will be
// automatically refreshed at the specified interval.
//
// Parameters:
//   - `aRefreshInterval`: Optional interval in minutes to refresh the cache.
//
// Returns:
//   - `*Resolver`: A new `Resolver` instance.
func New(aRefreshInterval uint8) *TResolver {
	return NewWithOptions(TResolverOptions{
		RefreshInterval: aRefreshInterval,
		Resolver:        nil, // use default
		CacheSize:       64,  // default size
	})
} // New()

// `NewWithOptions()` returns a new DNS resolver with custom options.
//
// Parameters:
//   - `aOptions`: Options for the resolver.
//
// Returns:
//   - `*Resolver`: A new `Resolver` instance.
func NewWithOptions(aOptions TResolverOptions) *TResolver {
	cacheSize := aOptions.CacheSize
	if 0 >= cacheSize {
		cacheSize = 64
	}

	resolver := aOptions.Resolver
	if nil == resolver {
		resolver = net.DefaultResolver
	}

	result := &TResolver{
		abort:      make(chan struct{}),
		resolver:   resolver,
		tCacheList: make(tCacheList, cacheSize),
	}

	if 0 < aOptions.RefreshInterval {
		go result.autoRefresh(time.Minute * time.Duration(aOptions.RefreshInterval))
		runtime.Gosched() // yield to the new goroutine
	}

	return result
} // NewWithOptions()

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

// `Close()` stops the background refresh goroutine if it's running.
//
// This method should be called when the background updates are no
// longer needed. The resolver remains usable after calling Close(),
// but cached entries will no longer be automatically refreshed.
func (r *TResolver) Close() {
	select {
	case r.abort <- struct{}{}:
		// Signal sent successfully
		runtime.Gosched()
	default:
		// Channel already closed or no goroutine listening
		return
	}

	// Note: We don't clear the cache here as the resolver
	// remains usable, and cached entries are still valid
} // Close()

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

	// Use a context with timeout for the entire refresh operation
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	ips, err := r.Lookup(ctx, aHostname)
	if nil != err {
		return nil, err
	}

	// slow path: we need to resolve this hostname
	return ips, err
} // Fetch()

// `FetchOne()` returns the first IP address for a given hostname.
//
// If the hostname has multiple IP addresses, the first one is returned;
// to get a random IP address, use [FetchRandom].
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

// `FetchRandom()` returns a random IP address for a given hostname.
//
// If the hostname has multiple IP addresses, a random one is returned;
// to get always the first one, use [FetchOne] instead.
//
// Parameters:
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `net.IP`: Random IP address for the given hostname.
//   - `error`: `nil` if the hostname was resolved successfully, the error otherwise.
func (r *TResolver) FetchRandom(aHostname string) (net.IP, error) {
	ips, err := r.Fetch(aHostname)
	if (nil != err) || (nil == ips) {
		return nil, err
	}

	idx := 0
	if 0 < len(ips) {
		idx = rand.Intn(len(ips))
	}

	return ips[idx], nil
} // FetchRandom()

// `FetchRandomString()` returns a random IP address for a given hostname
// as a string.
//
// If the hostname has multiple IP addresses, a random one is returned;
// to get always the first one, use [FetchOneString] instead.
//
// Parameters:
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `string`: Random IP address for the given hostname as a string.
//   - `error`: `nil` if the hostname was resolved successfully, the error otherwise.
func (r *TResolver) FetchRandomString(aHostname string) (string, error) {
	ip, err := r.FetchRandom(aHostname)
	if nil != err {
		return "", err
	}

	return ip.String(), nil
} // FetchRandomString()

// `Lookup()` resolves a hostname with the given context and caches the result.
//
// Parameters:
//   - `aCtx`: Context for the lookup operation.
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `[]net.IP`: List of IP addresses for the given hostname.
//   - `error`: `nil` if the hostname was resolved successfully, the error otherwise.
func (r *TResolver) Lookup(aCtx context.Context, aHostname string) ([]net.IP, error) {
	ips, err := r.resolver.LookupIP(aCtx, "ip", aHostname)
	if nil != err {
		return nil, err
	}

	// Cache the result
	r.mtx.Lock()
	r.tCacheList[aHostname] = ips
	r.mtx.Unlock()

	return ips, nil
} // Lookup()

// `Refresh()` resolves all cached hostnames and updates the cache.
//
// This method is called automatically if a refresh interval was
// specified in the `New()` constructor.
func (r *TResolver) Refresh() {
	// Use a context with timeout for the entire refresh operation
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	// Use a rate limiter to avoid overwhelming DNS servers
	limiter := time.NewTicker(time.Second << 1)
	defer limiter.Stop()

	var dnsErr *net.DNSError

	r.mtx.RLock()
	// This is a shallow clone, the new keys and values
	// are set using ordinary assignment:
	hosts := maps.Clone(r.tCacheList)
	r.mtx.RUnlock()

	for host := range hosts {
		// Try to resolve each hostname up to 3 times:
		for loop := 0; loop < 3; {
			select {
			case <-ctx.Done():
				return // Context timeout or cancellation
			case <-limiter.C:
				// Lookup the hostname:
				_, err := r.Lookup(ctx, host)
				if nil != err {
					if errors.As(err, &dnsErr) {
						if dnsErr.IsNotFound {
							// We'e working on a (possibly outdated) copy
							// of the cache, but we delete the non-existing
							// host from our original cache:
							r.mtx.Lock()
							delete(r.tCacheList, host)
							r.mtx.Unlock()
							loop = 254 // host not found: Break retry loop
						}
					}
					loop++ // Continue retry loop
				} else {
					loop = 255 // Lookup succeeded: Break retry loop
				}
				runtime.Gosched() // yield to other goroutines
			}
		}
	}
} // Refresh()

/* _EoF_ */
