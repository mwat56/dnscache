/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package dnscache

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"runtime"
	"slices"
	"sync"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// `defCacheSize` is the default size of the cache list.
	defCacheSize = 64

	// `defExpireInterval` is the default interval at which expired cache entries are removed from the cache.
	defExpireInterval = uint8(1 << 4) // 16 minutes

	// `defRetries` is the default number of retries for DNS lookups.
	defRetries = 3

	// `defTTL` is the default time to live for a DNS cache entry.
	defTTL = time.Duration(time.Minute << 6) // 64 minutes
)

type (
	// `TResolverOptions` contain configuration options for creating a resolver.
	//
	// This are the public fields to configure a new `TResolver` instance:
	//
	//   - `DNSservers`: List of DNS servers to use, `nil` means use system default.
	//   - `CacheSize`: Initial cache size, `0` means use default (`64`).
	//   - `Resolver`: Custom resolver, `nil` means use default.
	//   - `ExpireInterval`: Optional interval (in minutes) to remove expired cache entries.
	//   - `MaxRetries`: Maximum number of retries for DNS lookup, `0` means use default (`3`).
	//   - `RefreshInterval`: Optional interval (in minutes) to refresh the cache.
	//   - `TTL`: Optional time to live (in minutes) for cache entries.
	TResolverOptions struct {
		DNSservers      []string
		CacheSize       int
		Resolver        *net.Resolver
		ExpireInterval  uint8
		MaxRetries      uint8
		RefreshInterval uint8
		TTL             uint8
	}

	// `TResolver` is a DNS resolver with an optional background refresh.
	//
	// It embeds a map of DNS cache entries to store the DNS cache entries
	// and uses a Mutex to synchronise access to that cache.
	TResolver struct {
		mtx          sync.RWMutex
		dnsServers   []string
		abortExpire  chan struct{} // signal to abort `autoExpire()`
		abortRefresh chan struct{} // signal to abort `autoRefresh()`
		resolver     *net.Resolver // DNS resolver to use
		tCacheList                 //list of DNS cache entries
		ttl          time.Duration // TTL for cache entries
		retries      uint8         // max. number of retries for DNS lookups
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
	optServers := aOptions.DNSservers
	if 0 == len(optServers) {
		// Use system default DNS servers
		var err error
		if optServers, err = getDNSServers(); (nil != err) ||
			(0 == len(optServers)) {
			optServers = nil
		}
	} else {
		// Check whether the provided DNS servers are valid,
		// and remove invalid entries from the list
		c := slices.Clone(optServers)
		for i, server := range c {
			if nil == net.ParseIP(server) {
				// Remove invalid server from list
				optServers = slices.Delete(optServers, i, i+1)
			}
		}
		c = nil // free memory

		if 0 == len(optServers) {
			// Use system default DNS servers because
			// the provided list was invalid
			var err error
			if optServers, err = getDNSServers(); (nil != err) ||
				(0 == len(optServers)) {
				optServers = nil
			}
		}
	}

	optCacheSize := aOptions.CacheSize
	if 0 >= optCacheSize {
		optCacheSize = defCacheSize
	}

	optResolver := aOptions.Resolver
	if nil == optResolver {
		optResolver = net.DefaultResolver
	}

	optRetries := aOptions.MaxRetries
	if 0 == optRetries {
		optRetries = defRetries
	}

	result := &TResolver{
		dnsServers:   optServers,
		abortExpire:  make(chan struct{}),
		abortRefresh: make(chan struct{}),
		resolver:     optResolver,
		tCacheList:   make(tCacheList, optCacheSize),
		retries:      optRetries,
	}

	if optTTL := aOptions.TTL; 0 == optTTL {
		result.ttl = defTTL
	} else {
		result.ttl = time.Minute * time.Duration(optTTL)
	}

	if 0 < aOptions.RefreshInterval {
		// Start the auto-refresh goroutine.
		go result.autoRefresh(time.Minute*time.Duration(aOptions.RefreshInterval), result.abortRefresh)
		runtime.Gosched() // yield to the new goroutine
	}

	optExpireInterval := aOptions.ExpireInterval
	if 0 == optExpireInterval {
		optExpireInterval = defExpireInterval
	}
	if 0 < optExpireInterval {
		// Start the auto-expire goroutine.
		go result.tCacheList.autoExpire(time.Minute*time.Duration(optExpireInterval), result.abortExpire)
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
//   - `aAbort`: Channel to receive a signal to abort.
func (r *TResolver) autoRefresh(aRate time.Duration, aAbort chan struct{}) {
	ticker := time.NewTicker(aRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.Refresh()

		case <-aAbort:
			return

		default:
			runtime.Gosched() // yield to other goroutines
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
	ips, ok := r.tCacheList.ips(aHostname)
	r.mtx.RUnlock()

	if ok && (0 < len(ips)) {
		incMetricsFields(&gMetrics.Lookups, &gMetrics.Hits)

		// fast path: we've already resolved this hostname
		return ips, nil
	}
	incMetricsFields(&gMetrics.Misses)

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
		idx = rand.Intn(len(ips)) //#nosec G404
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

// `lookup()` resolves `aHostname` with the given context.
//
// Parameters:
//   - `aCtx`: Context for the lookup operation.
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `rIPs`: List of IP addresses for the given hostname.
//   - `rErr`: `nil` if the hostname was resolved successfully, the error otherwise.
func (r *TResolver) lookup(aCtx context.Context, aHostname string) (rIPs []net.IP, rErr error) {
	if nil != r.dnsServers {
		for _, server := range r.dnsServers {
			if rIPs, rErr = lookupDNS(aCtx, server, aHostname); nil == rErr {
				return // lookup succeeded
			}
		}
	}

	if rIPs, rErr = r.resolver.LookupIP(aCtx, "ip", aHostname); nil == rErr {
		return // lookup succeeded
	}

	// Check if it's a "not found" DNS error
	var dnsErr *net.DNSError
	if errors.As(rErr, &dnsErr) && dnsErr.IsNotFound {
		// No need to retry for a non-existent host
		rIPs = nil
	}

	return
} // lookup()

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
	var (
		err error
		ips []net.IP
	)

	// Try to resolve the hostname several times
	for loop := uint8(0); loop < r.retries; loop++ {
		// Check if context is done before each attempt
		select {
		case <-aCtx.Done():
			// No metrics data to update yet
			return nil, aCtx.Err()

		default:
			// Continue with lookup
		}

		if ips, err = r.lookup(aCtx, aHostname); nil == err {
			// Update metrics
			if 0 < loop {
				incMetricsFields(&gMetrics.Retries)
			}
			break // lookup succeeded
		}

		// Short delay before retry, respecting context
		select {
		case <-aCtx.Done():
			if 0 < loop {
				incMetricsFields(&gMetrics.Retries)
			}
			return nil, aCtx.Err()

		default:
			runtime.Gosched() // yield to other goroutines
		}
	} // for loop

	if nil != err {
		incMetricsFields(&gMetrics.Lookups, &gMetrics.Errors)
		return nil, err
	}

	// Update metrics
	incMetricsFields(&gMetrics.Lookups)

	// Cache the result
	r.mtx.Lock()
	r.tCacheList.setEntry(aHostname, ips, r.ttl)
	setMetricsFieldMax(&gMetrics.Peak, uint32(r.tCacheList.len())) //#nosec G115
	r.mtx.Unlock()

	return ips, nil
} // Lookup()

// `Metrics()` returns the current metrics data.
//
// Returns:
//   - `*TMetrics`: Current metrics data.
func (r *TResolver) Metrics() (rMetrics *TMetrics) {
	return gMetrics.clone()
} // Metrics()

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
	hosts := r.tCacheList.clone()
	r.mtx.RUnlock()

	for hostname := range *hosts {
		select {
		case <-ctx.Done():
			return // Context timeout or cancellation
		case <-limiter.C:
			// Lookup the hostname:
			_, err := r.Lookup(ctx, hostname)
			if nil != err {
				if errors.As(err, &dnsErr) {
					if dnsErr.IsNotFound {
						// We'e working on a (possibly outdated) copy
						// of the cache, but we delete the non-existing
						// host from our original cache:
						r.mtx.Lock()
						r.tCacheList.delete(hostname)
						r.mtx.Unlock()
					}
				}
			}
			runtime.Gosched() // yield to other goroutines
		}
	}
} // Refresh()

// `StopExpire()` stops the background expiration goroutine if it's running.
//
// This method should be called when the background expirations are no
// longer needed. The resolver remains usable after calling `StopExpire()“,
// but cached entries will no longer be automatically expired.
func (r *TResolver) StopExpire() {
	select {
	case r.abortExpire <- struct{}{}:
		// Signal sent successfully
		runtime.Gosched()
	default:
		// Channel already closed or no goroutine listening
		return
	}
} // StopExpire()

// `StopRefresh()` stops the background refresh goroutine if it's running.
//
// This method should be called when the background updates are no
// longer needed. The resolver remains usable after calling `StopRefresh()“,
// but cached entries will no longer be automatically refreshed.
func (r *TResolver) StopRefresh() {
	select {
	case r.abortRefresh <- struct{}{}:
		// Signal sent successfully
		runtime.Gosched()
	default:
		// Channel already closed or no goroutine listening
		return
	}

	// Note: We don't clear the cache here as the resolver
	// remains usable, and cached entries are still valid
} // StopRefresh()

/* _EoF_ */
