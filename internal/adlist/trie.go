/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"context"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
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
	// `tRoot` is the root node of the trie.
	//
	// The root node is a special case as it doesn't have a label but
	// can have multiple children (i.e. the TLDs). Also it provides the
	// Mutex to use for locking access to the trie.
	tRoot struct {
		sync.RWMutex // barrier for concurrent access
		node         *tNode
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
		_            struct{}  // placeholder for embedding
		tTrieMetrics           // embedded metrics for the trie
		lastLoadTime time.Time // time of the trie's file loading
		filename     string    // filename for local storage
		url          string    // URL for the upstream source
		root         tRoot     // root node of the trie
	}
)

// ---------------------------------------------------------------------------
// `tTrie` constructor:

// `newTrie()` creates a new `tTrie` instance.
//
// Returns:
//   - `*tTrie`: A new `tTrie` instance.
func newTrie() *tTrie {
	return &tTrie{
		lastLoadTime: time.Now(), // time of the trie's file loading
		filename:     "Trie.txt", // default filename for local storage
		root: tRoot{
			node: newNode(),
		}, // root node of the trie
	}
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
	if nil == t {
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
	rOK = t.root.node.add(aCtx, parts)
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
	if (nil == t) || (nil == t.root.node) || (0 == len(t.root.node.tChildren)) {
		return
	}
	// Check for timeout or cancellation
	if nil != aCtx.Err() {
		return
	}

	t.root.RLock()
	rList = t.root.node.allPatterns(aCtx)
	t.root.RUnlock()

	return
} // AllPatterns()

/*
// `clone()` returns a deep copy of the trie.
//
// Returns:
//   - `*tTrie`: A deep copy of the trie.
func (t *tTrie) clone() *tTrie {
	clone := newTrie()

	t.tRoot.RLock()
	clone.tRoot = t.tRoot.clone()
	t.tRoot.RUnlock()

	return clone
} // clone()
*/

// `Count()` returns the number of nodes and patterns in the trie.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//
// Returns:
//   - `rNodes`: The number of nodes in the trie.
//   - `rPatterns`: The number of patterns in the trie.
func (t *tTrie) Count(aCtx context.Context) (rNodes, rPatterns int) {
	if nil == t || nil == t.root.node {
		return
	}
	// Check for timeout or cancellation
	if nil != aCtx.Err() {
		return
	}

	t.root.RLock()
	rNodes, rPatterns = t.root.node.count(aCtx)
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
	if nil == t || nil == t.root.node {
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
	rOK = t.root.node.delete(aCtx, parts)
	t.root.Unlock()

	return
} // Delete()

// `Equal()` checks whether the trie is equal to another one.
//
// NOTE: This method is of no practical use apart from unit-testing.
//
// Parameters:
//   - `aTrie`: The trie to compare with.
//
// Returns:
//   - `rOK`: `true` if the trie is equal to the other one, `false` otherwise.
func (t *tTrie) Equal(aTrie *tTrie) (rOK bool) {
	if nil == t {
		return (nil == aTrie)
	}
	if nil == aTrie {
		return
	}
	if t == aTrie {
		return true
	}

	t.root.RLock()
	rOK = t.root.node.Equal(aTrie.root.node)
	t.root.RUnlock()

	return
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
	if (nil == t) || (nil == t.root.node) || (nil == aFunc) {
		return
	}
	// Check for timeout or cancellation
	if nil != aCtx.Err() {
		return
	}

	t.root.RLock()
	t.root.node.forEach(aCtx, aFunc)
	t.root.RUnlock()
} // ForEach()

// `loadLocal()` reads hostname patterns (FQDN or wildcards) from `aFilename`
// and inserts them into the current trie.
//
// The method ignores empty lines and comment lines (starting with `#` or
// `;`). No attempt is made to validate the patterns regardless of FQDN or
// wildcard syntax, neither are the patterns checked for invalid characters
// or invalid endings.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aFilename`: The absolute path/name to read the patterns from.
//
// Returns:
//   - `error`: `nil` if the patterns were read successfully, the error otherwise.
func (t *tTrie) loadLocal(aCtx context.Context, aFilename string) (rErr error) {
	if nil == t {
		return ErrListNil
	}
	// The arguments are already checked by the calling `TADlist`,
	// so we can skip that here.
	if rErr = aCtx.Err(); nil != rErr {
		return
	}

	loader := &tSimpleLoader{}
	t.root.Lock()
	if rErr = loader.Load(aCtx, aFilename, t.root.node); nil == rErr {
		t.lastLoadTime = time.Now()
		t.filename = aFilename
		t.url = ""
	}
	t.root.Unlock()

	return
} // loadLocal()

const (
	downExt  = ".down"
	localExt = ".local"
)

// `loadRemote()` reads hostname patterns (FQDN or wildcards) from `aFilename`
// and inserts them into the trie.
//
// The method ignores empty lines and comment lines (starting with `#` or
// `;`). No attempt is made to validate the patterns regardless of FQDN or
// wildcard syntax, neither are the patterns checked for invalid characters
// or invalid endings.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aURL`: The URL to download the file from.
//   - `aFilename`: The absolute path/name to read the patterns from.
//
// Returns:
//   - `error`: `nil` if the patterns were read successfully, the error otherwise.
func (t *tTrie) loadRemote(aCtx context.Context, aURL, aFilename string) (rErr error) {
	if nil == t {
		return ErrListNil
	}
	// The arguments are already checked by the calling `TADlist`,
	// so we can skip that here.
	if rErr = aCtx.Err(); nil != rErr {
		return
	}
	var (
		mime   string
		loader ILoader
		saver  ISaver = &tSimpleSaver{}
	)

	//TODO: Check whether there's a local copy of the file to download and
	// use that instead of downloading it again. Consult the `lastLoadTime`
	// Trie field and compare it with the file's modification time.

	if aFilename, rErr = downloadFile(aURL, aFilename+downExt); nil != rErr {
		return
	}
	if rErr = aCtx.Err(); nil != rErr {
		return
	}
	// Check file type and use appropriate loader
	if mime, rErr = detectFileType(aFilename); nil != rErr {
		return
	}

	switch mime {
	case "text/x-abp":
		loader = &tABPLoader{}
	case "text/x-hosts":
		loader = &tHostsLoader{}
	case "text/x-hostnames":
		loader = &tSimpleLoader{}
	default:
		_ = os.Remove(aFilename)
		return ErrUnsupportedMime
	}

	// Load the file into a new trie
	newRoot := newTrie()
	if rErr = loader.Load(aCtx, aFilename, newRoot.root.node); nil != rErr {
		return
	}
	if rErr = aCtx.Err(); nil != rErr {
		return
	}

	//TODO: remove `downExt` from `aFilename`
	aFilename = strings.TrimSuffix(aFilename, downExt)
	localName := aFilename + localExt

	// Store the new trie in a local file
	localFile, err := os.OpenFile(localName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) //#nosec g304
	if nil != err {
		rErr = err
		return
	}
	defer localFile.Close()

	if rErr = saver.Save(aCtx, localFile, newRoot.root.node); nil != rErr {
		return
	}
	if rErr = aCtx.Err(); nil != rErr {
		return
	}

	newRoot.lastLoadTime = time.Now()
	newRoot.filename = aFilename
	newRoot.url = aURL
	t.root.Lock()
	t.root.node = newRoot.root.node
	t.root.Unlock()

	return
} // loadRemote()

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
	if (nil == t) || (nil == t.root.node) {
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
	rOK = t.root.node.match(aCtx, parts)
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
	if (nil == t) || (nil == t.root.node) {
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
	nodes, patterns := t.root.node.count(context.TODO())
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

// `storeFile()` writes all patterns currently in the trie to the file.
//
// The function uses a temporary file to write the patterns to, and then
// renames it to the target filename. This way, the target file is always
// either empty or contains a valid list of patterns. If `aFilename`
// already exists, it is replaced.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aFilename`: The absolute path/name to write the patterns to.
//
// Returns:
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
func (t *tTrie) storeFile(aCtx context.Context, aFilename string) error {
	if (nil == t) || (nil == t.root.node) {
		return ErrListNil
	}

	tmpName := aFilename + "~"
	if _, err := os.Stat(tmpName); nil == err {
		_ = os.Remove(tmpName)
	}

	// Check for timeout or cancellation
	if err := aCtx.Err(); nil != err {
		return err
	}

	// Create the temporary file
	file, err := os.OpenFile(tmpName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0660) //#nosec G302 G304
	if nil != err {
		_ = os.Remove(tmpName)
		return err
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(aCtx, time.Second<<2)
	defer cancel() // Ensure cancel is called

	t.root.RLock()
	err = t.root.node.store(ctx, file)
	t.root.RUnlock()

	if nil != err {
		_ = os.Remove(tmpName)
	} else {
		// Replace `aFilename` if it exists
		err = os.Rename(tmpName, aFilename)
	}

	return err
} // storeFile()

// `String()` implements the `fmt.Stringer` interface for the trie.
//
// Returns:
//   - `string`: The string representation of the trie.
func (t *tTrie) String() (rStr string) {
	if (nil == t) || (nil == t.root.node) {
		return ErrNodeNil.Error()
	}

	t.root.RLock()
	rStr = t.root.node.string("Trie")
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
	if (nil == t) || (nil == t.root.node) {
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
	rOK = t.root.node.update(aCtx, oldParts, newParts)
	t.root.Unlock()

	return
} // Update()

/* _EoF_ */
