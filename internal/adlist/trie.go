/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"context"
	"io"
	"runtime"
	"sync/atomic"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `tTrieMetrics` contains the metrics data of the trie.
	tTrieMetrics struct {
		numNodes    atomic.Uint32
		numPatterns atomic.Uint32
		numHits     atomic.Uint32
		numMisses   atomic.Uint32
		numReloads  atomic.Uint32
		numRetries  atomic.Uint32
	}

	//
	// `tTrie` is a thread-safe trie for FQDN wildcards. It
	// basically provides a CRUD interface for FQDN patterns.
	//
	//   - `C`: Create a new pattern [Add],
	//   - `R`: Retrieve a pattern [Match],
	//   - `U`: Update a pattern [Update],
	//   - `D`: Delete a pattern [Delete].
	tTrie struct {
		_            struct{}
		tTrieMetrics // embedded metrics for the trie
		root         *tNode
	}
)

// ---------------------------------------------------------------------------
// `tTrie` constructor:

// `newTrie()` creates a new `tTrie` instance.
//
// Returns:
//   - `*tTrie`: A new `tTrie` instance.
func newTrie() *tTrie {
	return &tTrie{root: newNode()}
} // newTrie()

// ---------------------------------------------------------------------------
// `tTrie` methods:

// `Add()` inserts an FQDN pattern (with optional wildcard) into the list.
//
// If `aPattern` is an empty string, the method returns `false`.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aPattern`: The FQDN pattern to insert.
//
// Returns:
//   - `rOK`: `true` if the pattern was added, `false` otherwise.
func (t *tTrie) Add(aCtx context.Context, aPattern string) (rOK bool) {
	if nil == t || nil == t.root {
		return
	}

	parts := pattern2parts(aPattern)
	if 0 == len(parts) {
		return
	}
	// Check for timeout or cancellation
	if nil != aCtx.Err() {
		return
	}

	t.root.Lock()
	rOK = t.root.add(aCtx, parts)
	t.root.Unlock()

	return
} // Add()

// `AllPatterns()` returns all patterns in the trie.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//
// Returns:
//   - `rList`: A list of all patterns in the trie.
func (t *tTrie) AllPatterns(aCtx context.Context) (rList tPartsList) {
	if nil == t || nil == t.root || (0 == len(t.root.tChildren)) {
		return
	}
	// Check for timeout or cancellation
	if nil != aCtx.Err() {
		return
	}

	t.root.RLock()
	rList = t.root.allPatterns(aCtx)
	t.root.RUnlock()

	return
} // AllPatterns()

// `clone()` returns a deep copy of the trie.
//
// Returns:
//   - `*tTrie`: A deep copy of the trie.
func (t *tTrie) clone() *tTrie {
	clone := newTrie()
	clone.root = t.root.clone()

	return clone
} // clone()

// `Count()` returns the number of nodes and patterns in the trie.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//
// Returns:
//   - `rNodes`: The number of nodes in the trie.
//   - `rPatterns`: The number of patterns in the trie.
func (t *tTrie) Count(aCtx context.Context) (rNodes, rPatterns int) {
	if nil == t || nil == t.root {
		return
	}
	// Check for timeout or cancellation
	if nil != aCtx.Err() {
		return
	}

	t.root.RLock()
	rNodes, rPatterns = t.root.count(aCtx)
	t.root.RUnlock()

	return
} // Count()

// `Delete()` removes a pattern (FQDN or wildcard) from the list.
//
// The method returns a boolean value indicating whether the pattern
// was found and deleted.
//
// If `aPattern` is an empty string, the method returns `false`.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aPattern`: The pattern to remove.
//
// Returns:
//   - `bool`: `true` if the pattern was found and deleted, `false` otherwise.
func (t *tTrie) Delete(aCtx context.Context, aPattern string) (rOK bool) {
	if nil == t || nil == t.root {
		return
	}

	parts := pattern2parts(aPattern) // reversed list of parts
	if 0 == len(parts) {
		return
	}
	// Check for timeout or cancellation
	if nil != aCtx.Err() {
		return
	}

	// To delete an FQDN or wildcard entry from your trie-based list:
	//
	// - Traverse the trie using the reversed labels of the entry.
	// - Unmark the terminal node’s isEnd or isWild flags.
	// - Recursively prune nodes that are no longer needed (i.e.,
	// nodes that are not terminal and have no children).

	t.root.Lock()
	rOK = t.root.delete(aCtx, parts)
	t.root.Unlock()

	return
} // Delete()

