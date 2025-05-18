/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `tChildren` is a map of children nodes.
	tChildren map[string]*tNode

	// `tPartsList` is a reversed list of a hostname's parts.
	tPartsList []string

	// `tNode` represents a node in the trie.
	//
	// The node is a leaf node if `isEnd` is `true` and it is a wildcard
	// node if `isWild` is `true`.
	tNode struct {
		sync.RWMutex        // barrier for concurrent access
		tChildren           // children nodes
		hits         uint32 // number of hits by `match()`
		isEnd        bool   // `true` if the node is a leaf node
		isWild       bool   // `true` if the node is a wildcard node
	}

	// `tTrie` is a thread-safe trie for FQDN wildcards. It
	// basically provides a CRUD interface for FQDN patterns.
	//
	//   - `C`: Create a new pattern [Add],
	//   - `R`: Retrieve a pattern [Match],
	//   - `U`: Update a pattern [Update],
	//   - `D`: Delete a pattern [Delete].
	tTrie struct {
		root *tNode
	}
)

// --------------------------------------------------------------------------
// internal node pool functionality:

var (
	// `nodePool` is an internal pool of `tNode` instances.
	nodePool = sync.Pool{
		New: func() any {
			return &tNode{tChildren: make(tChildren)}
		},
	} // postPool
)

// `init()` pre-allocates some nodes for the pool.
func init() {
	for range 255 {
		nodePool.Put(&tNode{tChildren: make(tChildren)})
	}
} // init()

// `putNode()` returns a node to the pool.
//
// Parameters:
//   - `aPost`: The node to return to the pool.
func putNode(aNode *tNode) {
	// We can't clear the fields yet since it might
	// still be used by another list or goroutine.
	nodePool.Put(aNode)
} // putNode()

// ---------------------------------------------------------------------------
// `tPartsList` methods:

// `Equal()` checks whether the current parts list is equal to the given one.
//
// Parameters:
//   - `aPartsList`: The parts list to compare with.
//
// Returns:
//   - `bool`: `true` if the parts list is equal to the given one, `false` otherwise.
func (pl tPartsList) Equal(aPartsList tPartsList) bool {
	if nil == pl {
		return (nil == aPartsList)
	}
	if nil == aPartsList {
		return false
	}
	if len(pl) != len(aPartsList) {
		return false
	}

	return slices.Equal(pl, aPartsList)
} // Equal()

// `String()` implements the `fmt.Stringer` interface for the parts list.
//
// Returns:
//   - `string`: The string representation of the parts list.
func (pl tPartsList) String() string {
	return fmt.Sprintf("%v", []string(pl))
} // String()

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
		rNode.hits, rNode.isEnd, rNode.isWild = 0, false, false
	} else {
		rNode = &tNode{tChildren: make(tChildren)}
	}

	return
} // newNode()

// ---------------------------------------------------------------------------
// `tNode` methods:

// `add()` recursively adds a pattern to the node's tree.
//
// Parameters:
//   - `aPartsList`: The list of parts of the pattern to add.
//   - `aDepth`: The current depth in the parts list.
//
// Returns:
//   - `bool`: `true` if a pattern was added, `false` otherwise.
func (n *tNode) add(aPartsList tPartsList, aDepth int) bool {
	if (nil == n) || (0 == len(aPartsList)) {
		return false
	}

	if aDepth >= len(aPartsList) {
		// We are at the end of the patterns to insert
		// and reached the final target node

		// Locking is done by the calling `tTrie`
		n.isWild = ("*" == aPartsList[aDepth-1])
		n.isEnd = (!n.isWild)

		return true // Once we're here, we added something
	}

	label := aPartsList[aDepth]
	// Locking is done by the calling `tTrie`
	child, ok := n.tChildren[label]

	if ok {
		// Finalise the child node
		return child.add(aPartsList, aDepth+1)
	}

	child = newNode() // We need a new child node
	if "*" == label { // wildcard in the middle ?!
		child.isWild = true
	}
	added := child.add(aPartsList, aDepth+1)

	// it's a new child, so we need to add it
	// Locking is done by the calling `tTrie`
	n.tChildren[label] = child
	if 0 < len(n.tChildren) {
		// Reset the finaliser flags
		n.isWild, n.isEnd = false, false
	}

	return added
} // add()

