/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"bufio"
	"errors"
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

	// `tNode` represents a node in the trie.
	//
	// The node is a leaf node if `isEnd` is `true` and it's a wildcard
	// node if `isWild` is `true`.
	tNode struct {
		sync.RWMutex        // barrier for concurrent access
		tChildren           // children nodes
		hits         uint32 // number of hits by `match()`
		isEnd        bool   // `true` if the node is a leaf node
		isWild       bool   // `true` if the node is a wildcard node
	}
)

var (
	// `ErrNodeNil` is returned if a node or a method's required
	// argument is `nil`.
	//
	// see [tTrie.Load], [tTrie.Store]
	ErrNodeNil = errors.New("node or argument is nil")
)

// ---------------------------------------------------------------------------
// `tNode` methods:

// `add()` adds a pattern to the node's tree.
//
// The method returns `true` if at least one part was added in order to
// have the whole `aPartsList` present in the trie or found an existing
// node that already represented the pattern (i.e. marked as end node
// or wildcard).
//
// Parameters:
//   - `aPartsList`: The list of parts of the pattern to add.
//
// Returns:
//   - `bool`: `true` if a pattern was added, `false` otherwise.
func (n *tNode) add(aPartsList tPartsList) bool {
	if (nil == n) || (0 == len(aPartsList)) {
		return false
	}
	var (
		idx, depth, ends int
		label            string
		ok               bool
	)

	node := n
	for idx, label = range aPartsList {
		if _, ok = node.tChildren[label]; !ok {
			node.tChildren[label] = newNode()
			depth++
		}

		// Descend into the child node
		node = node.tChildren[label]
		if "*" == label {
			node.isWild = true
			ends++
		}
		if (len(aPartsList) - 1) == idx {
			if node.isEnd = (!node.isWild); node.isEnd {
				ends++
			}
		}
	} // for parts

	return (0 < depth) || (0 < ends)
} // add()

// `allPatterns()` recursively collects all hostname patterns in the
// node's tree.
//
// The patterns are returned in sorted order.
//
// The method uses a stack to traverse the tree in a depth-first manner,
// which is more efficient than a recursive approach.
//
// Returns:
//   - `rList`: A list of all patterns in the node's tree.
func (n *tNode) allPatterns() (rList tPartsList) {
	if nil == n {
		return
	}
	type (
		tStackEntry struct {
			node *tNode     // respective node to process
			path tPartsList // path in the trie to the node
		}
	)
	var (
		cLen int
		kids tPartsList
	)
	stack := []tStackEntry{
		// Push the current node to the stack
		{node: n, path: tPartsList{}},
	}

	for 0 < len(stack) {
		// Pop the top of the stack
		current := stack[len(stack)-1]
		// Remove the top of the stack
		stack = stack[:len(stack)-1]

		// Check if current node is a terminal pattern
		if current.node.isEnd || current.node.isWild {
			// Reverse the path to get the original FQDN
			// in original order.
			if pLen := len(current.path); 0 < pLen {
				reversed := make(tPartsList, len(current.path))
				for idx, label := range current.path {
					reversed[len(current.path)-1-idx] = label
				}
				rList = append(rList, strings.Join(reversed, "."))
			}
		}

		if cLen = len(current.node.tChildren); 0 == cLen {
			continue
		}

		// Collect and sort children keys for deterministic order
		kids = make(tPartsList, 0, cLen)
		for label := range current.node.tChildren {
			kids = append(kids, label)
		}
		if 1 < len(kids) {
			sort.Strings(kids)
		}

		// Push children to stack in reverse-sorted order
		// (to process them in forward order when popped)
		for i := len(kids) - 1; 0 <= i; i-- {
			label := kids[i]
			child := current.node.tChildren[label]
			newPath := make(tPartsList, len(current.path)+1)
			copy(newPath, current.path)
			newPath[len(current.path)] = label
			stack = append(stack, tStackEntry{
				node: child,
				path: newPath,
			})
		}
	} // for stack

	return
} // allPatterns()

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

