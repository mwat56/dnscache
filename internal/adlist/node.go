/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
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
		tChildren        // children nodes
		terminator uint8 // flags for pattern end and wildcard
	}
)

var (
	// `ErrNodeNil` is returned if a node or a method's required
	// argument is `nil`.
	ErrNodeNil = ADlistError{errors.New("node or argument is nil")}
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
		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}
		if _, ok = node.tChildren[label]; !ok {
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
		cLen     int
		kidNames tPartsList
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
		if 0 != current.node.terminator { // either end or wildcard bits
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
		kidNames = make(tPartsList, 0, cLen)
		for label := range current.node.tChildren {
			kidNames = append(kidNames, label)
		}
		if 1 < len(kidNames) {
			sort.Strings(kidNames)
		}

		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Push children to stack in reverse-sorted order
		// (to process them in forward order when popped)
		for i := len(kidNames) - 1; 0 <= i; i-- {
			label := kidNames[i]
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

	var (
		child, node *tNode
		dec         int
	)
	stack := make([]*tNode, 0, 1024) // Pre-allocated buffer
	// Push the current node to the stack
	stack = append(stack, n)

	for 0 < len(stack) {
		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}
		// Pop the top of the stack
		node = stack[0]
		// Remove the top of the stack
		stack = stack[1:]

		rNodes++
		if 0 != node.terminator {
			// With either end or wildcard bits it's a complete pattern
			rPatterns++
		}
		if 0 == len(node.tChildren) {
			if (0 < rNodes) && (0 == dec) {
				// Un-count the node without children
				rNodes--
				dec++
			}
			continue
		}

		for _, child = range node.tChildren {
			stack = append(stack, child)
		}
	}

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
// NOTE: This method is of no practical use apart from unit-testing.
//
// Parameters:
//   - `aNode`: The node to compare with.
//
// Returns:
//   - `rOK`: `true` if the node is equal to the given one, `false` otherwise.
func (n *tNode) Equal(aNode *tNode) (rOK bool) {
	if nil == n {
		return (nil == aNode)
	}
	if nil == aNode {
		return
	}
	if n == aNode {
		rOK = true
		return
	}

	if n.terminator != aNode.terminator {
		return
	}
	if len(n.tChildren) != len(aNode.tChildren) {
		return
	}

	for label, child := range n.tChildren {
		otherChild, ok := aNode.tChildren[label]
		if !ok {
			return
		}
		if !child.Equal(otherChild) {
			return
		}
	}
	rOK = true

	return
} // Equal()

// `finalNode()` returns the node that matches the final part of ´aPartsList`.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aPartsList`: The list of parts of the pattern to check.
//
// Returns:
//   - `rNode`: The node that matches the pattern, `nil` otherwise.
//   - `rOK`: `true` if the pattern is in the node's Trie, `false` otherwise.
func (n *tNode) finalNode(aCtx context.Context, aPartsList tPartsList) (rNode *tNode, rOK bool) {
	if (nil == n) || (0 == len(aPartsList)) {
		return
	}

	var ( // avoid repeated allocations inside the loop
		child *tNode
		depth int
		label string
		ok    bool
	)

	current := n
	for depth, label = range aPartsList {
		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Check for a child with the next label
		if child, ok = current.tChildren[label]; !ok {
			// No child with that name, so check for a wildcard
			// match at the current level
			if child, ok = current.tChildren["*"]; ok {
				if rOK = (0 != child.terminator); rOK {
					rNode = child
				}
			}

			return
		}

		// Descend into the child node
		current = child
		if depth < len(aPartsList)-1 {
			if child, ok = current.tChildren["*"]; !ok {
				continue
			}

			// We're at an intermediate node with a
			// wildcard child, so we have a match.
			if rOK = ((child.terminator & wildMask) == wildMask); !rOK {
				return
			}
			rNode = child

			// Check whether there's also a literal match:
			if child, ok = current.tChildren[aPartsList[depth+1]]; ok {
				// Don't change `rOK` because we already have
				// a valid (wildcard) match.
				if ok = ((child.terminator & endMask) == endMask); ok {
					rNode = child
				}
			}

			return
		} else if len(aPartsList)-1 == depth {
			// We're at the last label of the pattern
			// hence check for a terminal match:
			if rOK = (0 != current.terminator); rOK {
				rNode = current
			}

			return
		}
	}

	return
} // finalNode()

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

		// Collect and sort children kidNames for deterministic order
		cLen := len(entry.node.tChildren)
		if 0 == cLen {
			continue
		}

		kidNames := make(tPartsList, 0, cLen)
		for label := range entry.node.tChildren {
			kidNames = append(kidNames, label)
		}
		if 1 < len(kidNames) {
			sort.Strings(kidNames)
		}

		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Push children to stack in reverse-sorted order
		// (to process them in forward order when popped)
		for idx := len(kidNames) - 1; 0 <= idx; idx-- {
			stack = append(stack, tStackEntry{
				node: entry.node.tChildren[kidNames[idx]],
			})
		}
	}
} // forEach()

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

	_, rOK = n.finalNode(aCtx, aPartsList)
	return
} // match()

