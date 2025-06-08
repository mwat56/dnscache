/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"fmt"
	"maps"
	"runtime"
	"strings"
	"sync"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// `DefaultCacheSize` is the initial size of the cache list.
	DefaultCacheSize = 1 << 10 // 1024
)

type (
	// `TCacheList` is a map of DNS cache entries
	// indexed by lowercased hostnames.
	TCacheList struct {
		sync.RWMutex
		Cache map[string]*TCacheEntry
	}
)

// ---------------------------------------------------------------------------
// `TCacheList` constructor:

// `New()` returns a new IP address cache list.
//
// If `aSize` is zero, the default size (`1024`) is used.
//
// Parameters:
//   - `aSize`: Initial size of the cache list.
//
// Returns:
//   - `*TCacheList`: A new IP address cache list.
func New(aSize uint) *TCacheList {
	if 0 == aSize {
		aSize = DefaultCacheSize
	}

	return &TCacheList{
		Cache: make(map[string]*TCacheEntry, aSize),
	}
} // New()

// ---------------------------------------------------------------------------
// `TCacheList` methods:

// `AutoExpire()` removes expired cache entries at a given interval.
//
// Parameters:
//   - `aRate`: Time interval to refresh the cache.
//   - `aAbort`: Channel to receive a signal to abort.
func (cl *TCacheList) AutoExpire(aRate time.Duration, aAbort chan struct{}) {
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

// `Clone()` returns a deep copy of the cache list.
//
// Returns:
//   - `*TCacheList`: A deep copy of the cache list.
func (cl *TCacheList) Clone() *TCacheList {
	if nil == cl {
		return nil
	}

	result := &TCacheList{
		Cache: make(map[string]*TCacheEntry, len(cl.Cache)),
	}
	cl.RLock()
	for host, ce := range cl.Cache {
		result.Cache[host] = ce.clone()
	}
	cl.RUnlock()

	return result
} // Clone()

// `Delete()` removes the cache entry for the given hostname.
//
// Parameters:
//   - `aHostname`: The hostname to remove the cache entry for.
//
// Returns:
//   - `*TCacheList`: The updated cache list.
func (cl *TCacheList) Delete(aHostname string) *TCacheList {
	if nil == cl {
		return nil
	}
	if aHostname = strings.TrimSpace(aHostname); 0 == len(aHostname) {
		return cl
	}
	aHostname = strings.ToLower(aHostname)

	cl.Lock()
	delete(cl.Cache, aHostname)
	cl.Unlock()

	return cl
} // Delete()

// `Equal()` checks whether the cache list is equal to the given one.
//
// Parameters:
//   - `aList`: Cache list to compare with.
//
// Returns:
//   - `bool`: `true` if the cache list is equal to the given one, `false` otherwise.
func (cl *TCacheList) Equal(aList *TCacheList) bool {
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
		otherEntry *TCacheEntry
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

// `expireEntries()` removes all expired cache entries.
//
// This method is called automatically by the `AutoExpire()` method.
func (cl *TCacheList) expireEntries() {
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

// `GetEntry()` returns the cache entry for the given hostname.
//
// Parameters:
//   - `aHostname`: The hostname to get the cache entry for.
//
// Returns:
//   - `*tCacheEntry`: The cache entry for the given hostname.
//   - `bool`: `true` if the hostname was found in the cache, `false` otherwise.
func (cl *TCacheList) GetEntry(aHostname string) (*TCacheEntry, bool) {
	if nil == cl {
		return nil, false
	}
	if aHostname = strings.TrimSpace(aHostname); 0 == len(aHostname) {
		return nil, false
	}
	aHostname = strings.ToLower(aHostname)

	cl.RLock()
	ce, ok := cl.Cache[strings.ToLower(aHostname)]
	cl.RUnlock()

	return ce, ok
} // GetEntry()

// `IPs()` returns the IP addresses for the given hostname.
//
// Parameters:
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `TIpList`: List of IP addresses for the given hostname.
//   - `bool`: `true` if the hostname was found in the cache, `false` otherwise.
func (cl *TCacheList) IPs(aHostname string) (tIpList, bool) {
	if (nil == cl) || (0 == len(cl.Cache)) {
		return nil, false
	}
	if aHostname = strings.TrimSpace(aHostname); 0 == len(aHostname) {
		return nil, false
	}
	aHostname = strings.ToLower(aHostname)

	cl.RLock()
	ce, ok := cl.Cache[aHostname]
	cl.RUnlock()
	if ok {
		return ce.ips, true
	}

	return nil, false
} // IPs()

// `Len()` returns the number of entries in the cache list.
//
// Returns:
//   - `int`: Number of entries in the cache list.
func (cl *TCacheList) Len() int {
	if nil == cl {
		return 0
	}

	return len(cl.Cache)
} // Len()

// `SetEntry()` adds a new cache entry for the given hostname.
//
// If the given IP list is empty, the cache entry's IP list is cleared/removed.
//
// Parameters:
//   - `aHostname`: The hostname to add a cache entry for.
//   - `aIPs`: List of IP addresses to add to the cache entry.
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `*TCacheList`: The updated cache list.
func (cl *TCacheList) SetEntry(aHostname string, aIPs tIpList, aTTL time.Duration) *TCacheList {
	if nil == cl {
		return cl
	}
	if aHostname = strings.TrimSpace(aHostname); 0 == len(aHostname) {
		return cl
	}
	aHostname = strings.ToLower(aHostname)

	if 0 == len(cl.Cache) {
		cl.Cache = make(map[string]*TCacheEntry, DefaultCacheSize)
	}

	cl.Lock()
	if ce, ok := cl.Cache[aHostname]; ok {
		cl.Cache[aHostname] = ce.Update(aIPs, aTTL)
	} else {
		cl.Cache[aHostname] = newCacheEntry(aTTL).Update(aIPs, aTTL)
	}
	cl.Unlock()

	return cl
} // SetEntry()

// `String()` implements the `fmt.Stringer` interface for a string
// representation of the cache list.
//
// Returns:
//   - `string`: String representation of the cache list.
func (cl *TCacheList) String() string {
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

/* _EoF_ */