// `delete()` removes path patterns from the node's tree.
//
// The method returns `true` if at least one node is deleted, `false`
// otherwise.
// A wildcard pattern has no special meaning here but is just another
// label to match.
//
// Parameters:
//   - `aPartsList`: The list of parts of the pattern to delete.
//
// Returns:
//   - `bool`: `true` if the node is deleted, `false` otherwise.
func (n *tNode) delete(aPartsList tPartsList) (rOK bool) {
	if (nil == n) || (0 == len(aPartsList)) {
		return
	}
	type (
		tStackEntry struct {
			label string
			node  *tNode
		}
	)
	var (
		stack []tStackEntry
	)
	current := n

	// Traverse and build up the stack
	for _, part := range aPartsList {
		child, ok := current.tChildren[part]
		if !ok {
			// Pattern does not exist; nothing to delete
			return
			//TODO: continue instead?
		}
		stack = append(stack, tStackEntry{node: current, label: part})
		current = child
	}

	// Unset terminal markers at the end node
	current.isEnd, current.isWild = false, false

	// Backtrack and prune
	for i := len(stack) - 1; 0 <= i; i-- {
		parent := stack[i].node
		label := stack[i].label
		child := parent.tChildren[label]

		// If child has children stop pruning
		if 0 != len(child.tChildren) {
			break
		}
		// Safe to delete this child
		delete(parent.tChildren, label)
		putNode(child) // Return the child to the pool
		if 0 == len(parent.tChildren) {
			parent.isWild = ("*" == label) //TODO: ??
			parent.isEnd = (!parent.isWild)
		}
		rOK = true
	}

	return
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
	type tStackEntry struct {
		node *tNode
	}
	stack := []tStackEntry{{node: n}}

	for 0 < len(stack) {
		// Pop from stack
		entry := stack[len(stack)-1]
		// Remove from stack
		stack = stack[:len(stack)-1]

		aFunc(entry.node)

		// Collect and sort children kids for deterministic order
		cLen := len(entry.node.tChildren)
		if 0 == cLen {
			continue
		}

		kids := make(tPartsList, 0, cLen)
		for k := range entry.node.tChildren {
			kids = append(kids, k)
		}
		if 1 < len(kids) {
			sort.Strings(kids)
		}

		// Push children to stack in reverse-sorted order
		// (to process them in forward order when popped)
		for idx := len(kids) - 1; 0 <= idx; idx-- {
			label := kids[idx]
			stack = append(stack, tStackEntry{node: entry.node.tChildren[label]})
		}
	}
} // forEach()

// `Hits()` returns the number of hits on the node.
//
// Returns:
//   - `uint32`: The number of hits on the node.
func (n *tNode) Hits() uint32 {
	return atomic.LoadUint32(&n.hits)
} // Hits()

// `load()` reads patterns from the reader and adds them to the node's tree.
//
// Parameters:
//   - `aReader`: The reader to read the patterns from.
//
// Returns:
//   - `error`: `nil` if the patterns were read successfully, the error otherwise.
func (n *tNode) load(aReader io.Reader) error {
	if (nil == n) || (nil == aReader) {
		return ErrNodeNil
	}

	scanner := bufio.NewScanner(aReader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Ignore empty lines
		if 0 == len(line) {
			continue
		}
		// Ignore comment lines
		if "#" == string(line[0]) || ";" == string(line[0]) {
			continue
		}

		parts := pattern2parts(line)
		if 0 == len(parts) {
			continue
		}

		// Add the pattern to the node
		n.add(parts)
	}

	return scanner.Err()
} // load()

// `match()` checks whether the node's tree contains the given pattern.
//
// Parameters:
//   - `aPartsList`: The list of parts of the pattern to check.
//   - `aHitCounter`: `true` if the node's number of hits should be incremented.
//
// Returns:
//   - `bool`: `true` if the pattern is in the node's trie, `false` otherwise.
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

	if matched {
		if aHitCounter {
			atomic.AddUint32(&child.hits, 1)
		}

		return true
	}

	return false
} // match()

