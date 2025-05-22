/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
	"sync"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// `endMask` is the bit mask to use for marking a node as a leaf node.
	endMask = 192 // 11000000

	// `wildMask` is the bit mask to use for marking a node as a wildcard node.
	wildMask = 3 // 00000011
)

type (
	// `tChildren` is a map of children nodes.
	tChildren map[string]*tNode

	// `tNode` represents a node in the trie.
	//
	// The node is a leaf node if `terminator` has the `endMask` bit set and
	// it's a wildcard node if `terminator` has the `wildMask` bit set.
	tNode struct {
		sync.RWMutex       // barrier for concurrent access
		tChildren          // children nodes
		terminator   uint8 // flags for pattern end and wildcard
	}
)

var (
	// `ErrNodeNil` is returned if a node or a method's required
	// argument is `nil`.
	ErrNodeNil = errors.New("node or argument is nil")
)

// ---------------------------------------------------------------------------
// Helper function:

// `pattern2parts()` converts a hostname pattern to a reversed list of parts.
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
// `tNode` methods:

// `add()` adds a pattern to the node's tree.
//
// The method returns `true` if at least one part was added in order to
// have the whole `aPartsList` present in the trie or found an existing
// node that already represented the pattern (i.e. marked as end node
// or wildcard).
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aPartsList`: The list of parts of the pattern to add.
//
// Returns:
//   - `bool`: `true` if a pattern was added, `false` otherwise.
func (n *tNode) add(aCtx context.Context, aPartsList tPartsList) (rOK bool) {
	if (nil == n) || (0 == len(aPartsList)) {
		return
	}
	var (
		depth, added, ends int
		isEnd, isWild, ok  bool
		label              string
	)

	node := n
	for depth, label = range aPartsList {
		if _, ok = node.tChildren[label]; !ok {
			// Check for timeout or cancellation
			if nil != aCtx.Err() {
				return
			}
			node.tChildren[label] = newNode()
			added++
		}

		// Descend into the child node
		node = node.tChildren[label]
		if isWild = ("*" == label); isWild {
			node.terminator = wildMask
			ends++
		}
		if (len(aPartsList) - 1) == depth {
			if isEnd = (!isWild); isEnd {
				node.terminator |= endMask
				ends++
			}
		}
	} // for parts
	rOK = (0 < added) || (0 < ends)

	return
} // add()

// `allPatterns()` collects all hostname patterns in the node's tree.
//
// The patterns are returned in sorted order.
//
// The method uses a stack to traverse the tree in a depth-first manner,
// which is more efficient than a recursive approach.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//
// Returns:
//   - `rList`: A list of all patterns in the node's tree.
func (n *tNode) allPatterns(aCtx context.Context) (rList tPartsList) {
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
		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Pop the top of the stack
		current := stack[len(stack)-1]
		// Remove the top of the stack
		stack = stack[:len(stack)-1]

		// Check if current node is a terminal pattern
		if ((current.node.terminator & endMask) == endMask) ||
			((current.node.terminator & wildMask) == wildMask) {
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

		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
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
		tChildren:  make(tChildren),
		terminator: n.terminator,
	}

	// Clone the children nodes
	for label, child := range n.tChildren {
		clone.tChildren[label] = child.clone()
	}

	return clone
} // clone()

// `count()` returns the number of nodes and patterns in the node's tree.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//
// Returns:
//   - `rNodes`: The number of nodes in the node's tree.
//   - `rPatterns`: The number of patterns in the node's tree.
func (n *tNode) count(aCtx context.Context) (rNodes, rPatterns int) {
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
		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Pop the top of the stack
		current := stack[len(stack)-1]
		// Remove the top of the stack
		stack = stack[:len(stack)-1]
		rNodes++

		// Check if current node is a terminal pattern
		if ((current.node.terminator & endMask) == endMask) ||
			((current.node.terminator & wildMask) == wildMask) {
			rPatterns++
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
		for idx := len(kids) - 1; 0 <= idx; idx-- {
			label := kids[idx]
			newPath := make(tPartsList, len(current.path)+1)
			copy(newPath, current.path)
			newPath[len(current.path)] = label
			stack = append(stack, tStackEntry{
				node: current.node.tChildren[label],
				path: newPath,
			})
		}
	} // for stack

	return
} // count()

// `delete()` removes path patterns from the node's tree.
//
// The method returns `true` if at least one node is deleted, `false`
// otherwise.
// A wildcard pattern has no special meaning here but is just another
// label to match.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aPartsList`: The list of parts of the pattern to delete.
//
// Returns:
//   - `rOK`: `true` if the node is deleted, `false` otherwise.
func (n *tNode) delete(aCtx context.Context, aPartsList tPartsList) (rOK bool) {
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
		isEnd, isWild bool
		stack         []tStackEntry
	)
	current := n

	// Traverse and build up the stack
	for _, part := range aPartsList {
		child, ok := current.tChildren[part]
		if !ok {
			// Pattern does not exist; nothing to delete
			return
		}
		stack = append(stack, tStackEntry{node: current, label: part})
		current = child
	}

	// Unset terminal markers at the end node
	current.terminator = 0

	// Backtrack and prune
	for idx := len(stack) - 1; 0 <= idx; idx-- {
		parent := stack[idx].node
		label := stack[idx].label
		child := parent.tChildren[label]

		// If child has children stop pruning
		if 0 < len(child.tChildren) {
			break
		}

		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Safe to delete this child
		putNode(child) // Return the child to the pool
		delete(parent.tChildren, label)
		if 0 == len(parent.tChildren) {
			if isWild = ("*" == label); isWild {
				parent.terminator = wildMask
			}
			if isEnd = (!isWild); isEnd {
				parent.terminator |= endMask
			}
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

	if n.terminator != aNode.terminator {
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
// this package would be to gather statistics or calling the node's public
// `String()` method.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aFunc`: The function to call for each node.
func (n *tNode) forEach(aCtx context.Context, aFunc func(aNode *tNode)) {
	if (nil == n) || (nil == aFunc) {
		return
	}
	type tStackEntry struct {
		node *tNode
	}
	stack := []tStackEntry{{node: n}}

	for 0 < len(stack) {
		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

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

		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Push children to stack in reverse-sorted order
		// (to process them in forward order when popped)
		for idx := len(kids) - 1; 0 <= idx; idx-- {
			stack = append(stack, tStackEntry{
				node: entry.node.tChildren[kids[idx]],
			})
		}
	}
} // forEach()

// `load()` reads patterns from the reader and adds them to the node's tree.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aReader`: The reader to read the patterns from.
//
// Returns:
//   - `error`: `nil` if the patterns were read successfully, the error otherwise.
func (n *tNode) load(aCtx context.Context, aReader io.Reader) error {
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
		if ok := n.add(aCtx, parts); !ok {
			if err := aCtx.Err(); nil != err {
				return fmt.Errorf("failed to add pattern %q: %w", line, err)
			}

			return fmt.Errorf("failed to add pattern %q", line)
		}

		// Check for timeout or cancellation
		if err := aCtx.Err(); nil != err {
			return err
		}
	}

	return scanner.Err()
} // load()

// `match()` checks whether the node's tree contains the given pattern.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aPartsList`: The list of parts of the pattern to check.
//
// Returns:
//   - `rOK`: `true` if the pattern is in the node's trie, `false` otherwise.
func (n *tNode) match(aCtx context.Context, aPartsList tPartsList) (rOK bool) {
	if (nil == n) || (0 == len(aPartsList)) {
		return
	}
	if 0 == len(n.tChildren) {
		// No children, thus no match
		return
	}

	var ( // avoid repeated allocations inside the loop
		child *tNode
		depth int
		isEnd bool
		label string
		ok    bool
	)

	for depth, label = range aPartsList {
		if nil != aCtx.Err() {
			return
		}
		child, ok = n.tChildren[label]
		if !ok {
			// No child with that name, check for wildcard
			if child, ok = n.tChildren["*"]; ok {
				rOK = (((child.terminator & wildMask) == wildMask) ||
					((child.terminator & endMask) == endMask))
			}
			break
		}

		n = child
		// First check for a wildcard match at the current level,
		if child, ok = n.tChildren["*"]; ok {
			rOK = (((child.terminator & wildMask) == wildMask) ||
				((child.terminator & endMask) == endMask))
			break
		} else {
			// Remember the last non-wildcard node
			child = n
		}
		isEnd = ((n.terminator & endMask) == endMask)
		if ((n.terminator & wildMask) == wildMask) ||
			(len(aPartsList)-1 == depth) && isEnd {
			rOK = true
			break
		}
	}

	return
} // match()

// `store()` writes all patterns currently in the node to the writer,
// one hostname pattern per line.
//
// If `aIP` is not an empty string, it is prepended to each pattern line.
// This is useful for writing `hosts(5)` format files instead of a simple
// list of hostnames.
//
// If `aWriter` returns an error during processing, the method stops
// writing and returns that error to the caller.
//
// The method uses an internal stack to traverse the tree in a depth-first
// manner, which is more efficient than a recursive approach. The patterns
// are written in sorted order.
//
// The method is not thread-safe in itself but expects to be RLocked
// by the calling `tTrie` instance.
//
// If `aWriter` returns an error during processing, the method stops
// writing and returns that error to the caller.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aWriter`: The writer to write the patterns to.
//   - `aIP`: An IP address to prepend to each pattern line.
//
// Returns:
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
func (n *tNode) store(aCtx context.Context, aWriter io.Writer, aIP string) error {
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

		// Check for timeout or cancellation
		if err := aCtx.Err(); nil != err {
			return err
		}

		// Process current node
		if ((entry.node.terminator & endMask) == endMask) ||
			((entry.node.terminator & wildMask) == wildMask) {
			// Reverse path to original FQDN format
			pLen := len(entry.path)
			reversed := make(tPartsList, pLen)
			for idx, part := range entry.path {
				reversed[pLen-1-idx] = part
			}
			fqdn := strings.Join(reversed, ".")
			if 0 < len(aIP) {
				fqdn = aIP + " " + fqdn
			}

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

		// Check for timeout or cancellation
		if err := aCtx.Err(); nil != err {
			return err
		}

		// Push children in reverse-sorted order for
		// correct processing sequence
		for idx := len(kids) - 1; 0 <= idx; idx-- {
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
			line := fmt.Sprintf("%q:\n%sisEnd: %v\n%sisWild: %v\n",
				entry.name,
				indent, ((entry.node.terminator & endMask) == endMask),
				indent, ((entry.node.terminator & wildMask) == wildMask))
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
// The method adds the new pattern and tries to delete the old one. If the
// new pattern couldn't be added, the old one is not deleted. But as long
// as the new pattern could be added the method returns `true`.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aOldParts`: The list of parts of the old pattern to update.
//   - `aNewParts`: The list of parts of the new pattern to update to.
//
// Returns:
//   - `bool`: `true` if the pattern was updated, `false` otherwise.
func (n *tNode) update(aCtx context.Context, aOldParts, aNewParts tPartsList) bool { //TODO: use Context
	if (nil == n) || (0 == len(aOldParts)) || (0 == len(aNewParts)) || aOldParts.Equal(aNewParts) {
		return false
	}
	var added bool
	// Locking is done by the calling `tTrie`

	// 1. Add the new pattern
	if added = n.add(aCtx, aNewParts); added {
		// 2. Only after successful addition, delete the old pattern.
		// This might fail if the old pattern is part of a longer
		// pattern, but that's not a problem as the new pattern is
		// already in place.
		_ = n.delete(aCtx, aOldParts)
	}

	return added
} // update()

/* _EoF_ */