// `Equal()` checks whether the trie is equal to another one.
//
// Parameters:
//   - `aTrie`: The trie to compare with.
//
// Returns:
//   - `bool`: `true` if the trie is equal to the other one, `false` otherwise.
func (t *tTrie) Equal(aTrie *tTrie) bool {
	if nil == t {
		return (nil == aTrie)
	}
	if nil == aTrie {
		return false
	}
	if t == aTrie {
		return true
	}
	if nil == t.root {
		return (nil == aTrie.root)
	}
	if nil == aTrie.root {
		return false
	}

	t.root.RLock()
	result := t.root.Equal(aTrie.root)
	t.root.RUnlock()

	return result
} // Equal()

// `ForEach()` calls the given function for each node in the trie.
//
// Since all fields of the nodes in this trie are private, this method
// doesn't provide access to the node's data. Its only use from outside
// this package would be to gather statistics.
//
// The given `aFunc()` is called in a locked R/O context for each node in
// the trie. That means that `aFunc()` can safely access the node's public
// `String()` method while all of the node's internal fields remain private
// (i.e. inaccessible).
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aFunc`: The function to call for each node.
func (t *tTrie) ForEach(aCtx context.Context, aFunc func(aNode *tNode)) {
	if (nil == t) || (nil == t.root) || (nil == aFunc) {
		return
	}
	// Check for timeout or cancellation
	if nil != aCtx.Err() {
		return
	}

	t.root.RLock()
	t.root.forEach(aCtx, aFunc)
	t.root.RUnlock()
} // ForEach()

// // `Hits()` returns the number of hits on the trie nodes.
// //
// // Returns:
// //   - `uint32`: The number of hits on the node.
// func (t *tTrie) Hits() uint32 {
// 	if (nil == t) || (nil == t.root) {
// 		return 0
// 	}

// 	return t.numHits.Load()
// } // Hits()

// `Load()` reads hostname patterns (FQDN or wildcards) from the reader
// and inserts them into the list.
//
// The given reader is expected to return one pattern per line. The
// method ignores empty lines and comment lines (starting with `#` or
// `;`). No attempt is made to validate the patterns regardless of FQDN
// or wildcard syntax, neither are the patterns checked for invalid
// characters or invalid endings.
//
// The method returns an error, if any. If it returns an error, the
// loading process has encountered a problem while reading the patterns
// and the trie may have not loaded all patterns.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aReader`: The reader to read the patterns from.
//
// Returns:
//   - `error`: `nil` if the patterns were read successfully, the error otherwise.
//     see [Store]
func (t *tTrie) Load(aCtx context.Context, aReader io.Reader) error {
	if (nil == t) || (nil == t.root) || (nil == aReader) {
		return ErrNodeNil
	}
	// Check for timeout or cancellation
	if err := aCtx.Err(); nil != err {
		return err
	}

	t.root.Lock()
	err := t.root.load(aCtx, aReader)
	t.root.Unlock()

	return err
} // Load()

// `Match()` checks if the given hostname matches any pattern in the list.
//
// If aHostname is an empty string, the method returns `false`.
//
// The given hostname is matched against the patterns in the list
// in a case-insensitive manner.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aHostPattern`: The hostname to check.
//
// Returns:
//   - `rOK`: `true` if the hostname matches any pattern, `false` otherwise.
func (t *tTrie) Match(aCtx context.Context, aHostPattern string) (rOK bool) {
	if nil == t || nil == t.root {
		return
	}

	parts := pattern2parts(aHostPattern)
	if 0 == len(parts) {
		return
	}

	// Check for timeout or cancellation
	if nil != aCtx.Err() {
		return
	}

	t.root.RLock()
	rOK = t.root.match(aCtx, parts)
	t.root.RUnlock()

	if rOK {
		t.numHits.Add(1)
	} else {
		t.numMisses.Add(1)
	}

	return
} // Match()