// `clone()` returns a deep copy of the node.
//
// Returns:
//   - `*tNode`: A deep copy of the node.
func (n *tNode) clone() *tNode {
	if nil == n {
		return nil
	}

	clone := &tNode{
		tChildren: make(tChildren),
		isEnd:     n.isEnd,
		isWild:    n.isWild,
	}

	// Clone the children nodes
	for label, child := range n.tChildren {
		clone.tChildren[label] = child.clone()
	}

	return clone
} // clone()

// `delete()` recursively removes patterns from the node's tree.
//
// The method returns `true` if the node is deleted, `false` otherwise.
// A wildcard pattern has no special meaning here but is just another
// label to match.
//
// Parameters:
//   - `aPartsList`: The list of parts of the pattern to delete.
//   - `aDepth`: The current depth in the node's trie.
//
// Returns:
//   - `bool`: `true` if the node is deleted, `false` otherwise.
func (n *tNode) delete(aPartsList tPartsList, aDepth int) bool {
	if (nil == n) || (0 == len(aPartsList)) {
		return false
	}

	if aDepth >= len(aPartsList) {
		// We are at the end of the patterns to delete
		if 0 < len(n.tChildren) {
			// Node has other children apart from the one
			// we're about to delete, so we're not done yet
			n.isEnd, n.isWild = false, false

			return false
		}

		// Node has no more matching children, so we're done
		n.isWild = ("*" == aPartsList[aDepth-1])
		n.isEnd = (!n.isWild)

		return true // Once we're here, we deleted something
	}

	label := aPartsList[aDepth]
	// Locking is done by the calling `tTrie`
	child, ok := n.tChildren[label]
	if !ok {
		// No child with that name to delete
		return false
	}

	// Recursively delete the child node
	if child.delete(aPartsList, aDepth+1) {
		// Cleanup if child became deletable: Only remove it if it
		// has no children apart from the one we just deleted before
		if 0 == len(child.tChildren) {
			delete(n.tChildren, label)
			putNode(child) // Return the child to the pool
			if 0 == len(n.tChildren) {
				n.isWild = ("*" == label)
				n.isEnd = (!n.isWild)
			}
		}
		return true
	}

	return false
} // delete()

// `Equal()` checks whether the current node is equal to the given one.
//
// Parameters:
//   - `aNode`: The node to compare with.
//
// Returns:
//   - `bool`: `true` if the node is equal to the given one, `false` otherwise.
func (n *tNode) Equal(aNode *tNode) bool {
	if nil == n {
		return (nil == aNode)
	}
	if nil == aNode {
		return false
	}
	if n == aNode {
		return true
	}
	// The node's own lock is done by the calling `tTrie`
	aNode.RLock()
	defer aNode.RUnlock()

	if n.isEnd != aNode.isEnd {
		return false
	}
	if n.isWild != aNode.isWild {
		return false
	}
	if len(n.tChildren) != len(aNode.tChildren) {
		return false
	}

	for label, child := range n.tChildren {
		otherChild, ok := aNode.tChildren[label]
		if !ok {
			return false
		}
		if !child.Equal(otherChild) {
			return false
		}
	}

	return true
} // Equal()

// `forEach()` calls the given function for each node in the trie.
//
// The given `aFunc()` is called by the owning trie in a locked R/O context
// for each node in the trie.
//
// Since all fields of all sub-nodes of the current node are private, this
// method doesn't provide access to a node's data. Its only use from outside
// this package would be to gather statistics for example by calling a node's
// public methods like `Hits()` or `String()`.
//
// Parameters:
//   - `aFunc`: The function to call for each node.
func (n *tNode) forEach(aFunc func(aNode *tNode)) {
	if (nil == n) || (nil == aFunc) {
		return
	}
	// Locking is done by the calling `tTrie`
	aFunc(n)

	if 0 == len(n.tChildren) {
		return
	}

	var (
		child *tNode
		label string
	)

	kids := make(tPartsList, 0, len(n.tChildren))
	for label = range n.tChildren {
		kids = append(kids, label)
	}
	sort.Strings(kids)

	for _, label := range kids {
		child = n.tChildren[label]
		child.forEach(aFunc)
	}
} // forEach()

