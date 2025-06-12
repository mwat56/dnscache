/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"runtime"
	"sync"
	"sync/atomic"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// `triePoolDropMask` is the bit mask to use for limiting the pool returns.
	triePoolDropMask = 7 // 0111

	// `triePoolInitSize` is the number of items to pre-allocate for the
	// pool during initialisation.
	// Four times this value is used as the pool's maximum size.
	triePoolInitSize = 1 << 9 // 512
)

type (
	// `tTriePoolMetrics` contains the metrics data for the pool.
	//
	// These are the fields providing the metrics data:
	//
	//   - `Created`: Number of items created by the pool.
	//   - `Returned`: Number of items returned to the pool.
	//   - `Size`: Current number of items in the pool.
	tTriePoolMetrics struct {
		Created  uint32
		Returned uint32
		Size     int
	}

	// `tTriePool` is a bounded pool of Trie nodes.
	//
	// The pool is inherently thread-safe. Its size is fixed and can't
	// be changed after creation.
	//
	// The pool's `get()` and `new()` methods return an unspecified item
	// (`any`) from the pool and the internal channel accepts `any` item.
	// That way the constructor function (calling the pool's `get()`
	// method) can differentiate between an item returned by the pool
	// and a newly created one.
	//
	tTriePool struct {
		new      func() any    // Factory function
		nodes    chan any      // Bounded channel
		created  atomic.Uint32 // Number of items created
		returned atomic.Uint32 // Number of items returned
	}
)

var (
	// `triePool` is the active pool of `tTrieNode` instances.
	triePool *tTriePool

	// Make sure, the trie node pool is only initialised once.
	triePoolInit sync.Once
)

// ---------------------------------------------------------------------------
// Initialise the trie node pool:

// `init()` pre-allocates some nodes for the pool.
func init() {
	initTriePool()
} // init()

// `initTriePool()` pre-allocates some nodes for the pool.
//
// This function is called only once during package initialisation.
//
// During unit testing, this function could be called manually.
func initTriePool() {
	triePoolInit.Do(func() {
		triePool = &tTriePool{
			nodes: make(chan any, int(triePoolInitSize<<2)),
			new: func() any {
				node := &tTrieNode{tChildren: make(tChildren)}
				triePool.created.Add(1)
				//TODO: Go 1.24:
				// runtime.AddCleanup(node, func() {
				// 	pool.put(node)
				// })
				return node
			},
		}

		for range triePoolInitSize {
			// Pre-allocate some nodes for the pool:
			triePool.put(&tTrieNode{tChildren: make(tChildren)})
		}
	}) // triePoolInit.Do()
} // initTriePool()

// ---------------------------------------------------------------------------
// `tTriePool` methods:

// `get()` returns an item from the pool.
//
// If the pool is empty, a new item is created.
//
// Returns:
//   - `any`: An item from the pool.
func (tp *tTriePool) get() any {
	if nil == tp {
		initTriePool() // initialise the node pool
		tp = triePool
	}

	select {
	case node := <-tp.nodes:
		return node
	default:
		return tp.new()
	}
} // get()

// `put()` returns an item to the pool.
//
// If the pool is full, the item is dropped.
//
// Parameters:
//   - `aNode`: The node to return to the pool.
func (tp *tTriePool) put(aNode *tTrieNode) {
	if nil == tp {
		initTriePool() // initialise the node pool
		tp = triePool
	}
	if (tp.returned.Add(1) & triePoolDropMask) == triePoolDropMask {
		// Drop the node if the drop mask matches.
		// This leaves the given `aNode` for GC.
		// With a drop mask of `7` (0111) we drop 1 in 8 nodes.
		runtime.GC()
		return
	}

	select {
	case tp.nodes <- aNode:
		// Item was added to pool
		runtime.Gosched()

	default:
		// Drop if pool is full
		runtime.Gosched()
	}
} // put()

// ---------------------------------------------------------------------------
// `tTrieNode` constructor:

// `newTrieNode()` returns a new `tTrieNode` instance.
//
// Returns:
//   - `*tTrieNode`: A new `tTrieNode` instance.
func newTrieNode() *tTrieNode {
	node, ok := triePool.get().(*tTrieNode)
	if ok {
		// Clear/reset the old field values
		node.tCachedIP = tCachedIP{}
		if 0 < len(node.tChildren) {
			node.tChildren = make(tChildren)
		}
	} else {
		node = &tTrieNode{tChildren: make(tChildren)}
	}

	return node
} // newTrieNode()

// ---------------------------------------------------------------------------
// Helper functions:

// `triePoolMetrics()` returns the current pool metrics.
//
// Returns:
//   - `*tTriePoolMetrics`: Current pool metrics.
func triePoolMetrics() *tTriePoolMetrics {
	if nil == triePool {
		initTriePool() // initialise the node pool
	}
	tp := triePool

	return &tTriePoolMetrics{
		Created:  tp.created.Load(),
		Returned: tp.returned.Load(),
		Size:     len(tp.nodes),
	}
} // triePoolMetrics()

// `putNode()` returns a node to the pool.
//
// Parameters:
//   - `aNode`: The node to return to the pool.
func putNode(aNode *tTrieNode) {
	// We can't clear the node's fields yet since it might
	// still be used by another list or goroutine.
	triePool.put(aNode)
} // putNode()

/* _EoF_ */
