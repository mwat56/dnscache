/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package dnscache

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/mwat56/dnscache/cache"
	adl "github.com/mwat56/dnscache/internal/adlist"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	//
	// `defExpireInterval` is the default interval at which expired
	// cache entries are removed from the cache.
	defExpireInterval = uint8(1 << 4) // 16 minutes

	//
	// `defRetries` is the default number of retries for DNS lookups.
	defRetries = 3

	//
	// `defLookupTimeout` is the default timeout for DNS lookups.
	defLookupTimeout = time.Minute << 1
)

type (
	//
	// `TResolverOptions` contain configuration options for creating a resolver.
	//
	// This are the public fields to configure a new `TResolver` instance:
	//
	//   - `BlockLists`: List of URLs to download blocklists from.
	//   - `DNSservers`: List of DNS servers to use, `nil` means use system default.
	//   - `AllowList`: Path/file name to read the 'allow' patterns from.
	//   - `DataDir`: Directory to store local allow and deny lists.
	//   - `CacheSize`: Initial cache size, `0` means use default (`512`).
	//   - `Resolver`: Custom resolver, `nil` means use default.
	//   - `ExpireInterval`: Optional interval (in minutes) to remove expired cache entries.
	//   - `MaxRetries`: Maximum number of retries for DNS lookup, `0` means use default (`3`).
	//   - `RefreshInterval`: Optional interval (in minutes) to refresh the cache.
	//   - `TTL`: Optional time to live (in minutes) for cache entries.
	TResolverOptions struct {
		BlockLists      []string
		DNSservers      []string
		AllowList       string
		DataDir         string
		CacheSize       int
		Resolver        *net.Resolver
		ExpireInterval  uint8
		MaxRetries      uint8
		RefreshInterval uint8
		TTL             uint8
	}

	//
	// `TResolver` is a DNS resolver with an optional background refresh.
	//
	// It embeds a map of DNS cache entries to store the DNS cache entries
	// and uses a Mutex to synchronise access to that cache.
	TResolver struct {
		sync.RWMutex
		dnsServers       []string
		cache.ICacheList               //list of DNS cache entries
		abortExpire      chan struct{} // signal to abort `autoExpire()`
		abortRefresh     chan struct{} // signal to abort `autoRefresh()`
		adlist           *adl.TADlist  // allow/deny list to check before DNS
		resolver         *net.Resolver // DNS resolver to use
		ttl              time.Duration // TTL for cache entries
		retries          uint8         // max. number of retries for DNS lookups
	}
)

// ---------------------------------------------------------------------------
// Helper functions:

// `validateDNSServers()` validates the given list of DNS server IPs.
//
// Parameters:
//   - `aServerList`: List of DNS server IPs to validate.
//
// Returns:
//   - `[]string`: List of valid DNS server IPs.
func validateDNSServers(aServerList []string) []string {
	if 0 == len(aServerList) {
		return nil
	}

	validIPs := make([]string, 0, len(aServerList))
	for _, server := range aServerList {
		if nil != net.ParseIP(server) {
			validIPs = append(validIPs, server)
		}
	}

	return validIPs
} // validateDNSServers()

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
//   - `*TResolver`: A new `TResolver` instance.
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
//   - `*TResolver`: A new `TResolver` instance.
func NewWithOptions(aOptions TResolverOptions) *TResolver {
	optServers := validateDNSServers(aOptions.DNSservers)
	if 0 == len(optServers) {
		// Use system default DNS servers
		var err error
		if optServers, err = getDNSServers(); (nil != err) || (0 == len(optServers)) {
			optServers = nil
		}
	}

	optCacheSize := uint(aOptions.CacheSize) //#nosec G115
	if 0 >= optCacheSize {
		optCacheSize = cache.DefaultCacheSize
	}

	optDataDir := strings.TrimSpace(aOptions.DataDir)
	if 0 == len(optDataDir) {
		optDataDir = os.TempDir()
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
		adlist:       adl.New(optDataDir),
		resolver:     optResolver,
		ICacheList:   cache.New(cache.CacheTypeTrie, optCacheSize),
		retries:      optRetries,
	}

	if optTTL := aOptions.TTL; 0 == optTTL {
		result.ttl = cache.DefaultTTL
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
		go result.ICacheList.AutoExpire(time.Minute*time.Duration(optExpireInterval), result.abortExpire)
		runtime.Gosched() // yield to the new goroutine
	}

	// Load the allow list
	optAllowList := strings.TrimSpace(aOptions.AllowList)
	if 0 < len(optAllowList) {
		if err := result.LoadAllowlist(optAllowList); nil != err {
			// Log the error, but don't fail because of that
			log.Printf("Failed to load allowlist: %v", err)
		}
	}

	// Load the deny list
	if 0 < len(aOptions.BlockLists) {
		if err := result.LoadBlocklists(aOptions.BlockLists); nil != err {
			// Log the error, but don't fail because of that
			log.Printf("Failed to load blocklists: %v", err)
		}
	}

	return result
} // NewWithOptions()