// `merge()` merges the subtree of `aSrc` into the current node.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aSrc`: The source node to merge from.
//
// Returns:
//   - `*tNode`: The merged node.
func (n *tNode) merge(aCtx context.Context, aSrc *tNode) *tNode {
	if nil == n {
		return aSrc
	}
	if nil == aSrc {
		return n
	}
	type (
		tStackEntry struct {
			srcNode  *tNode
			destNode *tNode
		}
	)
	var label string

	stack := []tStackEntry{{aSrc, n}}

	for 0 < len(stack) {
		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			break
		}
		entry := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Merge terminal flags using OR
		if 0 != entry.srcNode.terminator {
			entry.destNode.terminator |= entry.srcNode.terminator
		}

		// Collect and sort children keys for deterministic order
		cLen := len(entry.srcNode.tChildren)
		if 0 == cLen {
			continue
		}

		kidNames := make(tPartsList, 0, cLen)
		for label = range entry.srcNode.tChildren {
			kidNames = append(kidNames, label)
		}
		if 1 < len(kidNames) {
			sort.Strings(kidNames)
		}

		for _, label = range kidNames {
			srcChild := entry.srcNode.tChildren[label]
			destChild, exists := entry.destNode.tChildren[label]

			if !exists {
				// Create new destination child
				destChild = newNode()
				destChild.terminator = srcChild.terminator
				entry.destNode.tChildren[label] = destChild
			}

			// Push to stack for deeper merge
			stack = append(stack, tStackEntry{srcChild, destChild})
		}
	}

	return n
} // merge()

// `store()` writes all patterns currently in the node to the writer,
// one hostname pattern per line.
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
//
// Returns:
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
func (n *tNode) store(aCtx context.Context, aWriter io.Writer) error {
	if (nil == n) || (nil == aWriter) {
		return ErrNodeNil
	}
	type (
		tStackEntry struct {
			path tPartsList
			node *tNode
		}
	)

	stack := []tStackEntry{
		{
			path: tPartsList{},
			node: n,
		},
	}
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
		if 0 != entry.node.terminator { // either end or wildcard bits
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

		// Collect and sort children kidNames
		kidNames := make(tPartsList, 0, cLen)
		for label := range entry.node.tChildren {
			kidNames = append(kidNames, label)
		}
		if 1 < len(kidNames) {
			sort.Strings(kidNames)
		}

		// Check for timeout or cancellation
		if err := aCtx.Err(); nil != err {
			return err
		}

		// Push children in reverse-sorted order for
		// correct processing sequence
		for idx := len(kidNames) - 1; 0 <= idx; idx-- {
			label := kidNames[idx]
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
			kidNames tPartsList // sorted child keys
			name     string     // label of the node
			node     *tNode     // respective node to process
			depth    int        // depth in the tree
			childIdx int        // index of next child to process
		}
	)
	// Locking is done by the calling `tTrie`
	stack := []tStackEntry{
		{
			kidNames: nil,
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
		if nil == entry.kidNames {
			line := fmt.Sprintf("%q:\n%sisEnd: %v\n%sisWild: %v\n",
				entry.name,
				indent, ((entry.node.terminator & endMask) == endMask),
				indent, ((entry.node.terminator & wildMask) == wildMask))
			builder.WriteString(line)

			// Prepare sorted child keys
			entry.kidNames = make(tPartsList, 0, len(entry.node.tChildren))
			for label := range entry.node.tChildren {
				entry.kidNames = append(entry.kidNames, label)
			}
			if 1 < len(entry.kidNames) {
				sort.Strings(entry.kidNames)
			}
		}

		// If there are unprocessed children, process the next one
		if entry.childIdx < len(entry.kidNames) {
			// Indent for the child node
			builder.WriteString(indent)

			label := entry.kidNames[entry.childIdx]
			child := entry.node.tChildren[label]
			entry.childIdx++

			// Push the child node to the stack
			stack = append(stack, tStackEntry{
				kidNames: nil,
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
