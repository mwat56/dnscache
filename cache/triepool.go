/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"sync"

	np "github.com/mwat56/dnscache/internal/nodepool"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

var (
	// `trieNodePool` is the active pool of `tTrieNode` instances.
	trieNodePool *np.TPool

	// `trieNodePoolInit` makes sure, the node pool is initialised only once.
	trieNodePoolInit sync.Once
)

// ---------------------------------------------------------------------------
// Initialise the node pool:

func init() {
	initTriePool()
} // init()

// `initTriePool()` pre-allocates some nodes for the pool.
//
// This function is called only once during package initialisation.
//
// During unit testing, this function could be called manually.
func initTriePool() {
	trieNodePoolInit.Do(func() {
		trieNodePool = np.Init(func() any {
			return &tTrieNode{tChildren: make(tChildren)}
		}, 0)
	})
} // initTriePool()

// ---------------------------------------------------------------------------
// `tTrieNode` constructor:

// `newTrieNode()` creates a new `tTrieNode` instance.
//
// Returns:
//   - `*tTrieNode`: A new `tTrieNode` instance.
func newTrieNode() (rNode *tTrieNode) {
	if nil == trieNodePool {
		initTriePool() // lazy initialisation
	}

	item, err := trieNodePool.Get()
	if nil != err {
		rNode = &tTrieNode{tChildren: make(tChildren)}
	} else {
		var ok bool
		if rNode, ok = item.(*tTrieNode); ok {
			if nil == rNode {
				// Uninitialised pool during testing
				rNode = &tTrieNode{tChildren: make(tChildren)}
				return
			}
			// Clear/reset the old field values
			rNode.tCachedIP = tCachedIP{}
			if 0 < len(rNode.tChildren) {
				rNode.tChildren = make(tChildren)
			}
		}
	}

	return
} // newTrieNode()

// `putNode()` throws a node back into the pool.
//
// Parameters:
//   - `aNode`: The node to return to the pool.
func putNode(aNode *tTrieNode) {
	if nil == trieNodePool {
		initTriePool() // lazy initialisation
	}

	// We can't clear the node's fields yet since it might
	// still be used by another trie or goroutine.
	_ = trieNodePool.Put(aNode) // ignore the (here impossible) error
} // putNode()

/* _EoF_ */