// `Hits()` returns the number of hits on the node.
//
// Returns:
//   - `uint32`: The number of hits on the node.
func (n *tNode) Hits() uint32 {
	return atomic.LoadUint32(&n.hits)
} // Hits()

// `match()` checks whether the node's tree contains the given pattern.
//
// Parameters:
//   - `aPartsList`: The list of parts of the pattern to check.
//   - `aHitCounter`: `true` if the node's number of hits should be incremented.
//
// Returns:
//   - `bool`: `true` if the pattern is in the node's trie, `false` otherwise.

// func (n *tNode) match(aPartsList tPartsList, aDepth int) bool {
// 	if (nil == n) || (0 == len(aPartsList)) {
// 		return false
// 	}
// 	if aDepth >= len(aPartsList) {
// 		// We are at the end of the patterns to check
// 		// and found the final target node
// 		if n.isWild {
// 			return true
// 		}
// 		return n.isEnd
// 	}
// 	// First check for a wildcard match at the current level,
// 	// then check for an exact match if there is no wildcard.
// 	label := aPartsList[aDepth]
// 	// Locking is done by the calling `tTrie`
// 	child, ok := n.tChildren["*"]
// 	if !ok {
// 		child, ok = n.tChildren[label]
// 	}
// 	if !ok {
// 		// No child with that name or wildcard
// 		return false
// 	}
// 	// Recursively check the child node
// 	return child.match(aPartsList, aDepth+1)
// } // match()

func (n *tNode) match(aPartsList tPartsList, aHitCounter bool) bool {
	if (nil == n) || (0 == len(aPartsList)) {
		return false
	}
	if 0 == len(n.tChildren) {
		// No children, thus no match
		return false
	}

	var ( // avoid repeated allocations inside the loop
		child   *tNode
		depth   int
		label   string
		matched bool
		ok      bool
	)

	startNode := n // Remember the start node for un-/locking
	startNode.RLock()
	for depth, label = range aPartsList {
		child, ok = n.tChildren[label]
		if !ok {
			// No child with that name, check for wildcard
			if child, ok = n.tChildren["*"]; ok {
				matched = (child.isWild || child.isEnd)
			} else {
				// Remember the last non-wildcard node
				child = n
			}
			break
		}

		n = child
		// First check for a wildcard match at the current level,
		if child, ok = n.tChildren["*"]; ok {
			matched = (child.isWild || child.isEnd)
			break
		} else {
			// Remember the last non-wildcard node
			child = n
		}
		if n.isWild || (len(aPartsList)-1 == depth) && n.isEnd {
			matched = true
			break
		}
	}
	startNode.RUnlock()

	if matched {
		if aHitCounter {
			atomic.AddUint32(&child.hits, 1)
		}

		return true
	}

	return false
} // match()

// `prepString()` prepares a string representation of the node.
//
// Parameters:
//   - `aBuilder`: The string builder to use.
//   - `aLabel`: The label of the node.
//   - `aLevel`: The level of the node.
//
// Returns:
//   - `string`: The string representation of the node.
func (n *tNode) prepString(aBuilder *strings.Builder, aLabel string, aLevel int) {
	if (nil == n) || (nil == aBuilder) {
		return
	}

	var (
		child *tNode
		label string
	)
	prefix1 := strings.Repeat("  ", aLevel+1)
	prefix2 := strings.Repeat("  ", aLevel+2)

	// Locking is done by the calling `tTrie`
	fmt.Fprintf(aBuilder, "%q:\n%sisEnd: %v\n%sisWild: %v\n%shits: %d\n",
		aLabel, prefix1, n.isEnd, prefix1, n.isWild, prefix1, n.hits)
	if 0 == len(n.tChildren) {
		return
	}

	fmt.Fprintf(aBuilder, "%sChildren:\n", prefix1)

	// Collect and sort child kids
	kids := make(tPartsList, 0, len(n.tChildren))
	for label = range n.tChildren {
		kids = append(kids, label)
	}
	sort.Strings(kids)

	// Output children in sorted order
	for _, label = range kids {
		child = n.tChildren[label]
		fmt.Fprint(aBuilder, prefix2)
		// Recursively call `prepString()` for all children
		child.prepString(aBuilder, label, aLevel+2)
	}
} // prepString()