// `Metrics()` returns the current metrics data.
//
// Returns:
//   - `*TMetrics`: Current metrics data.
func (t *tTrie) Metrics() *TMetrics {
	if (nil == t) || (nil == t.root) {
		return nil
	}

	// Get the pool metrics before we do anything else
	pm := poolMetrics()

	// Force a garbage collection cycle
	runtime.GC()
	runtime.Gosched()
	// Read full memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	t.root.RLock()
	nodes, patterns := t.root.count(context.TODO())
	t.root.RUnlock()
	t.numNodes.Store(uint32(nodes))       //#nosec G115
	t.numPatterns.Store(uint32(patterns)) //#nosec G115

	return &TMetrics{
		PoolCreations:  pm.Created,
		PoolReturns:    pm.Returned,
		PoolSize:       pm.Size,
		Nodes:          t.numNodes.Load(),
		Patterns:       t.numPatterns.Load(),
		Hits:           t.numHits.Load(),
		Misses:         t.numMisses.Load(),
		Reloads:        t.numReloads.Load(),
		Retries:        t.numRetries.Load(),
		HeapAllocs:     m.Mallocs,
		HeapFrees:      m.Frees,
		GCPauseTotalNs: m.PauseTotalNs,
	}
} // Metrics()

// `Store()` writes all patterns currently in the list to the writer,
// one per line.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aWriter`: The writer to write the patterns to.
//
// Returns:
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
//     see [Load]
func (t *tTrie) Store(aCtx context.Context, aWriter io.Writer) error {
	if (nil == t) || (nil == t.root) || (nil == aWriter) {
		return ErrNodeNil
	}
	// Check for timeout or cancellation
	if err := aCtx.Err(); nil != err {
		return err
	}

	t.root.RLock()
	err := t.root.store(aCtx, aWriter, "")
	t.root.RUnlock()

	return err
} // Store()

// `String()` implements the `fmt.Stringer` interface for the trie.
//
// Returns:
//   - `string`: The string representation of the trie.
func (t *tTrie) String() (rStr string) {
	if (nil == t) || (nil == t.root) {
		return ErrNodeNil.Error()
	}

	t.root.RLock()
	rStr = t.root.string("Trie")
	t.root.RUnlock()

	return
} // String()

// `Update()` replaces an old pattern with a new one.
//
// The method first adds the new pattern and tries to delete the old one.
// If the new pattern couldn't be added, the old one is not deleted. The
// deletion of the old pattern might fail if it is part of a longer
// pattern, but that's not a problem as the new pattern is already in
// place. However, as long as the new pattern could be added the method
// returns `true`.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aOldPattern`: The old pattern to replace.
//   - `aNewPattern`: The new pattern to replace the old one with.
//
// Returns:
//   - `rOK`: `true` if the pattern was updated, `false` otherwise.
func (t *tTrie) Update(aCtx context.Context, aOldPattern, aNewPattern string) (rOK bool) {
	if nil == t || nil == t.root {
		return
	}

	oldParts := pattern2parts(aOldPattern) // reversed list of parts
	if 0 == len(oldParts) {
		return
	}

	newParts := pattern2parts(aNewPattern)
	if 0 == len(newParts) {
		return
	}

	if oldParts.Equal(newParts) {
		return
	}

	// Check for timeout or cancellation
	if nil != aCtx.Err() {
		return
	}

	t.root.Lock()
	rOK = t.root.update(aCtx, oldParts, newParts)
	t.root.Unlock()

	return
} // Update()

/* _EoF_ */
