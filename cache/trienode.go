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
		tIpList              // IP addresses for this entry
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

// --------------------------------------------------------------------------

// `init()` ensures proper interface implementation.
func init() {
	var (
		_ iCacheNode = (*tCacheNode)(nil)
	)
} // init()

// ---------------------------------------------------------------------------
// `tCacheNode` methods:

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
		parts tPartsList  // path in the trie to the node
		node  *tCacheNode // respective node to process
	}
	var (
		cLen, idx, pLen int
		kidNames        tPartsList
		label           string
	)
	stack := []tStackEntry{
		// Push the current node to the stack
		{parts: tPartsList{}, node: cn},
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
		if 0 < len(current.node.tCachedIP.tIpList) {
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
				parts: newParts,
				node:  child,
			})
		}
	} // for stack

	return
} // allPatterns()

// `count()` returns the number of nodes and patterns in the node's trie.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//
// Returns:
//   - `rNodes`: The number of nodes in the node's trie.
//   - `rPatterns`: The number of patterns in the node's trie.
func (cn *tCacheNode) count(aCtx context.Context) (rNodes, rPatterns int) {
	if nil == cn {
		return
	}

	type (
		tStackEntry struct {
			parts tPartsList  // path in the trie to the node
			node  *tCacheNode // respective node to process
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
		{parts: tPartsList{}, node: cn},
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
		ipLen = len(current.node.tCachedIP.tIpList)
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
				parts: newParts,
				node:  current.node.tChildren[label],
			})
		}
	} // for stack

	return
} // count()

// `Create()` inserts a pattern to the node's trie.
//
// The method returns `true` if at least one part was added in order to
// have the whole `aPartsList` present in the trie or found an existing
// node that already represented the pattern.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aPartsList`: The list of parts of the pattern to create.
//   - `aIPs`: The list of IP addresses to store with the cache entry.
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `bool`: `true` if a pattern was added, `false` otherwise.
func (cn *tCacheNode) Create(aCtx context.Context, aPartsList tPartsList, aIPs tIpList, aTTL time.Duration) (rOK bool) {
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
				node.Update(aCtx, aIPs, aTTL)
				rOK = true
			}
		}
	}

	return
} // Create()

