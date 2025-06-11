/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"context"
	"fmt"
	"maps"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	//
	// `tCacheList` is a map of DNS cache entries
	// indexed by lowercased hostnames.
	tCacheList struct {
		sync.RWMutex
		Cache map[string]*tCacheEntry
	}
)

// ---------------------------------------------------------------------------
// `tCacheList` constructor:

// `newMap()` returns a new IP address cache list.
//
// If `aSize` is zero, the default size (`1024`) is used.
//
// Parameters:
//   - `aSize`: Initial size of the cache list.
//
// Returns:
//   - `*tCacheList`: A new IP address cache list.
func newMap(aSize uint) *tCacheList {
	if 0 == aSize {
		aSize = DefaultCacheSize
	}

	return &tCacheList{
		Cache: make(map[string]*tCacheEntry, aSize),
	}
} // newMap()

// ---------------------------------------------------------------------------

// `init()` ensures proper interface implementation.
func init() {
	var (
		_ ICacheList = (*tCacheList)(nil)
	)
} // init()

// ---------------------------------------------------------------------------
// `tCacheList` methods:

// `AutoExpire()` removes expired cache entries at a given interval.
//
// Parameters:
//   - `aRate`: Time interval to refresh the cache.
//   - `aAbort`: Channel to receive a signal to abort.
func (cl *tCacheList) AutoExpire(aRate time.Duration, aAbort chan struct{}) {
	ticker := time.NewTicker(aRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cl.expireEntries()

		case <-aAbort:
			return

		default:
			runtime.Gosched() // yield to other goroutines
		}
	}
} // AutoExpire()

// `Clone()` creates a deep copy of the cache list.
//
// Returns:
//   - `ICacheList`: A deep copy of the cache list.
func (cl *tCacheList) Clone() ICacheList {
	if nil == cl {
		return nil
	}

	clone := &tCacheList{
		Cache: make(map[string]*tCacheEntry, len(cl.Cache)),
	}
	cl.RLock()
	for host, ce := range cl.Cache {
		clone.Cache[host] = ce.clone()
	}
	cl.RUnlock()

	return clone
} // Clone()

// `Create()` adds a new cache entry for the given hostname.
//
// If the given IP list is empty, the cache entry's IP list is cleared/removed.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aHostname`: The hostname to add a cache entry for.
//   - `aIPs`: List of IP addresses to add to the cache entry.
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `ICacheList`: The updated cache list.
func (cl *tCacheList) Create(aCtx context.Context, aHostname string, aIPs []net.IP, aTTL time.Duration) ICacheList {
	if nil == cl {
		return nil
	}

	return cl.Update(aCtx, aHostname, aIPs, aTTL)
} // Create()

// `Delete()` removes the cache entry for the given hostname.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aHostname`: The hostname to remove the cache entry for.
//
// Returns:
//   - `bool`: `true` if the cache entry was found and deleted, `false` otherwise.
func (cl *tCacheList) Delete(aCtx context.Context, aHostname string) (rOK bool) {
	if nil == cl {
		return
	}
	if aHostname = strings.TrimSpace(aHostname); 0 == len(aHostname) {
		return
	}
	aHostname = strings.ToLower(aHostname)

	cl.Lock()
	if ce, ok := cl.Cache[aHostname]; ok {
		rOK = ce.Delete(aCtx, nil)
		//TODO: return `ce` to pool
		delete(cl.Cache, aHostname)
	}
	cl.Unlock()

	return
} // Delete()

// `Equal()` checks whether the cache list is equal to the given one.
//
// Parameters:
//   - `aList`: Cache list to compare with.
//
// Returns:
//   - `bool`: `true` if the cache list is equal to the given one, `false` otherwise.
func (cl *tCacheList) Equal(aList *tCacheList) bool {
	if nil == cl {
		return nil == aList
	}
	if nil == aList {
		return false
	}
	if len(cl.Cache) != len(aList.Cache) {
		return false
	}
	var (
		otherEntry *tCacheEntry
		ok         bool = true
	)

	cl.RLock()
	aList.RLock()
	// Check whether all entries in `cl` are also in `aList`
	for host, myEntry := range cl.Cache {
		if otherEntry, ok = aList.Cache[host]; !ok {
			break
		}
		if ok = myEntry.Equal(otherEntry); !ok {
			break
		}
	}
	aList.RUnlock()
	cl.RUnlock()

	return ok
} // Equal()

