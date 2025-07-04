/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// `defTimeFormat` is the default time format for the cache entries'
	// string representation of the `bestBefore` field.
	defTimeFormat = "2006-01-02 15:04:05.999999999"
)

type (
	//
	// `tMapEntry` is a DNS cache entry.
	tMapEntry struct {
		ips        tIpList   // IP addresses for this entry
		bestBefore time.Time // time after which the entry is not valid
	}
)

// ---------------------------------------------------------------------------

// `init()` ensures proper interface implementation.
func init() {
	var (
		_ iCacheNode = (*tMapEntry)(nil)
	)
} // init()

// ---------------------------------------------------------------------------
// `tMapEntry` methods:

// `clone()` creates a deep copy of the cache entry.
//
// Returns:
//   - `*tMapEntry`: A deep copy of the cache entry.
func (ce *tMapEntry) clone() *tMapEntry {
	if nil == ce {
		return nil
	}

	clone := newMapEntry()
	clone.bestBefore = ce.bestBefore

	if iLen := len(ce.ips); 0 < iLen {
		clone.ips = make(tIpList, iLen)
		copy(clone.ips, ce.ips)
	}

	return clone
} // clone()

// `Create()` creates a cache entry with the given IP addresses and TTL.
//
// If the cache entry already exists, it is updated with the given IP
// addresses and TTL.
//
// NOTE: This method's implementation ignores both, the `aCtx` and
// `aPartsList` arguments required by the `iCacheNode` interface.
//
// Parameters:
//   - `context.Context`: The timeout context to use for the operation.
//   - `tPartsList`: The list of parts of the hostname to use.
//   - `tIpList`: List of IP addresses to store with the cache entry.
//   - `time.Duration`: Time to live for the cache entry.
//
// Returns:
//   - `bool`: `true` if the cache entry was created, `false` otherwise.
func (ce *tMapEntry) Create(aCtx context.Context, aPartsList tPartsList, aIPs tIpList, aTTL time.Duration) bool {
	if nil == ce {
		ce = newMapEntry()
	}
	ce = ce.Update(aCtx, aIPs, aTTL).(*tMapEntry)

	return (nil != ce)
} // Create()

// `Delete()` clears the cache entry by zeroing the IP list and the
// expiration time.
//
// NOTE: This method's implementation ignores both, the `Context` and
// `tPartsList` arguments required by the `iCacheNode` interface.`
//
// Parameters:
//   - `context.Context`: The timeout context to use for the operation.
//   - `tPartsList`: List of parts of the hostname pattern to delete.
//
// Returns:
//   - `bool`: `true` if a node was deleted, `false` otherwise.
func (ce *tMapEntry) Delete(context.Context, tPartsList) bool {
	if nil != ce {
		ce.ips = tIpList{}
		ce.bestBefore = time.Time{}
	}

	return true
} // Delete()

// `Equal()` checks whether the cache entry is equal to the given one.
//
// Note: The `bestBefore` field is not compared.
//
// Parameters:
//   - `aEntry`: Cache entry to compare with.
//
// Returns:
//   - `bool`: `true` if the cache entry is equal to the given one, `false` otherwise.
func (ce *tMapEntry) Equal(aEntry *tMapEntry) bool {
	if nil == ce {
		return (nil == aEntry)
	}
	if nil == aEntry {
		return false
	}
	if ce == aEntry {
		return true
	}
	if 0 == len(ce.ips) {
		return (0 == len(aEntry.ips))
	}
	if 0 == len(aEntry.ips) {
		return false
	}
	if len(ce.ips) != len(aEntry.ips) {
		return false
	}

	// Do NOT compare the `bestBefore` field because even nanoseconds
	// would make a difference.

	return ce.ips.Equal(aEntry.ips)
} // Equal()

// `First()` returns the first IP address in the cache entry.
//
// Returns:
//   - `net.IP`: First IP address in the cache entry.
func (ce *tMapEntry) First() net.IP {
	if nil == ce {
		return nil
	}

	return ce.ips.First()
} // First()

// `isExpired()` returns `true` if the cache entry is expired.
//
// Returns:
//   - `bool`: `true` if the cache entry is expired, `false` otherwise.
func (ce *tMapEntry) isExpired() bool {
	if (nil == ce) || (0 == len(ce.ips)) {
		return true
	}

	return ce.bestBefore.Before(time.Now())
} // isExpired()

// `Len()` returns the number of IP addresses in the cache entry.
//
// Returns:
//   - `int`: Number of IP addresses in the cache entry.
func (ce *tMapEntry) Len() int {
	if nil == ce {
		return 0
	}

	return ce.ips.Len()
} // Len()

// `Retrieve()` returns the IP addresses cached by this cache entry.
//
// NOTE: This method's implementation ignores both, the `Context` and
// `tPartsList` arguments required by the `iCacheNode` interface.`
//
// Returns:
//   - `tIpList`: The list of IP addresses for the given pattern.
func (ce *tMapEntry) Retrieve(context.Context, tPartsList) tIpList {
	if nil == ce {
		return tIpList{}
	}

	return ce.ips
} // Retrieve()

// `String()` implements the `fmt.Stringer` interface for the cache entry.
//
// Returns:
//   - `string`: String representation of the cache entry.
func (ce *tMapEntry) String() string {
	if nil == ce {
		return ""
	}
	var builder strings.Builder

	if 0 < len(ce.ips) {
		fmt.Fprint(&builder, ce.ips.String())
		fmt.Fprint(&builder, "\n")
	}
	fmt.Fprint(&builder, ce.bestBefore.Format(defTimeFormat))

	return builder.String()
} // String()

// `Update()` updates the cache entry with the given IP addresses returning
// the updated cache entry.
//
// If the given IP list is empty, the cache entry's IP list is cleared/removed.
//
// NOTE: This method's implementation ignores  the `Context` argument
// required by the `iCacheNode` interface.`
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aIPs`: List of IP addresses to Update the cache entry with.
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `iCacheNode`: The updated cache entry.
func (ce *tMapEntry) Update(aCtx context.Context, aIPs tIpList, aTTL time.Duration) iCacheNode {
	if nil == ce {
		return nil
	}
	if ce.ips.Equal(aIPs) {
		return ce
	}
	if 0 == aTTL {
		aTTL = DefaultTTL
	}

	if iLen := len(aIPs); 0 < iLen {
		// Assume ownership of `aIPs`
		ce.ips = make(tIpList, iLen)
		copy(ce.ips, aIPs)

		// Update expiration time
		ce.bestBefore = time.Now().Add(aTTL)
	} else {
		ce.ips = tIpList{}
		ce.bestBefore = time.Time{}
	}

	return ce
} // Update()

/* _EoF_ */