// `String()` implements the `fmt.Stringer` interface for the node.
//
// Returns:
//   - `string`: The string representation of the node.
func (n *tNode) String() string {
	if nil == n {
		return ""
	}
	var builder strings.Builder

	n.RLock()
	n.prepString(&builder, "Node", 0)
	n.RUnlock()

	return builder.String()
} // String()

// `update()` updates a pattern in the node's tree.
//
// The method deletes the old pattern and inserts the new one. If the
// old pattern couldn't be deleted, the new one is not added and the
// method returns `false`. Also, if the new pattern couldn't be added,
// the old one is restored and the method returns `false`.
//
// Parameters:
//   - `aOldParts`: The list of parts of the old pattern to update.
//   - `aNewParts`: The list of parts of the new pattern to update to.
//
// Returns:
//   - `bool`: `true` if the pattern was updated, `false` otherwise.
func (n *tNode) update(aOldParts, aNewParts tPartsList) bool {
	if (nil == n) || (0 == len(aOldParts)) || (0 == len(aNewParts)) || aOldParts.Equal(aNewParts) {
		return false
	}
	var added bool
	// Locking is done by the calling `tTrie`

	// 1. Insert the new pattern
	if added = n.add(aNewParts, 0); added {
		// 2. Only after successful insert, delete the old pattern.
		// This might fail if the old pattern is part of a longer
		// pattern, but that's not a problem as the new pattern is
		// already in place.
		_ = n.delete(aOldParts, 0)
	}

	return added
} // update()

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
// `tTrie` helper functions:

// `pattern2parts()` converts a pattern to a reversed list of parts.
//
// The pattern is expected to be a valid FQDN or wildcard pattern, and it's
// not checked for validity. It is trimmed and converted to lower case for
// case-insensitive matching.
//
// Parameters:
//   - `aPattern`: The pattern to check and convert.
//
// Returns:
//   - `tPartsList`: The list of parts.
func pattern2parts(aPattern string) tPartsList {
	if aPattern = strings.TrimSpace(aPattern); 0 == len(aPattern) {
		return nil
	}

	parts := strings.Split(strings.ToLower(aPattern), ".")
	slices.Reverse(parts)

	return parts
} // pattern2parts()

// ---------------------------------------------------------------------------
// `tTrie` methods:

// `Add()` inserts an FQDN pattern (with optional wildcard) into the list.
//
// If `aPattern` is an empty string, the method returns `nil, false`.
//
// Parameters:
//   - `aPattern`: The FQDN pattern to insert.
//
// Returns:
//   - `*tTrie`: The updated trie.
//   - `bool`: `true` if the pattern was added, `false` otherwise.
func (t *tTrie) Add(aPattern string) (*tTrie, bool) {
	if nil == t || nil == t.root {
		return nil, false
	}

	parts := pattern2parts(aPattern)
	if 0 == len(parts) {
		return nil, false
	}

	t.root.Lock()
	added := t.root.add(parts, 0)
	t.root.Unlock()

	return t, added
} // Add()

// Collects all FQDNs and wildcards in sorted order.
// Returns them as a list of dot-separated strings in original (non-reversed) order.
func (t *tTrie) AllPatterns() tPartsList {
	if nil == t || nil == t.root || (0 == len(t.root.tChildren)) {
		return tPartsList{}
	}
	// var (
	// 	results, path tPartsList
	// 	// walk          func(n *tNode, aPath tPartsList)
	// )
	// walk = func(n *tNode, aPath tPartsList) {
	// 	// If this node marks an original pattern, collect it
	// 	if n.isEnd || n.isWild {
	// 		// Reverse the path to get the original FQDN order
	// 		reversed := make(tPartsList, len(aPath))
	// 		for idx, part := range aPath {
	// 			reversed[len(aPath)-1-idx] = part
	// 		}
	// 		results = append(results, strings.Join(reversed, "."))
	// 	}
	// 	if 0 == len(n.tChildren) {
	// 		return
	// 	}
	// 	// Collect and sort children kids
	// 	kids := make(tPartsList, 0, len(n.tChildren))
	// 	for label := range n.tChildren {
	// 		kids = append(kids, label)
	// 	}
	// 	sort.Strings(kids)
	// 	// Recurse into each child in sorted order
	// 	for _, label := range kids {
	// 		walk(n.tChildren[label], append(aPath, label))
	// 	}
	// }
	// t.root.RLock()
	// walk(t.root, path)
	// t.root.RUnlock()
	// return results

	t.root.RLock()
	results := t.root.allPatterns(tPartsList{})
	t.root.RUnlock()

	return *results
} // AllPatterns()

