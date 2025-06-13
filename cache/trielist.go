/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"context"
	"net"
	"runtime"
	"sort"
	"sync"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	//
	// `tRoot` is the root node of the Trie.
	//
	// The root node is a special case as it doesn't have a label but
	// can have multiple children (i.e. the TLDs). Also it provides
	// the Mutex to use for locking access to the Trie.
	tRoot struct {
		sync.RWMutex // barrier for concurrent access
		node         *tTrieNode
	}

	//
	// `tTrieList` is a thread-safe Trie for FQDN wildcards. It
	// basically provides a CRUD interface for FQDN patterns.
	//
	//   - `C`: Create a new pattern [Add],
	//   - `R`: Retrieve a pattern [Match],
	//   - `U`: Update a pattern [Update],
	//   - `D`: Delete a pattern [Delete].
	tTrieList struct {
		_     struct{} // placeholder for embedding
		tRoot          // embedded root node of the Trie
	}
)

// ---------------------------------------------------------------------------
// `tTrieList` constructor:

// `newTrie()` creates a new `tTrieList` instance.
//
// Returns:
//   - `*tTrieList`: A new `tTrieList` instance.
func newTrie() *tTrieList {
	return &tTrieList{
		tRoot: tRoot{
			node: newTrieNode(),
		}, // root node of the Trie
	}
} // newTrie()

// ---------------------------------------------------------------------------

// `init()` ensures proper interface implementation.
func init() {
	var (
		_ ICacheList = (*tTrieList)(nil)
	)
} // init()

// ---------------------------------------------------------------------------
// `tTrieList` methods:

// `AutoExpire()` removes expired cache entries at a given interval.
//
// Parameters:
//   - `aRate`: Time interval to refresh the cache.
//   - `aAbort`: Channel to receive a signal to abort.
func (tl *tTrieList) AutoExpire(aRate time.Duration, aAbort chan struct{}) {
	ticker := time.NewTicker(aRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go tl.expireEntries()
			runtime.Gosched() // yield to other goroutines

		case <-aAbort:
			return
		}
	}
} // AutoExpire()

// `Clone()` creates a deep copy of the Trie.
//
// Returns:
//   - `*tTrieList`: A deep copy of the Trie.
func (tl *tTrieList) Clone() ICacheList {
	if nil == tl {
		return nil
	}

	tl.RLock()
	root := tl.tRoot.node.clone()
	tl.RUnlock()
	if nil == root {
		return nil
	}

	return &tTrieList{
		tRoot: tRoot{
			node: root,
		},
	}
} // Clone()

// `Create()` adds a new cache entry for the given hostname.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aHostname`: The hostname to add a cache entry for.
//   - `aIPs`: List of IP addresses to add to the cache entry.
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `*tTrieList`: The updated cache list.
func (tl *tTrieList) Create(aCtx context.Context, aHostname string, aIPs []net.IP, aTTL time.Duration) ICacheList {
	if nil == tl {
		return nil
	}

	parts := pattern2parts(aHostname)
	tl.Lock()
	tl.node.Create(aCtx, parts, aIPs, aTTL)
	tl.Unlock()

	return tl
} // Create()

// `Delete()` removes the cache entry for the given hostname.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aHostname`: The hostname to remove the cache entry for.
//
// Returns:
//   - `*tTrieList`: The updated Trie.
func (tl *tTrieList) Delete(aCtx context.Context, aHostname string) (rOK bool) {
	if nil == tl {
		return
	}

	tl.Lock()
	rOK = tl.node.Delete(aCtx, pattern2parts(aHostname))
	tl.Unlock()

	return
} // Delete()

// `Equal()` checks whether the cache list is equal to the given one.
//
// Parameters:
//   - `aList`: Cache list to compare with.
//
// Returns:
//   - `bool`: `true` if the cache list is equal to the given one, `false` otherwise.
func (tl *tTrieList) Equal(aList *tTrieList) (rOK bool) {
	if nil == tl {
		return (nil == aList)
	}
	if nil == aList {
		return
	}

	tl.RLock()
	aList.RLock()
	rOK = tl.node.Equal(aList.node)
	aList.RUnlock()
	tl.RUnlock()

	return
} // Equal()

// `Exists()` checks whether the given hostname is cached.
//
// Parameters:
//   - `context.Context`: Timeout context to use for the operation.
//   - `string`: The hostname to check for.
//
// Returns:
//   - `bool`: `true` if the hostname was found in the cache, `false` otherwise.
func (tl *tTrieList) Exists(aCtx context.Context, aHostname string) (rOK bool) {
	if nil == tl {
		return
	}

	parts := pattern2parts(aHostname)
	tl.RLock()
	_, rOK = tl.node.finalNode(aCtx, parts)
	tl.RUnlock()

	return
} // Exists()

// `expireEntries()` removes all expired cache entries.
//
// This method is called automatically by the `AutoExpire()` method.
func (tl *tTrieList) expireEntries() {
	if nil == tl {
		return
	}

	tl.Lock()
	tl.node.expire(context.TODO())
	tl.Unlock()
} // expireEntries()

