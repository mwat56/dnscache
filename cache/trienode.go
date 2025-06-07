/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `tCachedIP` is a DNS cache entry.
	tCachedIP struct {
		TIpList              // IP addresses for this entry
		bestBefore time.Time // time after which the entry is invalid
	}

	// `tChildren` is a map of children nodes.
	tChildren map[string]*tCacheNode

	// `tCacheNode` represents a node in the trie.
	//
	// The node is considered a leaf node if no IPs are assigned,
	// otherwise it's an end node finishing a hostname pattern and
	// storing the IP addresses for the hostname pattern.
	tCacheNode struct {
		tCachedIP // cached data for this node
		tChildren // children nodes
	}
)

var (
	// `ErrNodeNil` is returned if a node or a method's required
	// argument is `nil`.
	ErrNodeNil = errors.New("node or argument is nil")
)

// ---------------------------------------------------------------------------
// `tCacheNode` constructor:

// `newNode()` creates a new `tCacheNode` instance.
//
// Returns:
//   - `*tCacheNode`: A new `tCacheNode` instance.
func newNode() *tCacheNode {
	// TODO: Use a pool for `tCacheNode` instances.

	return &tCacheNode{tChildren: make(tChildren)}
} // newNode()

// ---------------------------------------------------------------------------
// `tCacheNode` methods:

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
//   - `aIPs`: The list of IP addresses to store with the cache entry.
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `bool`: `true` if a pattern was added, `false` otherwise.
func (cn *tCacheNode) add(aCtx context.Context, aPartsList tPartsList, aIPs TIpList, aTTL time.Duration) (rOK bool) {
	if (nil == cn) || (0 == len(aPartsList)) {
		return
	}
	if 0 == aTTL {
		aTTL = DefaultTTL
	}

	node, ok := cn, false
	for depth, label := range aPartsList {
		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Create a new child node if it doesn't exist
		if nil == node.tChildren {
			node.tChildren = make(tChildren)
			node.tChildren[label] = newNode()
		} else if _, ok = node.tChildren[label]; !ok {
			node.tChildren[label] = newNode()
		}

		// Descend into the child node
		if node, ok = node.tChildren[label]; ok {
			if (len(aPartsList) - 1) == depth {
				node.update(aIPs, aTTL)
				rOK = true
			}
		}
	}

	return
} // add()

