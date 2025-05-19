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
	// `tPool` is a bounded pool of items.
	//
	// The pool is inherently thread-safe. Its size is fixed and can't
	// be changed after creation.
	//
	// The pool's factory function `New()` is called to create new items.
	//
	tPool struct {
		New      func() any    // Factory function
		items    chan any      // Bounded channel
		created  atomic.Uint32 // Number of items created
		returned atomic.Uint32 // Number of items returned
	}
)

var (
	nodePool *tPool
	once     sync.Once
)

// ---------------------------------------------------------------------------
// Initialise the node pool:

// `init()` pre-allocates some nodes for the pool.
func init() {
	once.Do(func() {
		nodePool = newPool(int(poolInitSize<<2), nil)
		nodePool.New = func() any {
			nodePool.created.Add(1)
			return &tNode{tChildren: make(tChildren)}
		}
	}) // once.Do()

	for range poolInitSize {
		nodePool.Put(&tNode{tChildren: make(tChildren)})
	}
} // init()

// ---------------------------------------------------------------------------
// `tPool` constructor:

// `newPool()` returns a new bounded pool.
//
// Parameters:
//   - `aSize`: The maximum number of items in the pool.
//   - `newFunc`: The function to create a new item.
//
// Returns:
//   - `*tPool`: A new `tPool` instance.
func newPool(aSize int, newFunc func() any) *tPool {
	if 0 >= aSize {
		aSize = poolInitSize
	}
	return &tPool{
		items: make(chan any, aSize),
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
	select {
	case item := <-p.items:
		return item
	default:
		p.created.Add(1)
		return p.New()
	}
} // Get()

// `Metrics()` returns the current pool metrics.
//
// Returns:
//   - `rCreated`: Number of nodes created by the pool.
//   - `rReturned`: Number of nodes returned to the pool.
//   - `rSize`: Current number of items in the pool.
func (p *tPool) Metrics() (rCreated, rReturned uint32, rSize int) {
	rCreated, rReturned, rSize = p.created.Load(),
		p.returned.Load(),
		len(p.items)

	return
} // Metrics()

// `Put()` returns an item to the pool.
//
// If the pool is full, the item is dropped.
//
// Parameters:
//   - `x`: The item to return to the pool.
func (p *tPool) Put(x any) {
	if (p.returned.Add(1) & poolDropMask) == poolDropMask {
		// Drop the node if the drop mask matches.
		// This leaves the given `aNode` for GC.
		// With a drop mask of `7` (111) we drop 1 in 8 nodes.
		return
	}

	select {
	case p.items <- x:
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
func newNode() (rNode *tNode) {
	var ok bool
	if rNode, ok = nodePool.Get().(*tNode); ok {
		// Clear/reset the old field values
		if 0 < len(rNode.tChildren) {
			rNode.tChildren = make(tChildren)
		}
		rNode.hits.Store(0)
		rNode.isEnd, rNode.isWild = false, false
	} else {
		rNode = &tNode{tChildren: make(tChildren)}
	}

	return
} // newNode()

// ---------------------------------------------------------------------------
// Helper function:

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