// ---------------------------------------------------------------------------
// `TResolver` methods:

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
	if adl.ADdeny == r.adlist.Match(context.Background(), aHostname) {
		incMetricsFields(&gMetrics.Lookups, &gMetrics.Hits)

		return append([]net.IP{}, net.IPv4zero), nil
	}

	// Use a context with timeout for the entire lookup operation
	ctx, cancel := context.WithTimeout(context.Background(), defLookupTimeout)
	defer cancel()

	// Check the local cache
	r.RLock()
	ips, ok := r.ICacheList.IPs(ctx, aHostname)
	r.RUnlock()

	if ok && (0 < len(ips)) {
		incMetricsFields(&gMetrics.Lookups, &gMetrics.Hits)

		// fast path: we've already resolved this hostname
		return ips, nil
	}
	incMetricsFields(&gMetrics.Misses)

	return r.LookupHost(ctx, aHostname)
} // Fetch()

// `FetchFirst()` returns the first IP address for a given hostname.
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
func (r *TResolver) FetchFirst(aHostname string) (net.IP, error) {
	ips, err := r.Fetch(aHostname)
	if nil != err {
		return nil, err
	}

	return ips[0], nil
} // FetchFirst()

// `FetchFirstString()` returns the first IP address for a given hostname
// as a string.
//
// Parameters:
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `string`: First IP address for the given hostname as a string.
//   - `error`: `nil` if the hostname was resolved successfully, the error otherwise.
func (r *TResolver) FetchFirstString(aHostname string) (string, error) {
	ip, err := r.FetchFirst(aHostname)
	if nil != err {
		return "", err
	}

	return ip.String(), nil
} // FetchFirstString()

// `FetchRandom()` returns a random IP address for a given hostname.
//
// If the hostname has multiple IP addresses, a random one is returned;
// to get always the first one, use [FetchFirst] instead.
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
// to get always the first one, use [FetchFirstString] instead.
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

// `LoadAllowlist()` loads the allowlist from the given file.
//
// Parameters:
//   - `aFilename`: The path/file name to read the 'allow' patterns from.
//
// Returns:
//   - `error`: An error in case of problems, or `nil` otherwise.
func (r *TResolver) LoadAllowlist(aFilename string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second<<2)
	defer cancel()

	return r.adlist.LoadAllow(ctx, aFilename)
} // LoadAllowlist()

// `LoadBlocklists()` loads the blocklists from the given URLs.
//
// Parameters:
//   - `aURLs`: The URLs to download the blocklists from.
//
// Returns:
//   - `error`: An error in case of problems, or `nil` otherwise.
func (r *TResolver) LoadBlocklists(aURLs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second<<2)
	defer cancel()

	return r.adlist.LoadDeny(ctx, aURLs)
} // LoadBlocklists()

