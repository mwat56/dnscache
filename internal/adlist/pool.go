/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"sync"

	"github.com/mwat56/dnscache/internal/nodepool"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

var (
	// `adNodePool` is the active pool of `tNode` instances.
	adNodePool *nodepool.TPool

	// `adNodePoolInit` makes sure, the node pool is initialised only once.
	adNodePoolInit sync.Once
)

// ---------------------------------------------------------------------------
// Initialise the node pool:

func init() {
	initADnodePool()
} // init()

// `initADnodePool()` pre-allocates some nodes for the pool.
//
// This function is called only once during package initialisation.
//
// During unit testing, this function could be called manually.
func initADnodePool() {
	adNodePoolInit.Do(func() {
		adNodePool = &nodepool.TPool{
			New: func() any {
				return &tNode{tChildren: make(tChildren)}
			},
		}
		_ = adNodePool.Put(nil) // reset the pool
	})
} // initADnodePool()

// ---------------------------------------------------------------------------
// `tNode` constructor:

// `newNode()` creates a new `tNode` instance.
//
// Returns:
//   - `*tNode`: A new `tNode` instance.
func newNode() (rNode *tNode) {
	if nil == adNodePool {
		initADnodePool() // lazy initialisation
	}

	item, err := adNodePool.Get()
	if nil != err {
		rNode = &tNode{tChildren: make(tChildren)}
	} else {
		var ok bool
		if rNode, ok = item.(*tNode); ok {
			// Clear/reset the old field values
			if 0 < len(rNode.tChildren) {
				rNode.tChildren = make(tChildren)
			}
			rNode.terminator = 0
		}
	}

	return
} // newNode()

// `putNode()` throws a node back into the pool.
//
// Parameters:
//   - `aNode`: The node to return to the pool.
func putNode(aNode *tNode) {
	if nil == adNodePool {
		initADnodePool() // lazy initialisation
	}

	// We can't clear the node's fields yet since it might
	// still be used by another list or goroutine.
	_ = adNodePool.Put(aNode) // ignore the (here impossible) error
} // putNode()

// ---------------------------------------------------------------------------
// Helper/mapper function:

// `adPoolMetrics()` returns the current pool metrics.
//
// Returns:
//   - `*nodepool.TPoolMetrics`: Current pool metrics.
func adPoolMetrics() (rMetrics *nodepool.TPoolMetrics) {
	if nil == adNodePool {
		initADnodePool() // lazy initialisation
	}
	rMetrics, _ = adNodePool.Metrics()

	return
} // adPoolMetrics()

/* _EoF_ */
