/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"fmt"
	"net"
	"slices"
	"strings"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// `defTimeFormat` is the default time format for the cache entries'
	// string representation of the `bestBefore` field.
	defTimeFormat = "2006-01-02 15:04:05.999999999"

	// `DefaultTTL` is the default time to live for a DNS cache entry.
	DefaultTTL = time.Duration(time.Minute << 6) // 64 minutes
)

type (
	// `TCacheEntry` is a DNS cache entry.
	TCacheEntry struct {
		ips        TIpList   // IP addresses for this entry
		bestBefore time.Time // time after which the entry is not valid
	}
)

// ---------------------------------------------------------------------------
// `TCacheEntry` constructor:

// `newCacheEntry()` returns a new cache entry with the given TTL.
//
// Parameters:
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `*TCacheEntry`: A new cache entry.
func newCacheEntry(aTTL time.Duration) *TCacheEntry {
	if 0 == aTTL {
		aTTL = DefaultTTL
	}

	return &TCacheEntry{
		bestBefore: time.Now().Add(aTTL),
	}
} // newCacheEntry()

// ---------------------------------------------------------------------------
// `TCacheEntry` methods:

// `clone()` returns a deep copy of the cache entry.
//
// Returns:
//   - `*TCacheEntry`: A deep copy of the cache entry.
func (ce *TCacheEntry) clone() *TCacheEntry {
	if nil == ce {
		return nil
	}

	result := &TCacheEntry{
		bestBefore: ce.bestBefore,
	}
	if (nil != ce.ips) && (0 < len(ce.ips)) {
		result.ips = slices.Clone(ce.ips)
	}

	return result
} // clone()

// `Equal()` checks whether the cache entry is equal to the given one.
//
// Note: The `bestBefore` field is not compared.
//
// Parameters:
//   - `aEntry`: Cache entry to compare with.
//
// Returns:
//   - `bool`: `true` if the cache entry is equal to the given one, `false` otherwise.
func (ce *TCacheEntry) Equal(aEntry *TCacheEntry) bool {
	if nil == ce {
		return (nil == aEntry)
	}
	if nil == aEntry {
		return false
	}
	if ce == aEntry {
		return true
	}
	if nil == ce.ips {
		return (nil == aEntry.ips)
	}
	if nil == aEntry.ips {
		return false
	}
	if len(ce.ips) != len(aEntry.ips) {
		return false
	}

	// Do NOT compare the `bestBefore` field because even nanoseconds
	// make a differences.

	return ce.ips.Equal(aEntry.ips)
} // Equal()

// `First()` returns the first IP address in the cache entry.
//
// Returns:
//   - `net.IP`: First IP address in the cache entry.
func (ce *TCacheEntry) First() net.IP {
	if nil == ce {
		return nil
	}

	return ce.ips.First()
} // First()

// `isExpired()` returns `true` if the cache entry is expired.
//
// Returns:
//   - `bool`: `true` if the cache entry is expired, `false` otherwise.
func (ce *TCacheEntry) isExpired() bool {
	if nil == ce {
		return true
	}

	return ce.bestBefore.Before(time.Now())
} // isExpired()

// `Len()` returns the number of IP addresses in the cache entry.
//
// Returns:
//   - `int`: Number of IP addresses in the cache entry.
func (ce *TCacheEntry) Len() int {
	if nil == ce {
		return 0
	}

	return ce.ips.Len()
} // Len()

// `String()` implements the `fmt.Stringer` interface for the cache entry.
//
// Returns:
//   - `string`: String representation of the cache entry.
func (ce *TCacheEntry) String() string {
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

// `update()` updates the cache entry with the given IP addresses.
//
// Parameters:
//   - `aIPs`: List of IP addresses to update the cache entry with.
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `*TCacheEntry`: The updated cache entry.
func (ce *TCacheEntry) update(aIPs TIpList, aTTL time.Duration) *TCacheEntry {
	if (nil == ce) || (nil == aIPs) || (0 == len(aIPs)) {
		return ce
	}

	if !ce.ips.Equal(aIPs) {
		ce.ips = aIPs
	}

	// Update expiration time:
	ce.bestBefore = time.Now().Add(aTTL)

	return ce
} // update()

/* _EoF_ */