// `lookup()` resolves `aHostname` with the given context.
//
// Parameters:
//   - `aCtx`: Context for the lookup operation.
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `[]net.IP`: List of IP addresses for the given hostname.
//   - `error`: `nil` if the hostname was resolved successfully, the error otherwise.
func (r *TResolver) lookup(aCtx context.Context, aHostname string) ([]net.IP, error) {
	if nil != r.dnsServers {
		// Resolve the hostname with multiple DNS servers in parallel
		results := make(chan []net.IP, len(r.dnsServers))

		// Create child context with cancellation control
		ctx, cancel := context.WithCancel(aCtx)
		defer cancel() // Always release resources

		var wg sync.WaitGroup
		for _, server := range r.dnsServers {
			wg.Add(1)
			go func(aServer, aHostname string) {
				defer wg.Done()

				if ips, err := lookupDNS(ctx, aServer, aHostname); nil == err {
					if 0 < len(ips) {
						select {
						case results <- ips:
							// Successfully sent result
						case <-ctx.Done():
							// Context is already canceled, discard result
							return
						}
						// We have a valid result, hence
						// cancel all other lookups
						cancel()
					}
				}
			}(server, aHostname)
		}
		wg.Wait()
		close(results)
		if ips, ok := <-results; ok {
			return ips, nil
		}
	}

	// Reaching this point of execution means that we have no DNS
	// servers configured, or that all of them failed. Hence we
	// fallback to the default resolver.
	ips, err := r.resolver.LookupIP(aCtx, "ip", aHostname)
	if nil == err {
		return ips, nil
	}

	// Check if it's a "not found" DNS error
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
		// No need to retry for a non-existent host
		ips = nil
	}

	return ips, err
} // lookup()

// `LookupHost()` resolves a hostname with the given context and
// caches the result.
//
// This method is called by `Fetch()` and `Refresh()` internally and not
// intended for public use because it bypasses both, the allow/deny lists
// and the internal cache.
//
// Parameters:
//   - `aCtx`: Context for the lookup operation.
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `[]net.IP`: List of IP addresses for the given hostname.
//   - `error`: `nil` if the hostname was resolved successfully, the error otherwise.
func (r *TResolver) LookupHost(aCtx context.Context, aHostname string) ([]net.IP, error) {
	var (
		err error
		ips []net.IP
	)

	// Try to resolve the hostname several times
	for loop := uint8(0); loop < r.retries; loop++ {
		// Check if context is terminated before each attempt
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
	r.Lock()
	r.ICacheList.Create(aCtx, aHostname, ips, r.ttl)
	setMetricsFieldMax(&gMetrics.Peak, uint32(r.ICacheList.Len())) //#nosec G115
	r.Unlock()

	return ips, nil
} // LookupHost()

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
	var dnsErr *net.DNSError

	// Use a context with timeout for the entire refresh operation
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute<<2)
	defer cancel()

	// Use a rate limiter to avoid overwhelming DNS servers
	limiter := time.NewTicker(time.Second << 1)
	defer limiter.Stop()

	r.RLock()
	// This is a shallow clone, the new keys and values
	// are set using ordinary assignment:
	cacheList := r.ICacheList.Clone()
	r.RUnlock()

	for hostname := range cacheList.Range(ctx) {
		select {
		case <-ctx.Done():
			return // Context timeout or cancellation
		case <-limiter.C:
			// Lookup the hostname:
			_, err := r.LookupHost(ctx, hostname)
			if nil != err {
				if errors.As(err, &dnsErr) {
					if dnsErr.IsNotFound {
						// We'e working on a (possibly outdated) copy
						// of the cache, but we delete the non-existing
						// host from our original cache:
						r.Lock()
						r.ICacheList.Delete(ctx, hostname)
						r.Unlock()
					}
				}
			}
			runtime.Gosched() // yield to other goroutines
		}
	}

	//
	//TODO: Reload allow and deny lists
	//
} // Refresh()

// `StopExpire()` stops the background expiration goroutine if it's running.
//
// This method should be called when the background expirations are no
// longer needed. The resolver remains usable after calling `StopExpire()“,
// but cached entries will no longer be automatically expired.
func (r *TResolver) StopExpire() *TResolver {
	select {
	case r.abortExpire <- struct{}{}:
		// Signal sent successfully
		runtime.Gosched()

	default:
		// Channel already closed or no goroutine listening
	}

	return r
} // StopExpire()

// `StopRefresh()` stops the background refresh goroutine if it's running.
//
// This method should be called when the background updates are no
// longer needed. The resolver remains usable after calling `StopRefresh()“,
// but cached entries will no longer be automatically refreshed.
func (r *TResolver) StopRefresh() *TResolver {
	select {
	case r.abortRefresh <- struct{}{}:
		// Signal sent successfully
		runtime.Gosched()

	default:
		// Channel already closed or no goroutine listening
	}

	return r

	// Note: We don't clear the cache here as the resolver
	// remains usable, and cached entries are still valid
} // StopRefresh()

/* _EoF_ */