// `Exists()` checks whether the given hostname is cached.
//
// Parameters:
//   - `context.Context`: Timeout context to use for the operation.
//   - `string`: The hostname to check for.
//
// Returns:
//   - `bool`: `true` if the hostname was found in the cache, `false` otherwise.
func (cl *tCacheList) Exists(aCtx context.Context, aHostname string) (rOK bool) {
	if nil == cl {
		return
	}
	if aHostname = strings.TrimSpace(aHostname); 0 == len(aHostname) {
		return
	}
	aHostname = strings.ToLower(aHostname)

	cl.RLock()
	_, rOK = cl.Cache[aHostname]
	cl.RUnlock()

	return
} // Exists()

// `expireEntries()` removes all expired cache entries.
//
// This method is called automatically by the `AutoExpire()` method.
func (cl *tCacheList) expireEntries() {
	if nil == cl {
		return
	}

	clone := maps.Clone(cl.Cache)
	for hostname, ce := range clone {
		if ce.isExpired() {
			cl.Lock()
			delete(cl.Cache, hostname)
			cl.Unlock()
		}
	}
	clone = nil
} // expireEntries()

// `IPs()` returns the IP addresses for the given hostname.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `TIpList`: List of IP addresses for the given hostname.
//   - `bool`: `true` if the hostname was found in the cache, `false` otherwise.
func (cl *tCacheList) IPs(aCtx context.Context, aHostname string) ([]net.IP, bool) {
	if (nil == cl) || (0 == len(cl.Cache)) {
		return nil, false
	}
	if aHostname = strings.TrimSpace(aHostname); 0 == len(aHostname) {
		return nil, false
	}
	aHostname = strings.ToLower(aHostname)
	var ips []net.IP

	cl.RLock()
	if ce, ok := cl.Cache[aHostname]; ok {
		ips = make([]net.IP, len(ce.ips))
		copy(ips, ce.ips)
	}
	cl.RUnlock()

	return ips, (0 < len(ips))
} // IPs()

// `Len()` returns the number of cached hostnames.
//
// Returns:
//   - `int`: Number of cached hostnames.
func (cl *tCacheList) Len() int {
	if nil == cl {
		return 0
	}

	return len(cl.Cache)
} // Len()

// `Range()` returns a channel that yields all FQDNs in sorted order.
//
// Usage: for fqdn := range ICacheList.Range() { ... }
//
// The channel is closed automatically when all entries have been yielded.
//
// Parameters:
//   - `aCtx`: Timeout context to use for the operation.
//
// Returns:
//   - `chan string`: Channel that yields all FQDNs in sorted order.
func (cl *tCacheList) Range(aCtx context.Context) <-chan string {
	ch := make(chan string)
	if nil == cl {
		close(ch)
		return ch
	}

	// Collect all hostnames
	cl.RLock()
	hostnames := make([]string, 0, len(cl.Cache))
	for fqdn := range cl.Cache {
		hostnames = append(hostnames, fqdn)
	}
	cl.RUnlock()

	sortHostnames(hostnames)

	go func(aHostList []string) {
		defer close(ch)

		// Send sorted hostnames through channel
		for _, fqdn := range aHostList {
			if nil != aCtx.Err() {
				// Leaving the goroutine will close the
				// channel (due to `defer`).
				return
			}
			ch <- fqdn
		}
	}(hostnames)

	return ch
} // Range()

// `String()` implements the `fmt.Stringer` interface for a string
// representation of the cache list.
//
// Returns:
//   - `string`: String representation of the cache list.
func (cl *tCacheList) String() string {
	if nil == cl {
		return ""
	}
	var builder strings.Builder

	cl.RLock()
	for host, ce := range cl.Cache {
		fmt.Fprintf(&builder, "%s: %s\n", host, ce.String())
	}
	cl.RUnlock()

	return builder.String()
} // String()

// `Update()` updates the cache entry for the given hostname.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aHostname`: The hostname to update the cache entry for.
//   - `aIPs`: List of IP addresses to update the cache entry with.
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `ICacheList`: The updated cache list.
func (cl *tCacheList) Update(aCtx context.Context, aHostname string, aIPs []net.IP, aTTL time.Duration) ICacheList {
	if nil == cl {
		return nil
	}

	if aHostname = strings.TrimSpace(aHostname); 0 == len(aHostname) {
		return cl
	}
	aHostname = strings.ToLower(aHostname)

	if 0 == len(cl.Cache) {
		cl.Cache = make(map[string]*tCacheEntry, DefaultCacheSize)
	}

	if nil != aCtx.Err() {
		return cl
	}

	ce := &tCacheEntry{}
	cl.Lock()
	cl.Cache[aHostname] = ce.Update(aCtx, aIPs, aTTL).(*tCacheEntry)
	cl.Unlock()

	return cl
} // Update()

/* _EoF_ */