// `store()` writes all patterns currently in the list to the writer,
// one per line.
//
// Parameters:
//   - `aWriter`: The writer to write the patterns to.
//
// Returns:
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
func (n *tNode) store(aWriter io.Writer) error {
	if (nil == n) || (nil == aWriter) {
		return ErrNodeNil
	}
	type (
		tStackEntry struct {
			path tPartsList
			node *tNode
		}
	)

	stack := []tStackEntry{{node: n, path: tPartsList{}}}
	for 0 < len(stack) {
		// Pop from stack
		entry := stack[len(stack)-1]
		// Remove current entry from stack
		stack = stack[:len(stack)-1]

		// Process current node
		if entry.node.isEnd || entry.node.isWild {
			// Reverse path to original FQDN format
			pLen := len(entry.path)
			reversed := make(tPartsList, pLen)
			for idx, part := range entry.path {
				reversed[pLen-1-idx] = part
			}
			fqdn := strings.Join(reversed, ".")

			// Write to writer with newline
			if _, err := fmt.Fprintln(aWriter, fqdn); nil != err {
				return err
			}
		}
		cLen := len(entry.node.tChildren)
		if 0 == cLen {
			continue
		}

		// Collect and sort children kids
		kids := make(tPartsList, 0, cLen)
		for label := range entry.node.tChildren {
			kids = append(kids, label)
		}
		if 1 < len(kids) {
			sort.Strings(kids)
		}

		// Push children in reverse-sorted order for
		// correct processing sequence
		for idx := len(kids) - 1; idx >= 0; idx-- {
			label := kids[idx]
			newPath := make(tPartsList, len(entry.path)+1)
			copy(newPath, entry.path)
			newPath[len(entry.path)] = label

			stack = append(stack, tStackEntry{
				node: entry.node.tChildren[label],
				path: newPath,
			})
		}
	}

	return nil
} // store()

// `string()` returns a string representation of the node.
//
// Parameters:
//   - `aLabel`: The label to use for the current node.
//
// Returns:
//   - `string`: The string representation of the node.
func (n *tNode) string(aLabel string) string {
	if nil == n {
		return ""
	}

	type (
		tStackEntry struct {
			kids     tPartsList // sorted child keys
			name     string     // label of the node
			node     *tNode     // respective node to process
			depth    int        // depth in the tree
			childIdx int        // index of next child to process
		}
	)
	// Locking is done by the calling `tTrie`
	stack := []tStackEntry{
		{
			kids:     nil,
			name:     aLabel,
			node:     n,
			depth:    0,
			childIdx: 0,
		},
	}
	var (
		// Buffer for the string representation's parts
		builder strings.Builder
	)

	for 0 < len(stack) {
		// Get the top entry of the stack
		entry := &stack[len(stack)-1]
		indent := strings.Repeat("  ", entry.depth+1)

		// If this is the first time visiting this node,
		// print its details
		if nil == entry.kids {
			line := fmt.Sprintf("%q:\n%sisEnd: %v\n%sisWild: %v\n%shits: %d\n",
				entry.name, indent, entry.node.isEnd, indent, entry.node.isWild, indent, entry.node.hits)
			builder.WriteString(line)

			// Prepare sorted child keys
			entry.kids = make(tPartsList, 0, len(entry.node.tChildren))
			for label := range entry.node.tChildren {
				entry.kids = append(entry.kids, label)
			}
			if 1 < len(entry.kids) {
				sort.Strings(entry.kids)
			}
		}

		// If there are unprocessed children, process the next one
		if entry.childIdx < len(entry.kids) {
			// Indent for the child node
			builder.WriteString(indent)

			label := entry.kids[entry.childIdx]
			child := entry.node.tChildren[label]
			entry.childIdx++

			// Push the child node to the stack
			stack = append(stack, tStackEntry{
				kids:     nil,
				name:     label,
				node:     child,
				depth:    entry.depth + 2,
				childIdx: 0,
			})

			// Continue with the next iteration
			continue
		}

		// Pop the stack
		stack = stack[:len(stack)-1]
	} // for stack

	return builder.String()
} // string()

// `String()` implements the `fmt.Stringer` interface for the node.
//
// Returns:
//   - `string`: The string representation of the node.
func (n *tNode) String() string {
	if nil == n {
		return ""
	}
	// Locking is done by the calling `tTrie`
	return n.string("Node")
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
	if added = n.add(aNewParts); added {
		// 2. Only after successful insert, delete the old pattern.
		// This might fail if the old pattern is part of a longer
		// pattern, but that's not a problem as the new pattern is
		// already in place.
		_ = n.delete(aOldParts)
	}

	return added
} // update()

// ---------------------------------------------------------------------------
// Helper function:

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

/* _EoF_ */
