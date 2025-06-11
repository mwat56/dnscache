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
	// `poolDropMask` is the bit mask to use for limiting the pool returns.
	poolDropMask = 7 // 0111

	// `poolInitSize` is the number of items to pre-allocate for the
	// pool during initialisation.
	// Four times this value is used as the pool's maximum size.
	poolInitSize = 1 << 9 // 512
)

type (
	// `tPoolMetrics` contains the metrics data for the pool.
	//
	// These are the fields providing the metrics data:
	//
	//   - `Created`: Number of items created by the pool.
	//   - `Returned`: Number of items returned to the pool.
	//   - `Size`: Current number of items in the pool.
	tPoolMetrics struct {
		Created  uint32
		Returned uint32
		Size     int
	}

	// `tTriePool` is a bounded pool of Trie nodes.
	//
	// The pool is inherently thread-safe. Its size is fixed and can't
	// be changed after creation.
	//
	// The pool's factory function `New()` is called to create new items.
	//
	tTriePool struct {
		New      func() any    // Factory function
		nodes    chan any      // Bounded channel
		created  atomic.Uint32 // Number of items created
		returned atomic.Uint32 // Number of items returned
	}
)

var (
	// `triePool` is the running pool of `tTrieNode` instances.
	triePool *tTriePool

	// Make sure, the trie node pool is only initialised once.
	triePoolInit sync.Once
)

// ---------------------------------------------------------------------------
// Initialise the trie node pool:

// `init()` pre-allocates some nodes for the pool.
func init() {
	initReal()
} // init()

// `initReal()` pre-allocates some nodes for the pool.
//
// This function is called only once during package initialisation.
//
// During unit testing, this function could be called manually.
func initReal() {
	triePoolInit.Do(func() {
		triePool = &tTriePool{
			nodes: make(chan any, int(poolInitSize<<2)),
			New: func() any {
				node := &tTrieNode{tChildren: make(tChildren)}
				triePool.created.Add(1)
				/*TODO: Go 1.24:
				runtime.AddCleanup(node, func() {
					pool.Put(node)
				})
				*/
				return node
			},
		}

		for range poolInitSize {
			// Pre-allocate some nodes for the pool:
			triePool.Put(&tTrieNode{tChildren: make(tChildren)})
		}
	}) // triePoolInit.Do()
} // initReal()

// ---------------------------------------------------------------------------
// `tTriePool` methods:

// `Get()` returns an item from the pool.
//
// If the pool is empty, a new item is created.
//
// Returns:
//   - `any`: An item from the pool.
func (np *tTriePool) Get() any {
	if nil == np {
		initReal() // initialise the node pool
		np = triePool
	}

	select {
	case item := <-np.nodes:
		return item
	default:
		return np.New()
	}
} // Get()

// `Put()` returns an item to the pool.
//
// If the pool is full, the item is dropped.
//
// Parameters:
//   - `aNode`: The node to return to the pool.
func (np *tTriePool) Put(aNode *tTrieNode) {
	if nil == np {
		initReal() // initialise the node pool
		np = triePool
	}
	if (np.returned.Add(1) & poolDropMask) == poolDropMask {
		// Drop the node if the drop mask matches.
		// This leaves the given `aNode` for GC.
		// With a drop mask of `7` (0111) we drop 1 in 8 nodes.
		return
	}

	select {
	case np.nodes <- aNode:
		// Item was added to pool
		runtime.Gosched()

	default:
		// Drop if pool is full
		runtime.Gosched()
	}
} // Put()

// ---------------------------------------------------------------------------
// `tTrieNode` constructor:

// `newTrieNode()` creates a new `tTrieNode` instance.
//
// Returns:
//   - `*tTrieNode`: A new `tTrieNode` instance.
func newTrieNode() *tTrieNode {
	node, ok := triePool.Get().(*tTrieNode)
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

// `poolMetrics()` returns the current pool metrics.
//
// Returns:
//   - `rCreated`: Number of nodes created by the pool.
//   - `rReturned`: Number of nodes returned to the pool.
//   - `rSize`: Current number of items in the pool.
func poolMetrics() *tPoolMetrics {
	if nil == triePool {
		initReal() // initialise the node pool
	}
	np := triePool

	return &tPoolMetrics{
		Created:  np.created.Load(),
		Returned: np.returned.Load(),
		Size:     len(np.nodes),
	}
} // poolMetrics()

// `putNode()` returns a node to the pool.
//
// Parameters:
//   - `aNode`: The node to return to the pool.
func putNode(aNode *tTrieNode) {
	// We can't clear the node's fields yet since it might
	// still be used by another list or goroutine.
	triePool.Put(aNode)
} // putNode()

/* _EoF_ */
