/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"fmt"
	"strings"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `TMetrics` contains the metrics data for the node pool and the trie.
	//
	// These are the fields to access the metrics data:
	//
	//   - `PoolCreations`: Number of nodes created by the pool.
	//   - `PoolReturns`: Number of nodes returned to the pool.
	//   - `PoolSize`: Current number of items in the pool.
	//   - `Nodes`: Number of nodes in the trie.
	//   - `Patterns`: Number of patterns in the trie.
	//   - `Hits`: Number of times a pattern was found.
	//   - `Misses`: Number of times a pattern was not found.
	//   - `Reloads`: Number of times the list was reloaded.
	//   - `Retries`: Number of times a reload was retried.
	//   - `HeapAllocs`: Number of heap objects allocated.
	//   - `HeapFrees`: Number of heap objects freed.
	//   - `GCPauseTotalNs`: Cumulative nanoseconds in GC stop-the-world pauses.
	TMetrics struct {
		PoolCreations  uint32
		PoolReturns    uint32
		PoolSize       int
		Nodes          uint32
		Patterns       uint32
		Hits           uint32
		Misses         uint32
		Reloads        uint32
		Retries        uint32
		HeapAllocs     uint64
		HeapFrees      uint64
		GCPauseTotalNs uint64
	}
)

// ---------------------------------------------------------------------------
// `TMetrics` methods:

// `clone()` returns a copy of the current metrics data.
//
// Returns:
//   - `*TMetrics`: Current metrics data.
func (m *TMetrics) clone() *TMetrics {
	if nil == m {
		return nil
	}

	return &TMetrics{
		PoolCreations:  m.PoolCreations,
		PoolReturns:    m.PoolReturns,
		PoolSize:       m.PoolSize,
		Nodes:          m.Nodes,
		Patterns:       m.Patterns,
		Hits:           m.Hits,
		Misses:         m.Misses,
		Reloads:        m.Reloads,
		Retries:        m.Retries,
		HeapAllocs:     m.HeapAllocs,
		HeapFrees:      m.HeapFrees,
		GCPauseTotalNs: m.GCPauseTotalNs,
	}
} // clone()

// `Equal()` checks whether the metrics data is equal to the given one.
//
// Parameters:
//   - `aMetrics`: Metrics data to compare with.
//
// Returns:
//   - `bool`: `true` if the metrics data is equal to the given one, `false` otherwise.
func (m *TMetrics) Equal(aMetrics *TMetrics) bool {
	if nil == m {
		return (nil == aMetrics)
	}
	if nil == aMetrics {
		return false
	}

	return (m.PoolCreations == aMetrics.PoolCreations) &&
		(m.PoolReturns == aMetrics.PoolReturns) &&
		(m.PoolSize == aMetrics.PoolSize) &&
		(m.Nodes == aMetrics.Nodes) &&
		(m.Patterns == aMetrics.Patterns) &&
		(m.Hits == aMetrics.Hits) &&
		(m.Misses == aMetrics.Misses) &&
		(m.Reloads == aMetrics.Reloads) &&
		(m.Retries == aMetrics.Retries)
	//NOTE: Ignore the runtime stats because they vary with every run.
	// (m.HeapAllocs == aMetrics.HeapAllocs) &&
	// (m.HeapFrees == aMetrics.HeapFrees) &&
	// (m.GCPauseTotalNs == aMetrics.GCPauseTotalNs)
} // Equal()

// `String()` implements the `fmt.Stringer` interface for the metrics data.
//
// Returns:
//   - `string`: String representation of the metrics data.
func (m *TMetrics) String() string {
	if nil == m {
		return ""
	}
	var builder strings.Builder

	fmt.Fprintf(&builder, "Pool.Creations: %d\n", m.PoolCreations)
	fmt.Fprintf(&builder, "Pool.Returns: %d\n", m.PoolReturns)
	fmt.Fprintf(&builder, "Pool.Size: %d\n", m.PoolSize)
	fmt.Fprintf(&builder, "Trie.Nodes: %d\n", m.Nodes)
	fmt.Fprintf(&builder, "Trie.Patterns: %d\n", m.Patterns)
	fmt.Fprintf(&builder, "Trie.Hits: %d\n", m.Hits)
	fmt.Fprintf(&builder, "Trie.Misses: %d\n", m.Misses)
	fmt.Fprintf(&builder, "Trie.Reloads: %d\n", m.Reloads)
	fmt.Fprintf(&builder, "Trie.Retries: %d\n", m.Retries)
	fmt.Fprintf(&builder, "Heap.Allocs: %d\n", m.HeapAllocs)
	fmt.Fprintf(&builder, "Heap.Frees: %d\n", m.HeapFrees)
	fmt.Fprintf(&builder, "GC.PauseTotalNs: %d\n", m.GCPauseTotalNs)

	return builder.String()
} // String()

/* _EoF_ */
