/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package dnscache

import (
	"fmt"
	"maps"
	"net"
	"runtime"
	"slices"
	"strings"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// `defTimeFormat` is the default time format for the cache entries'
	// string representation of the `created` field.
	defTimeFormat = "2006-01-02 15:04:05.999999999"
)

type (
	// `tIpList` is a list of IP addresses.
	tIpList []net.IP

	// `tCacheEntry` is a DNS cache entry.
	tCacheEntry struct {
		ips     tIpList       // IP addresses for this entry
		created time.Time     // time of creation
		ttl     time.Duration // time to live
		// reuse   uint32         // number of reuses
	}

	// `tCacheList` is a map of DNS cache entries.
	tCacheList map[string]*tCacheEntry
)

// ---------------------------------------------------------------------------
// `tIpList` methods:

// `Equal()` checks whether the IP list is equal to the given one.
//
// Parameters:
//   - `aList`: List to compare with.
//
// Returns:
//   - `bool`: `true` if the lists are equal, `false` otherwise.
func (il tIpList) Equal(aList tIpList) bool {
	if nil == il {
		return nil == aList
	}
	if nil == aList {
		return false
	}
	if len(il) != len(aList) {
		return false
	}

	return slices.EqualFunc(il, aList, func(ip1, ip2 net.IP) bool {
		return ip1.Equal(ip2)
	})
} // Equal()

// `String()` implements the `fmt.Stringer` interface for a string
// representation of the IP list.
//
// Returns:
//   - `string`: String representation of the IP list.
func (il *tIpList) String() string {
	if nil == il {
		return ""
	}
	lLen := len(*il)
	if 0 == lLen {
		return ""
	}
	if 1 == lLen {
		return (*il)[0].String()
	}

	var builder strings.Builder
	for i := range lLen {
		if nil != (*il)[i] {
			fmt.Fprint(&builder, (*il)[i].String())
			if i < lLen-1 {
				fmt.Fprintf(&builder, " - ")
			}
		}
	}

	return builder.String()
} // String()

// ---------------------------------------------------------------------------
// `tCacheEntry` constructor:

// `newCacheEntry()` returns a new cache entry with the given TTL.
//
// Parameters:
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `*tCacheEntry`: A new cache entry.
func newCacheEntry(aTTL time.Duration) *tCacheEntry {
	if 0 == aTTL {
		aTTL = defTTL
	}

	return &tCacheEntry{
		created: time.Now(),
		ttl:     aTTL,
	}
} // newCacheEntry()

// ---------------------------------------------------------------------------
// `tCacheEntry` methods:

// `clone()` returns a deep copy of the cache entry.
//
// Returns:
//   - `*tCacheEntry`: A deep copy of the cache entry.
func (ce *tCacheEntry) clone() *tCacheEntry {
	if nil == ce {
		return nil
	}

	result := &tCacheEntry{
		created: ce.created,
		ttl:     ce.ttl,
	}
	if (nil != ce.ips) && (0 < len(ce.ips)) {
		result.ips = slices.Clone(ce.ips)
	}

	return result
} // clone()

// `Equal()` checks whether the cache entry is equal to the given one.
//
// Note: The `created` and `ttl` fields are not compared.
//
// Parameters:
//   - `aEntry`: Cache entry to compare with.
//
// Returns:
//   - `bool`: `true` if the cache entry is equal to the given one, `false` otherwise.
func (ce *tCacheEntry) Equal(aEntry *tCacheEntry) bool {
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
	if ce.ttl != aEntry.ttl {
		return false
	}
	if nil == aEntry.ips {
		return (nil == ce.ips)
	}
	if len(ce.ips) != len(aEntry.ips) {
		return false
	}

	// Do NOT compare the `created` fields because even nanoseconds
	// make a differences.

	return ce.ips.Equal(aEntry.ips)
} // Equal()

// `isExpired()` returns `true` if the cache entry is expired.
//
// Returns:
//   - `bool`: `true` if the cache entry is expired, `false` otherwise.
func (ce *tCacheEntry) isExpired() bool {
	return ce.created.Add(ce.ttl).Before(time.Now())
} // isExpired()

// `String()` implements the `fmt.Stringer` interface for the cache entry.
//
// Returns:
//   - `string`: String representation of the cache entry.
func (ce *tCacheEntry) String() string {
	if nil == ce {
		return ""
	}
	var builder strings.Builder

	fmt.Fprint(&builder, ce.created.Format(defTimeFormat))
	fmt.Fprint(&builder, "\n")
	if 0 < len(ce.ips) {
		fmt.Fprint(&builder, ce.ips.String())
		fmt.Fprint(&builder, "\n")
	}
	fmt.Fprint(&builder, ce.ttl.String())

	return builder.String()
} // String()