// `Delete()` removes path patterns from the node's trie.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aPartsList`: The list of parts of the pattern to Delete.
//
// Returns:
//   - `rOK`: `true` if a node was deleted, `false` otherwise.
func (cn *tCacheNode) Delete(aCtx context.Context, aPartsList tPartsList) (rOK bool) {
	if (nil == cn) || (0 == len(aPartsList)) {
		return
	}
	type (
		tStackEntry struct {
			name string
			node *tCacheNode
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

	// The target node (the one specified by `aPartsList`).
	// If it has children, just clear its IPs and return.
	if 0 < len(current.tChildren) {
		current.tCachedIP = tCachedIP{}
		return
	}

	// Target node has no children, safe to delete it.
	// Start backtracking and pruning.
	for idx := len(stack) - 1; 0 <= idx; idx-- {
		label = stack[idx].name
		parent = stack[idx].node

		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Safe to delete the child node
		//TODO: Return the child to the pool
		// putNode(parent.tChildren[label])
		delete(parent.tChildren, label)
		rOK = true

		// If parent has other children or has its own IPs, stop pruning
		if 0 < len(parent.tChildren) || 0 < len(parent.tCachedIP.tIpList) {
			return
		}
	}

	return
} // Delete()

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

// `expire()` removes expired cache entries from the node's trie.
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
		name   string
		node   *tCacheNode
		parent *tCacheNode
	}

	// Start with root node (no parent)
	stack := []tStackEntry{{node: cn}}
	nodes2Delete := []tStackEntry{}

	// First pass: identify expired nodes and mark for deletion
	for 0 < len(stack) {
		// Pop from stack
		entry := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Check for timeout or cancellation
		if nil != aCtx.Err() {
			return
		}

		// Check if this node is expired
		if (0 < len(entry.node.tCachedIP.tIpList)) &&
			entry.node.tCachedIP.bestBefore.Before(time.Now()) {
			// Clear the expired data first
			entry.node.tCachedIP = tCachedIP{}

			// Mark for deletion if it has no children and has a parent,
			// i.e. it's not the root node
			if 0 == len(entry.node.tChildren) && entry.parent != nil {
				nodes2Delete = append(nodes2Delete, entry)
			}
			rOK = true
		}

		// Add children to stack
		for label, child := range entry.node.tChildren {
			stack = append(stack, tStackEntry{
				name:   label,
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
		// putNode(entry.parent.tChildren[entry.name])

		// Delete the node from its parent
		delete(entry.parent.tChildren, entry.name)
	}

	return
} // expire()

// `finalNode()` returns the node that matches the final part of ´aPartsList`.
//
// Parameters:
//   - `aPartsList`: The list of parts of the pattern to check.
//
// Returns:
//   - `rNode`: The node that matches the pattern, `nil` otherwise.
//   - `rOK`: `true` if the pattern is in the node's trie, `false` otherwise.
func (cn *tCacheNode) finalNode(aPartsList tPartsList) (rNode *tCacheNode, rOK bool) {
	if nil == cn {
		return
	}

	var ( // avoid repeated allocations inside the loop
		child *tCacheNode
		depth int
		label string
		ok    bool
	)

	current := cn
	for depth, label = range aPartsList {
		// Check for a child with the next label
		if child, ok = current.tChildren[label]; !ok {
			return
		}

		// Descend into the child node
		current = child
		if len(aPartsList)-1 == depth {
			// We're at the last label of the pattern
			// hence check for a terminal match:
			if rOK = (0 < len(current.tCachedIP.tIpList)); rOK {
				if current.isExpired() {
					rOK = false
				} else {
					rNode = current
				}
			}
			return
		}
	}

	return
} // finalNode()

// `isExpired()` returns `true` if the cache entry is expired.
//
// Returns:
//   - `bool`: `true` if the cache entry is expired, `false` otherwise.
func (cn *tCacheNode) isExpired() bool {
	if nil == cn {
		return true
	}

	return cn.tCachedIP.bestBefore.Before(time.Now())
} // isExpired()

// `match()` checks whether the node's trie contains the given pattern and
// returns the node that matched the pattern.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aPartsList`: The list of parts of the pattern to check.
//
// Returns:
//   - `rNode`: The node that matched the pattern, `nil` otherwise.
//   - `rOK`: `true` if the pattern is in the node's trie, `false` otherwise.
func (cn *tCacheNode) match(aCtx context.Context, aPartsList tPartsList) (rNode *tCacheNode, rOK bool) {
	if (nil == cn) || (0 == len(aPartsList)) {
		return
	}
	if 0 == len(cn.tChildren) {
		// No children, thus no match
		return
	}
	rNode, rOK = cn.finalNode(aPartsList)

	return
} // Match()

// `Retrieve()` returns the IP addresses for the given pattern.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aPartsList`: The list of parts of the pattern to check.
//
// Returns:
//   - `rIPs`: The list of IP addresses for the given pattern.
func (cn *tCacheNode) Retrieve(aCtx context.Context, aPartsList tPartsList) (rIPs tIpList) {
	if (nil == cn) || (0 == len(aPartsList)) || (0 == len(cn.tChildren)) {
		return
	}

	if node, ok := cn.finalNode(aPartsList); ok {
		rIPs = node.tCachedIP.tIpList
	}

	return
} // Retrieve()

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

		if 0 < len(entry.node.tCachedIP.tIpList) { // valid end node
			// Reverse path to original FQDN format
			pLen = len(entry.parts)
			reversed := make(tPartsList, pLen)
			for idx, part := range entry.parts {
				reversed[pLen-1-idx] = part
			}
			fqdn = strings.Join(reversed, ".")

			// Write hosts(5) style with IP addresses and FQDN
			for _, ip := range entry.node.tCachedIP.tIpList {
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
				parts: newPath,
				node:  entry.node.tChildren[label],
			})
		}
	}

	return nil
} // store()

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
	_ = cn.store(context.Background(), &builder)

	return builder.String()
} // String()

// `Update()` updates the cache entry with the given IP addresses returning
// the updated cache node.
//
// If the given IP list is empty, the cache node's IP list is cleared/removed.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aIPs`: List of IP addresses to Update the cache entry with.
//   - `aTTL`: Time to live for the cache entry.
//
// Returns:
//   - `iCacheNode`: The updated cache node.
func (cn *tCacheNode) Update(aCtx context.Context, aIPs tIpList, aTTL time.Duration) iCacheNode {
	if nil == cn {
		return nil
	}
	if 0 == aTTL {
		aTTL = DefaultTTL
	}

	// Update IPs
	if iLen := len(aIPs); 0 < iLen {
		// Assume ownership of `aIPs`
		cn.tCachedIP.tIpList = make(tIpList, iLen)
		copy(cn.tCachedIP.tIpList, aIPs)

		// Update expiration time
		cn.tCachedIP.bestBefore = time.Now().Add(aTTL)
	} else {
		// Clear cache data
		cn.tCachedIP = tCachedIP{}
	}

	return cn
} // Update()

/* _EoF_ */