// `IPs()` returns the IP addresses for the given hostname.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aHostname`: The hostname to lookup in the cache.
//
// Returns:
//   - `rIPs`: List of IP addresses for the given hostname.
//   - `rOK`: `true` if the hostname was found in the cache, `false` otherwise.
func (tl *tTrieList) IPs(aCtx context.Context, aHostname string) (rIPs []net.IP, rOK bool) {
	if nil == tl {
		return
	}

	tl.RLock()
	ips := tl.node.Retrieve(aCtx, pattern2parts(aHostname))
	rOK = (0 < len(ips))
	tl.RUnlock()

	if rOK {
		rIPs = make([]net.IP, len(ips))
		copy(rIPs, ips)
	}

	return
} // IPs()

// `Len()` returns the number of hostname entries in the cache list.
//
// Returns:
//   - `int`: Number of entries in the cache list.
func (tl *tTrieList) Len() int {
	if nil == tl {
		return 0
	}

	tl.RLock()
	_, patterns := tl.node.count(context.TODO())
	tl.RUnlock()

	return patterns
} // Len()

/* * /
// `RangeX()` returns a channel that yields all FQDNs in sorted order.
//
// Usage: for fqdn := range fqdnList.Range() { ... }
//
// The channel is closed automatically when all entries have been yielded.
func (tl *tTrieList) RangeX(aCtx context.Context) <-chan string {
	ch := make(chan string)
	if nil == tl {
		close(ch)
		return ch
	}

	go func() {
		defer close(ch)
		tl.RLock()
		for _, fqdn := range tl.node.allPatterns(aCtx) {
			ch <- fqdn
		}
		tl.RUnlock()
	}()

	return ch
} // RangeX()
/* */

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
func (tl *tTrieList) Range(aCtx context.Context) <-chan string {
	ch := make(chan string)
	if nil == tl {
		close(ch)
		return ch
	}

	go func() {
		defer close(ch)
		tl.RLock()
		defer tl.RUnlock()

		type tStackEntry struct {
			node *tTrieNode
			path tPartsList
		}
		stack := []tStackEntry{
			{node: tl.tRoot.node, path: []string{}},
		}
		var ( // avoid repeated allocations during loop
			cLen, idx          int
			entry              tStackEntry
			kidNames, newParts tPartsList
			label              string
		)

		for 0 < len(stack) {
			// Check for timeout or cancellation
			if nil != aCtx.Err() {
				// Leaving the goroutine will close the
				// channel (due to `defer close(ch)`).
				return
			}

			entry = stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			// Emit FQDN if terminal (i.e. has IP addresses)
			if 0 < len(entry.node.tCachedIP.tIpList) {
				// Send FQDN through channel
				select {
				case ch <- entry.path.String():
					// Successfully sent FQDN
					runtime.Gosched()
				case <-aCtx.Done():
					// Context is already canceled, discard FQDN.
					// Leaving the goroutine will close the
					// channel (due to `defer close(ch)`).
					return
				}
			}

			if cLen = len(entry.node.tChildren); 0 == cLen {
				continue
			}

			// Process children in sorted order
			kidNames = make(tPartsList, 0, cLen)
			for label = range entry.node.tChildren {
				kidNames = append(kidNames, label)
			}
			if 1 < len(kidNames) {
				sort.Strings(kidNames)
			}

			// Push children to stack in reverse-sorted order
			// (to process them in forward order when popped)
			for idx = len(kidNames) - 1; 0 <= idx; idx-- {
				label = kidNames[idx]

				newParts = make(tPartsList, len(entry.path)+1)
				copy(newParts, entry.path)
				newParts[len(entry.path)] = label

				stack = append(stack, tStackEntry{
					node: entry.node.tChildren[label],
					path: newParts,
				})
			}
		}
	}()

	return ch
} // Range()

// `String()` implements the `fmt.Stringer` interface for a string
// representation of the cache list.
//
// Returns:
//   - `string`: String representation of the cache list.
func (tl *tTrieList) String() (rStr string) {
	if nil == tl {
		return ""
	}

	tl.RLock()
	rStr = tl.node.String()
	tl.RUnlock()

	return
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
//   - `*tTrieList`: The updated cache list.
func (tl *tTrieList) Update(aCtx context.Context, aHostname string, aIPs []net.IP, aTTL time.Duration) ICacheList {
	if nil == tl {
		return nil
	}

	parts := pattern2parts(aHostname)
	tl.Lock()
	if cn, ok := tl.node.finalNode(aCtx, parts); ok {
		// There's actually a matching cache entry
		cn.Update(aCtx, aIPs, aTTL)
	} else {
		// No matching cache entry, thus create a new one
		tl.node.Create(aCtx, parts, aIPs, aTTL)
	}
	tl.Unlock()

	return tl
} // Update()

/* _EoF_ */