func (n *tNode) allPatterns(aPatternList tPartsList) *tPartsList {
	if nil == n || (0 == len(aPatternList)) {
		return nil
	}
	var (
		child    *tNode
		idx      int
		label    string
		newParts *tPartsList
		result   = &tPartsList{}
	)

	// If this node marks an original pattern, collect it
	if n.isEnd || n.isWild {
		// Reverse the path to get the original FQDN in original order
		reversed := make(tPartsList, len(aPatternList))
		for idx, label = range aPatternList {
			if 0 < len(label) {
				reversed[len(aPatternList)-1-idx] = label
			}
		}

		pattern := strings.Join(reversed, ".")
		if 0 == len(pattern) {
			return result
		}
		*result = append(*result, pattern)
	}
	cLen := len(n.tChildren)
	if 0 == cLen {
		return result
	}

	// Collect and sort children kids
	kids := make(tPartsList, 0, cLen)
	for label = range n.tChildren {
		kids = append(kids, label)
	}
	sort.Strings(kids)

	// Recurse into each child in sorted order
	for _, label = range kids {
		// Add the child's label to the pattern list
		aPatternList = append(aPatternList, label)
		child = n.tChildren[label]

		if newParts = child.allPatterns(aPatternList); (nil != newParts) &&
			(0 < len(*newParts)) {
			// Add the child's patterns to the result
			*result = append(*result, *newParts...)
		}
	}

	return result
} // allPatterns()

// `clone()` returns a deep copy of the trie.
//
// Returns:
//   - `*tTrie`: A deep copy of the trie.
func (t *tTrie) clone() *tTrie {
	clone := newTrie()
	clone.root = t.root.clone()

	return clone
} // clone()

