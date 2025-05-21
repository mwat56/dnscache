/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"runtime"
	"sync"
	"sync/atomic"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// `poolDropMask` is the bit mask to use for limiting the pool returns.
	poolDropMask = 7 // 111

	// `poolInitSize` is the number of items to pre-allocate for the
	// pool during initialisation.
	// Four times this value is used as the pool's maximum size.
	poolInitSize = 256
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

	// `tPool` is a bounded pool of items.
	//
	// The pool is inherently thread-safe. Its size is fixed and can't
	// be changed after creation.
	//
	// The pool's factory function `New()` is called to create new items.
	//
	tPool struct {
		New      func() *tNode // Factory function
		nodes    chan any      // Bounded channel
		created  atomic.Uint32 // Number of items created
		returned atomic.Uint32 // Number of items returned
	}
)

var (
	// `nodePool` is the running pool of `tNode` instances.
	nodePool *tPool

	// Make sure, the node pool is only initialised once.
	poolInit sync.Once
)

// ---------------------------------------------------------------------------
// Initialise the node pool:

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
	poolInit.Do(func() {
		nodePool = newPool(int(poolInitSize<<2), nil)
		nodePool.New = func() *tNode {
			nodePool.created.Add(1)
			return &tNode{
				tChildren: make(tChildren),
			}
		}

		for range poolInitSize {
			nodePool.Put(&tNode{tChildren: make(tChildren)})
		}
	}) // poolInit.Do()
} // initReal()

// ---------------------------------------------------------------------------
// `tPool` constructor:

// `newPool()` returns a new bounded pool.
//
// This function is called only once during package initialisation.
//
// The pool's size is fixed and can't be changed after creation.
// If `aSize` is less than or equal to zero, the default size
// (`poolInitSize`) is used.
//
// The created pool is inherently thread-safe by using a channel for
// storing nodes.
//
// Parameters:
//   - `aSize`: The maximum number of items in the pool.
//   - `newFunc`: The function to create a new item.
//
// Returns:
//   - `*tPool`: A new `tPool` instance.
func newPool(aSize int, newFunc func() *tNode) *tPool {
	if 0 >= aSize {
		aSize = poolInitSize
	}
	return &tPool{
		nodes: make(chan any, aSize),
		New:   newFunc,
	}
} // newPool()

// ---------------------------------------------------------------------------
// `tPool` methods:

// `Get()` returns an item from the pool.
//
// If the pool is empty, a new item is created.
//
// Returns:
//   - `any`: An item from the pool.
func (p *tPool) Get() any {
	if nil == p {
		initReal() // initialise the node pool
		p = nodePool
	}

	select {
	case item := <-p.nodes:
		return item
	default:
		return p.New()
	}
} // Get()

// `Put()` returns an item to the pool.
//
// If the pool is full, the item is dropped.
//
// Parameters:
//   - `aNode`: The node to return to the pool.
func (p *tPool) Put(aNode *tNode) {
	if nil == p {
		initReal() // initialise the node pool
		p = nodePool
	}
	if (p.returned.Add(1) & poolDropMask) == poolDropMask {
		// Drop the node if the drop mask matches.
		// This leaves the given `aNode` for GC.
		// With a drop mask of `7` (111) we drop 1 in 8 nodes.
		return
	}

	select {
	case p.nodes <- aNode:
		// Item was added to pool
		runtime.Gosched()
	default:
		// Drop if pool is full
		runtime.Gosched()
	}
} // Put()

// ---------------------------------------------------------------------------
// `tNode` constructor:

// `newNode()` creates a new `tNode` instance.
//
// Returns:
//   - `*tNode`: A new `tNode` instance.
func newNode() *tNode {
	result, ok := nodePool.Get().(*tNode)
	if ok {
		// Clear/reset the old field values
		if 0 < len(result.tChildren) {
			result.tChildren = make(tChildren)
		}
		result.isEnd, result.isWild = false, false
	} else {
		result = &tNode{tChildren: make(tChildren)}
	}

	return result
} // newNode()

// ---------------------------------------------------------------------------
// Helper functions:

// `poolMetrics()` returns the current pool metrics.
//
// Returns:
//   - `rCreated`: Number of nodes created by the pool.
//   - `rReturned`: Number of nodes returned to the pool.
//   - `rSize`: Current number of items in the pool.
func poolMetrics() *tPoolMetrics {
	if nil == nodePool {
		initReal() // initialise the node pool
	}
	np := nodePool

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
func putNode(aNode *tNode) {
	// We can't clear the node's fields yet since it might
	// still be used by another list or goroutine.
	nodePool.Put(aNode)
} // putNode()

/* _EoF_ */