// `allPatterns()` collects all hostname patterns in the node's tree.
//
// The method's result is returned in sorted order of the original
// hostname patterns.
//
// The method uses a stack to traverse the tree in a depth-first manner,
// which is more efficient than a recursive approach.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//
// Returns:
//   - `rList`: A list of all patterns in the node's tree.
func (cn *tCacheNode) allPatterns(aCtx context.Context) (rList tPatternList) {
	if nil == cn {
		return
	}

	type tStackEntry struct {
		node  *tCacheNode // respective node to process
		parts tPartsList  // path in the trie to the node
	}
	var (
		cLen, idx, pLen int
		kidNames        tPartsList
		label           string
	)
	stack := []tStackEntry{
		// Push the current node to the stack
		{node: cn, parts: tPartsList{}},
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

		// Check if current node finishes a pattern
		if 0 < len(current.node.tCachedIP.TIpList) {
			// Reverse the path to get the original FQDN
			// in original order.
			if pLen = len(current.parts); 0 < pLen {
				reversed := make(tPartsList, pLen)
				for idx, label = range current.parts {
					reversed[pLen-1-idx] = label
				}
				rList = append(rList, strings.Join(reversed, "."))
			}
		}

		if cLen = len(current.node.tChildren); 0 == cLen {
			continue
		}

		// Collect and sort children keys for deterministic order
		kidNames = make(tPartsList, 0, cLen)
		for label = range current.node.tChildren {
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
		for idx = len(kidNames) - 1; 0 <= idx; idx-- {
			label := kidNames[idx]
			child := current.node.tChildren[label]
			newParts := make(tPartsList, len(current.parts)+1)
			copy(newParts, current.parts)
			newParts[len(current.parts)] = label
			stack = append(stack, tStackEntry{
				node:  child,
				parts: newParts,
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
func (cn *tCacheNode) count(aCtx context.Context) (rNodes, rPatterns int) {
	if nil == cn {
		return
	}

	type (
		tStackEntry struct {
			node  *tCacheNode // respective node to process
			parts tPartsList  // path in the trie to the node
		}
	)
	var (
		cLen, ipLen int
		current     tStackEntry
		label       string
		kidNames    tPartsList
	)
	stack := []tStackEntry{
		// Push the current node to the stack
		{node: cn, parts: tPartsList{}},
	}

	for 0 < len(stack) {
		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Pop the top of the stack
		current = stack[len(stack)-1]
		// Remove the top of the stack
		stack = stack[:len(stack)-1]

		cLen = len(current.node.tChildren)
		ipLen = len(current.node.tCachedIP.TIpList)
		if (0 < cLen) || (0 < ipLen) {
			// valid node with either children or IPs
			rNodes++
		}
		if 0 < ipLen {
			// with IPs it's a complete pattern
			rPatterns++
		}
		if 0 == cLen {
			if 0 < rNodes {
				// Un-count the node without children
				rNodes--
			}
			continue
		}

		// Collect and sort children keys for deterministic order
		kidNames = make(tPartsList, 0, cLen)
		for label = range current.node.tChildren {
			kidNames = append(kidNames, label)
		}
		if 1 < len(kidNames) {
			sort.Strings(kidNames)
		}

		// Push children to stack in reverse-sorted order
		// (to process them in forward order when popped)
		for idx := len(kidNames) - 1; 0 <= idx; idx-- {
			label = kidNames[idx]
			newParts := make(tPartsList, len(current.parts)+1)
			copy(newParts, current.parts)
			newParts[len(current.parts)] = label
			stack = append(stack, tStackEntry{
				node:  current.node.tChildren[label],
				parts: newParts,
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
func (cn *tCacheNode) delete(aCtx context.Context, aPartsList tPartsList) (rOK bool) {
	if (nil == cn) || (0 == len(aPartsList)) {
		return
	}
	type (
		tStackEntry struct {
			label string
			node  *tCacheNode
		}
	)
	var (
		child, current, parent *tCacheNode
		label                  string
		ok                     bool
		stack                  []tStackEntry
	)

	current = cn
	// Traverse and build up the stack
	for _, label := range aPartsList {
		if child, ok = current.tChildren[label]; !ok {
			// Pattern does not exist: nothing to delete
			return
		}
		stack = append(stack, tStackEntry{label, current})
		current = child
	}

	// Backtrack and prune
	for idx := len(stack) - 1; 0 <= idx; idx-- {
		label = stack[idx].label
		parent = stack[idx].node
		child = parent.tChildren[label]

		if 0 < len(child.tChildren) {
			// If node has children stop pruning but
			// clear cache data
			child.tCachedIP = tCachedIP{}
			return
		}

		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Safe to delete this child
		//TODO: Return the child to the pool
		// putNode(parent.tChildren[label])
		delete(parent.tChildren, label)
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
func (cn *tCacheNode) Equal(aNode *tCacheNode) (rOK bool) {
	if nil == cn {
		return (nil == aNode)
	}
	if nil == aNode {
		return
	}
	if cn == aNode {
		rOK = true
		return
	}

	// We're only interested in the node structure so we ignore
	// the cached IPs and expiration times while comparing.

	if len(cn.tChildren) != len(aNode.tChildren) {
		return
	}

	for label, child := range cn.tChildren {
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

// `expire()` removes expired cache entries from the node's tree.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//
// Returns:
//   - `rOK`: `true` if at least one cache entry was removed, `false` otherwise.
func (cn *tCacheNode) expire(aCtx context.Context) (rOK bool) {
	if nil == cn {
		return
	}

	type tStackEntry struct {
		label  string
		node   *tCacheNode
		parent *tCacheNode
	}

	// Start with root node (no parent)
	stack := []tStackEntry{{node: cn}}
	nodes2Delete := []tStackEntry{}

	// First pass: identify expired nodes and mark for deletion
	for 0 < len(stack) {
		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Pop from stack
		entry := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Check if this node is expired
		if (0 < len(entry.node.tCachedIP.TIpList)) &&
			entry.node.tCachedIP.bestBefore.Before(time.Now()) {
			// Clear the expired data first in any case
			entry.node.tCachedIP = tCachedIP{}

			// Mark for deletion if it has no children
			if 0 == len(entry.node.tChildren) && entry.parent != nil {
				nodes2Delete = append(nodes2Delete, entry)
			}

			rOK = true
		}

		// Add children to stack
		for label, child := range entry.node.tChildren {
			stack = append(stack, tStackEntry{
				label:  label,
				node:   child,
				parent: entry.node,
			})
		}
	}

	// Second pass: delete marked nodes
	for _, entry := range nodes2Delete {
		if nil != aCtx.Err() {
			return
		}

		//TODO: Return the child to the pool
		// putNode(entry.parent.tChildren[entry.label])

		// Delete the node from its parent
		delete(entry.parent.tChildren, entry.label)
	}

	return
} // expire()

/* * /
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
func (cn *tCacheNode) forEach(aCtx context.Context, aFunc func(aNode *tCacheNode)) {
	if (nil == cn) || (nil == aFunc) {
		return
	}
	type tStackEntry struct {
		node *tCacheNode
	}
	stack := []tStackEntry{{node: cn}}

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
/* */

// `ips()` returns the IP addresses for the given pattern.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aPartsList`: The list of parts of the pattern to check.
//
// Returns:
//   - `rIPs`: The list of IP addresses for the given pattern.
func (cn *tCacheNode) ips(aCtx context.Context, aPartsList tPartsList) (rIPs TIpList) {
	if (nil == cn) || (0 == len(aPartsList)) || (0 == len(cn.tChildren)) {
		return
	}

	var ( // avoid repeated allocations inside the loop
		child *tCacheNode
		ok    bool
	)

	node := cn
	for depth, label := range aPartsList {
		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Check for a child with the next label
		if child, ok = node.tChildren[label]; !ok {
			return
		}

		// Descend into the child node
		node = child
		if len(aPartsList)-1 == depth {
			// We're at the last label of the pattern,
			// check for a terminal match
			if node.isExpired() {
				return
			}
			// if 0 < len(node.tCachedIP.TIpList) {
			rIPs = node.tCachedIP.TIpList
			// }
		}
	}

	return
} // ips()

// `isExpired()` returns `true` if the cache entry is expired.
//
// Returns:
//   - `bool`: `true` if the cache entry is expired, `false` otherwise.
func (cn *tCacheNode) isExpired() bool {
	if nil == cn { //  || (0 == len(cn.tCachedIP.TIpList))
		return true
	}

	return cn.tCachedIP.bestBefore.Before(time.Now())
} // isExpired()

// `match()` checks whether the node's tree contains the given pattern.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aPartsList`: The list of parts of the pattern to check.
//
// Returns:
//   - `rOK`: `true` if the pattern is in the node's trie, `false` otherwise.
func (cn *tCacheNode) match(aCtx context.Context, aPartsList tPartsList) (rOK bool) {
	if (nil == cn) || (0 == len(aPartsList)) {
		return
	}
	if 0 == len(cn.tChildren) {
		// No children, thus no match
		return
	}

	var ( // avoid repeated allocations inside the loop
		child *tCacheNode
		ok    bool
	)

	for depth, label := range aPartsList {
		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Check for a child with the next label
		if child, ok = cn.tChildren[label]; !ok {
			return
		}

		// Descend into the child node
		cn = child
		if len(aPartsList)-1 == depth {
			// We're at the last label of the pattern,
			// check for a terminal match
			rOK = (0 < len(cn.tCachedIP.TIpList))
		}
	}

	return
} // match()

/*
// `merge()` merges the subtree of `aSrc` into the current node.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aSrc`: The source node to merge from.
//
// Returns:
//   - `*tCacheNode`: The merged node.
func (n *tCacheNode) merge(aCtx context.Context, aSrc *tCacheNode) *tCacheNode {
	if nil == n {
		return aSrc
	}
	if nil == aSrc {
		return n
	}
	type (
		tStackEntry struct {
			srcNode  *tCacheNode
			destNode *tCacheNode
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
*/

// `store()` writes all patterns currently in the node to the writer, one
// hostname pattern per line.
//
// If `aWriter` returns an error during processing, the method stops
// writing and returns that error to the caller.
//
// The method uses an internal stack to traverse the tree in a depth-first
// manner, which is more efficient than a recursive approach. The patterns
// are written in sorted order.
//
// The method is not thread-safe in itself but expects to be RLocked
// by the calling trie instance.
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
func (cn *tCacheNode) store(aCtx context.Context, aWriter io.Writer) error {
	if (nil == cn) || (nil == aWriter) {
		return ErrNodeNil
	}
	type (
		tStackEntry struct {
			parts tPartsList
			node  *tCacheNode
		}
	)
	var (
		cLen, pLen int
		err        error
		fqdn       string
	)

	stack := []tStackEntry{{tPartsList{}, cn}}
	for 0 < len(stack) {
		// Pop from stack
		entry := stack[len(stack)-1]
		// Remove current entry from stack
		stack = stack[:len(stack)-1]

		// Check for timeout or cancellation
		if err = aCtx.Err(); nil != err {
			return err
		}

		if 0 < len(entry.node.tCachedIP.TIpList) { // valid end node
			// Reverse path to original FQDN format
			pLen = len(entry.parts)
			reversed := make(tPartsList, pLen)
			for idx, part := range entry.parts {
				reversed[pLen-1-idx] = part
			}
			fqdn = strings.Join(reversed, ".")

			// Write hosts(5) style with IP addresses and FQDN
			for _, ip := range entry.node.tCachedIP.TIpList {
				if _, err = fmt.Fprintf(aWriter, "%s %s\n", ip.String(), fqdn); nil != err {
					return err
				}
			}
		}

		if cLen = len(entry.node.tChildren); 0 == cLen {
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
		if err = aCtx.Err(); nil != err {
			return err
		}

		// Push children in reverse-sorted order for
		// correct processing sequence
		for idx := len(kidNames) - 1; 0 <= idx; idx-- {
			label := kidNames[idx]
			newPath := make(tPartsList, len(entry.parts)+1)
			copy(newPath, entry.parts)
			newPath[len(entry.parts)] = label

			stack = append(stack, tStackEntry{
				node:  entry.node.tChildren[label],
				parts: newPath,
			})
		}
	}

	return nil
} // store()

/* * /
// `string()` returns a string representation of the node.
//
// Parameters:
//   - `aLabel`: The label to use for the current node.
//
// Returns:
//   - `string`: The string representation of the node.
func (cn *tCacheNode) string(aLabel string) string {
	if nil == cn {
		return ""
	}

	type (
		tStackEntry struct {
			kidNames tPartsList  // sorted child keys
			name     string      // label of the node
			node     *tCacheNode // respective node to process
			depth    int         // depth in the tree
			childIdx int         // index of next child to process
		}
	)
	stack := []tStackEntry{
		{
			kidNames: nil,
			name:     aLabel,
			node:     cn,
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
			line := fmt.Sprintf("%s\n",
				entry.name)
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
/* */

// `String()` implements the `fmt.Stringer` interface for the node.
//
// Returns:
//   - `string`: The string representation of the node.
func (cn *tCacheNode) String() string {
	if nil == cn {
		return ""
	}

	// Buffer for the string representation's parts
	var builder strings.Builder
	cn.store(context.Background(), &builder)

	return builder.String()
	// return cn.string("Node")
} // String()

/*
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
func (n *tCacheNode) updateX(aCtx context.Context, aOldParts, aNewParts tPartsList) bool {
	if (nil == n) || (0 == len(aOldParts)) || (0 == len(aNewParts)) {
		return false
	}
	var added bool
	// Locking is done by the calling trie

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
*/

// `update()` updates the cache entry with the given IP addresses returning
// the updated cache node.
//
// If the given IP list is empty, the cache node's IP list is cleared/removed.
//
// Parameters:
//   - `aIPs`: List of IP addresses to update the cache entry with.
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `*tCacheNode`: The updated cache node.
func (cn *tCacheNode) update(aIPs TIpList, aTTL time.Duration) *tCacheNode {
	if nil == cn {
		return nil
	}
	if 0 == aTTL {
		aTTL = DefaultTTL
	}

	// Update IPs
	if iLen := len(aIPs); 0 < iLen {
		// Assume ownership of `aIPs`
		cn.tCachedIP.TIpList = make(TIpList, iLen)
		copy(cn.tCachedIP.TIpList, aIPs)

		// Update expiration time
		cn.tCachedIP.bestBefore = time.Now().Add(aTTL)
	} else {
		// Clear cache data
		cn.tCachedIP = tCachedIP{}
	}

	return cn
} // update()

/* _EoF_ */