// `Delete()` removes a pattern (FQDN or wildcard) from the list.
//
// The method returns the updated trie and a boolean value indicating
// whether the pattern was found and deleted.
//
// If `aPattern` is an empty string, the method returns `nil, false`.
//
// Parameters:
//   - `aPattern`: The pattern to remove.
//
// Returns:
//   - `*tTrie`: The updated trie.
//   - `bool`: `true` if the pattern was found and deleted, `false` otherwise.
func (t *tTrie) Delete(aPattern string) (*tTrie, bool) {
	if nil == t || nil == t.root {
		return nil, false
	}

	parts := pattern2parts(aPattern) // reversed list of parts
	if 0 == len(parts) {
		return nil, false
	}

	// To delete an FQDN or wildcard entry from your trie-based list:
	//
	// - Traverse the trie using the reversed labels of the entry.
	// - Unmark the terminal node’s isEnd or isWild flags.
	// - Recursively prune nodes that are no longer needed (i.e.,
	// nodes that are not terminal and have no children).

	t.root.Lock()
	deleted := t.root.delete(parts, 0)
	t.root.Unlock()

	return t, deleted
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
// this package would be to gather statistics for example by calling
// a node's public methods like `Hits()` or `String()`.
//
// The given `aFunc()` is called in a locked R/O context for each node in
// the trie. That means that `aFunc()` can safely access the node's public
// methods like [Hits] or [String] while all of the node's internal fields
// remain private (i.e. inaccessible).
//
// Parameters:
//   - `aFunc`: The function to call for each node.
func (t *tTrie) ForEach(aFunc func(aNode *tNode)) {
	if nil == t || nil == t.root || nil == aFunc {
		return
	}

	t.root.RLock()
	t.root.forEach(aFunc)
	t.root.RUnlock()
} // ForEach()

// `Load()` reads hostname patterns (FQDN or wildcards) from the reader
// and inserts them into the list.
//
// The given reader is expected to return one pattern per line. The
// method ignores empty lines and comment lines (starting with `#` or `;`).
// No attempt is made to validate the patterns regardless of FQDN or
// wildcard syntax, neither are the patterns checked for invalid characters
// or invalid endings.
//
// The method returns the updated trie and an error, if any. If it returns
// an error, the loading process has encountered a problem while reading the
// patterns and the trie may have not loaded all patterns.
//
// Parameters:
//   - `aReader`: The reader to read the patterns from.
//
// Returns:
//   - `*tTrie`: The updated trie.
//   - `error`: `nil` if the patterns were read successfully, the error otherwise.
//     see [Store]
func (t *tTrie) Load(aReader io.Reader) (*tTrie, error) {
	if (nil == t) || (nil == t.root) || (nil == aReader) {
		return t, nil
	}

	line := ""
	scanner := bufio.NewScanner(aReader)
	t.root.Lock()
	for scanner.Scan() {
		line = strings.TrimSpace(scanner.Text())
		// Ignore empty lines
		if 0 == len(line) {
			continue
		}
		// Ignore comment lines
		if "#" == string(line[0]) || ";" == string(line[0]) {
			continue
		}
		// Add the pattern to the trie
		t.Add(line)
	}
	t.root.Unlock()

	return t, scanner.Err()
} // Load()

// `Match()` checks if the given hostname matches any pattern in the list.
//
// If aHostname is an empty string, the method returns `nil, false`.
//
// The given hostname is matched against the patterns in the list #
// in a case-insensitive manner.
//
// Parameters:
//   - `aHostname`: The hostname to check.
//
// Returns:
//   - `*tTrie`: The matching trie.
//   - `bool`: `true` if the hostname matches any pattern, `false` otherwise.
func (t *tTrie) Match(aHostname string) (*tTrie, bool) {
	if nil == t || nil == t.root {
		return nil, false
	}
	parts := pattern2parts(aHostname)
	if 0 == len(parts) {
		return nil, false
	}

	if ok := t.root.match(parts, true); !ok {
		return nil, false
	}

	return t, true
} // Match()

// `Store()` writes all patterns currently in the list to the writer,
// one per line.
//
// Parameters:
//   - `aWriter`: The writer to write the patterns to.
//
// Returns:
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
//     see [Load]
func (t *tTrie) Store(aWriter io.Writer) error {
	if (nil == t) || (nil == t.root) || (nil == aWriter) {
		return nil
	}
	// We need this var declaration to allow for recursion
	var walk func(aNode *tNode, aPartsList tPartsList) error
	bw := bufio.NewWriter(aWriter)

	// `walk()` recursively walks the trie and writes the patterns
	// to the writer.
	walk = func(aNode *tNode, aPartsList tPartsList) error {
		for label, child := range aNode.tChildren {
			newParts := append(tPartsList{label}, aPartsList...)

			if child.isEnd || child.isWild {
				clone := slices.Clone(newParts)
				slices.Reverse(clone) //TODO: ???
				pattern := strings.Join(clone, ".")

				if _, err := bw.WriteString(pattern + "\n"); nil != err {
					return err
				}
			}

			if err := walk(child, newParts); nil != err {
				return err
			}
		}
		return nil
	} // walk()

	t.root.RLock()
	err := walk(t.root, nil)
	t.root.RUnlock()

	if nil != err {
		return err
	}

	return bw.Flush()
} // Store()

// `String()` implements the `fmt.Stringer` interface for the trie.
//
// Returns:
//   - `string`: The string representation of the trie.
func (t *tTrie) String() string {
	if (nil == t) || (nil == t.root) {
		return ""
	}
	var builder strings.Builder

	t.root.RLock()
	t.root.prepString(&builder, "Trie", 0)
	t.root.RUnlock()

	return builder.String()
} // String()

// `Update()` replaces an old pattern with a new one.
//
// Parameters:
//   - `aOldPattern`: The old pattern to replace.
//   - `aNewPattern`: The new pattern to replace the old one with.
//
// Returns:
//   - `*tTrie`: The updated trie.
func (t *tTrie) Update(aOldPattern, aNewPattern string) (*tTrie, bool) {
	if nil == t || nil == t.root {
		return nil, false
	}
	oldParts := pattern2parts(aOldPattern) // reversed list of parts
	if 0 == len(oldParts) {
		return nil, false
	}
	newParts := pattern2parts(aNewPattern)
	if 0 == len(newParts) {
		return nil, false
	}
	if oldParts.Equal(newParts) {
		return nil, false
	}

	t.root.Lock()
	updated := t.root.update(oldParts, newParts)
	t.root.Unlock()

	if updated {
		return t, updated
	}

	return nil, false
} // Update()

/* _EoF_ */
