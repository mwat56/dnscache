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
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// `DefaultCacheSize` is the initial size of the cache list.
	DefaultCacheSize = 1 << 9 // 512
)

type (
	// `TCacheList` is a map of DNS cache entries
	// indexed by the hostname.
	TCacheList map[string]*TCacheEntry
)

// ---------------------------------------------------------------------------
// `TCacheList` constructor:

// `New()` returns a new IP address cache list.
//
// If `aSize` is zero, the default size (`512`) is used.
//
// Parameters:
//   - `aSize`: Initial size of the cache list.
//
// Returns:
//   - `TCacheList`: A new IP address cache list.
func New(aSize uint) TCacheList {
	if 0 == aSize {
		aSize = DefaultCacheSize
	}

	return make(TCacheList, aSize)
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

	result := make(TCacheList, len(*cl))
	for host, ce := range *cl {
		result[host] = ce.clone()
	}

	return &result
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

	delete(*cl, aHostname)

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
	if len(*cl) != len(*aList) {
		return false
	}
	var (
		otherE *TCacheEntry
		ok     bool
	)

	// Check whether all entries in `cl` are also in `aList`
	for host, myE := range *cl {
		otherE, ok = (*aList)[host]
		if !ok {
			return false
		}
		if !myE.Equal(otherE) {
			return false
		}
	}

	return true
} // Equal()

// `expireEntries()` removes all expired cache entries.
//
// This method is called automatically by the `AutoExpire()` method.
func (cl *TCacheList) expireEntries() {
	if nil == cl {
		return
	}

	clone := maps.Clone(*cl)
	for hostname, ce := range clone {
		if ce.isExpired() {
			delete(*cl, hostname)
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

	if ce, ok := (*cl)[aHostname]; ok {
		return ce, true
	}

	return nil, false
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
	if (nil == cl) || (0 == len(*cl)) {
		return nil, false
	}

	if ce, ok := (*cl)[aHostname]; ok {
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

	return len(*cl)
} // Len()

// `SetEntry()` adds a new cache entry for the given hostname.
//
// Parameters:
//   - `aHostname`: The hostname to add a cache entry for.
//   - `aIPs`: List of IP addresses to add to the cache entry.
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `*TCacheList`: The updated cache list.
func (cl *TCacheList) SetEntry(aHostname string, aIPs tIpList, aTTL time.Duration) *TCacheList {
	if (nil == cl) || (nil == aIPs) || (0 == len(aIPs)) {
		return cl
	}

	if ce, ok := (*cl)[aHostname]; ok {
		(*cl)[aHostname] = ce.update(aIPs, aTTL)
	} else {
		(*cl)[aHostname] = newCacheEntry(aTTL).update(aIPs, aTTL)
	}

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
	for host, ce := range *cl {
		fmt.Fprintf(&builder, "%s: %s\n", host, ce.String())
	}

	return builder.String()
} // String()

/* _EoF_ */