// `update()` updates the cache entry with the given IP addresses.
//
// Parameters:
//   - `aIPs`: List of IP addresses to update the cache entry with.
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `*tCacheEntry`: The updated cache entry.
func (ce *tCacheEntry) update(aIPs tIpList, aTTL time.Duration) *tCacheEntry {
	if (nil == ce) || (nil == aIPs) || (0 == len(aIPs)) {
		return ce
	}

	if ce.ips.Equal(aIPs) {
		ce.ttl = aTTL
		// IPs are the same, no need to update
		return ce
	}

	ce.ips = aIPs
	ce.created = time.Now()
	ce.ttl = aTTL

	return ce
} // update()

// ---------------------------------------------------------------------------
// `tCacheList` constructor:

// `newCacheList()` returns a new TTL cache list.
//
// If `aSize` is zero, the default size (`64`) is used.
//
// Parameters:
//   - `aSize`: Initial size of the cache list.
//
// Returns:
//   - `*tCacheList`: A new TTL cache list.
func newCacheList(aSize uint) *tCacheList {
	if 0 == aSize {
		aSize = defCacheSize
	}

	result := make(tCacheList, aSize)

	return &result
} // newCacheList()

// ---------------------------------------------------------------------------
// `tCacheList` methods:

// `autoExpire()` removes expired cache entries at a given interval.
//
// Parameters:
//   - `aRate`: Time interval to refresh the cache.
//   - `aAbort`: Channel to receive a signal to abort.
func (cl *tCacheList) autoExpire(aRate time.Duration, aAbort chan struct{}) {
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
} // autoExpire()

// `clone()` returns a deep copy of the cache list.
//
// Returns:
//   - `*tCacheList`: A deep copy of the cache list.
func (cl *tCacheList) clone() *tCacheList {
	if nil == cl {
		return nil
	}

	result := make(tCacheList, len(*cl))
	for host, ce := range *cl {
		result[host] = ce.clone()
	}

	return &result
} // clone()

// `delete()` removes the cache entry for the given hostname.
//
// Parameters:
//   - `aHostname`: The hostname to remove the cache entry for.
//
// Returns:
//   - `*tCacheList`: The updated cache list.
func (cl *tCacheList) delete(aHostname string) *tCacheList {
	if nil == cl {
		return nil
	}

	delete(*cl, aHostname)

	return cl
} // delete()

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
	if len(*cl) != len(*aList) {
		return false
	}
	var (
		otherE *tCacheEntry
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
// This method is called automatically by the `autoRemove()` method.
func (cl *tCacheList) expireEntries() {
	clone := maps.Clone(*cl)
	for hostname, ce := range clone {
		if ce.isExpired() {
			delete(*cl, hostname)
		}
	}
	clone = nil
} // expireEntries()

// `getEntry()` returns the cache entry for the given hostname.
//
// Parameters:
//   - `aHostname`: The hostname to get the cache entry for.
//
// Returns:
//   - `*tCacheEntry`: The cache entry for the given hostname.
//   - `bool`: `true` if the hostname was found in the cache, `false` otherwise.
func (cl *tCacheList) getEntry(aHostname string) (*tCacheEntry, bool) {
	if (nil == cl) || (0 == len(*cl)) {
		return nil, false
	}

	if ce, ok := (*cl)[aHostname]; ok {
		return ce, true
	}

	return nil, false
} // getEntry()

// `ips()` returns the IP addresses for the given hostname.
//
// Parameters:
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `tIpList`: List of IP addresses for the given hostname.
//   - `bool`: `true` if the hostname was found in the cache, `false` otherwise.
func (cl *tCacheList) ips(aHostname string) (tIpList, bool) {
	if (nil == cl) || (0 == len(*cl)) {
		return nil, false
	}

	if ce, ok := (*cl)[aHostname]; ok {
		return ce.ips, true
	}

	return nil, false
} // ips()

// `len()` returns the number of entries in the cache list.
//
// Returns:
//   - `int`: Number of entries in the cache list.
func (cl *tCacheList) len() int {
	if nil == cl {
		return 0
	}

	return len(*cl)
} // len()

// `setEntry()` adds a new cache entry for the given hostname.
//
// Parameters:
//   - `aHostname`: The hostname to add a cache entry for.
//   - `aIPs`: List of IP addresses to add to the cache entry.
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `*tCacheList`: The updated cache list.
func (cl *tCacheList) setEntry(aHostname string, aIPs tIpList, aTTL time.Duration) *tCacheList {
	if (nil == cl) || (nil == aIPs) || (0 == len(aIPs)) {
		return cl
	}

	if ce, ok := (*cl)[aHostname]; ok {
		(*cl)[aHostname] = ce.update(aIPs, aTTL)
	} else {
		(*cl)[aHostname] = newCacheEntry(aTTL).update(aIPs, aTTL)
	}

	return cl
} // setEntry()

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
	for host, ce := range *cl {
		fmt.Fprintf(&builder, "%s: %s\n", host, ce.String())
	}

	return builder.String()
} // String()

/* _EoF_ */
