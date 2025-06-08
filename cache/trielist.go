/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"context"
	"runtime"
	"sync"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// `DefaultCacheSize` is the initial size of the cache list.
	XDefaultCacheSize = 1 << 9 // 512
)

type (

	//
	// `tRoot` is the root node of the trie.
	//
	// The root node is a special case as it doesn't have a label but
	// can have multiple children (i.e. the TLDs). Also it provides the
	// Mutex to use for locking access to the trie.
	tRoot struct {
		sync.RWMutex // barrier for concurrent access
		node         *tCacheNode
	}

	//
	// `tTrie` is a thread-safe trie for FQDN wildcards. It
	// basically provides a CRUD interface for FQDN patterns.
	//
	//   - `C`: Create a new pattern [Add],
	//   - `R`: Retrieve a pattern [Match],
	//   - `U`: Update a pattern [Update],
	//   - `D`: Delete a pattern [Delete].
	TTrieList struct {
		_     struct{} // placeholder for embedding
		tRoot          // embedded root node of the trie
	}
)

// ---------------------------------------------------------------------------
// `TTrieList` constructor:

// `newTrie()` creates a new `tTrie` instance.
//
// Returns:
//   - `*tTrie`: A new `tTrie` instance.
func newTrie() *TTrieList {
	return &TTrieList{
		tRoot: tRoot{
			node: newNode(),
		}, // root node of the trie
	}
} // newTrie()

// ---------------------------------------------------------------------------
// `TTrieList` methods:

// `AutoExpire()` removes expired cache entries at a given interval.
//
// Parameters:
//   - `aRate`: Time interval to refresh the cache.
//   - `aAbort`: Channel to receive a signal to abort.
func (tl *TTrieList) AutoExpire(aRate time.Duration, aAbort chan struct{}) {
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

// `Delete()` removes the cache entry for the given hostname.
//
// Parameters:
//   - `aHostname`: The hostname to remove the cache entry for.
//
// Returns:
//   - `*TTrieList`: The updated trie.
func (tl *TTrieList) Delete(aHostname string) *TTrieList {
	if nil == tl {
		return nil
	}

	tl.Lock()
	tl.node.Delete(context.TODO(), pattern2parts(aHostname))
	tl.Unlock()

	return tl
} // Delete()

// `Equal()` checks whether the cache list is equal to the given one.
//
// Parameters:
//   - `aList`: Cache list to compare with.
//
// Returns:
//   - `bool`: `true` if the cache list is equal to the given one, `false` otherwise.
func (tl *TTrieList) Equal(aList *TTrieList) (rOK bool) {
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

// `expireEntries()` removes all expired cache entries.
//
// This method is called automatically by the `AutoExpire()` method.
func (tl *TTrieList) expireEntries() {
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
//   - `aHostname`: The hostname to resolve.
//
// Returns:
//   - `TIpList`: List of IP addresses for the given hostname.
//   - `bool`: `true` if the hostname was found in the cache, `false` otherwise.
func (tl *TTrieList) IPs(aHostname string) (tIpList, bool) {
	if nil == tl {
		return nil, false
	}

	tl.RLock()
	ips := tl.node.Read(context.TODO(), pattern2parts(aHostname))
	tl.RUnlock()

	return ips, (0 < len(ips))
} // IPs()

// `Len()` returns the number of hostname entries in the cache list.
//
// Returns:
//   - `int`: Number of entries in the cache list.
func (tl *TTrieList) Len() int {
	if nil == tl {
		return 0
	}

	tl.RLock()
	_, patterns := tl.node.count(context.TODO())
	tl.RUnlock()

	return patterns
} // Len()

// `SetEntry()` adds a new cache entry for the given hostname.
//
// Parameters:
//   - `aHostname`: The hostname to add a cache entry for.
//   - `aIPs`: List of IP addresses to add to the cache entry.
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `*TTrieList`: The updated cache list.
func (tl *TTrieList) SetEntry(aHostname string, aIPs tIpList, aTTL time.Duration) *TTrieList {
	if (nil == tl) || (0 == len(aIPs)) {
		return tl
	}

	parts := pattern2parts(aHostname)
	tl.Lock()
	tl.node.Add(context.TODO(), parts, aIPs, aTTL)
	tl.Unlock()

	return tl
} // SetEntry()

// `String()` implements the `fmt.Stringer` interface for a string
// representation of the cache list.
//
// Returns:
//   - `string`: String representation of the cache list.
func (tl *TTrieList) String() (rStr string) {
	if nil == tl {
		return ""
	}

	tl.RLock()
	rStr = tl.node.String()
	tl.RUnlock()

	return
} // String()

/* _EoF_ */
