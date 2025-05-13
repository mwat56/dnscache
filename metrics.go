/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package dnscache

import (
	"fmt"
	"strings"
	"sync/atomic"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `TMetrics` contains the metrics data for the DNS cache.
	//
	// These are the public fields to access the metrics data:
	//
	//   - `Lookups`: Total number of lookups,
	//   - `Hits`: Number of cache hits,
	//   - `Misses`: Number of cache misses,
	//   - `Retries`: Number of lookup retries,
	//   - `Errors`: Number of lookup errors,
	//   - `Peak`: Peak number of cached entries.
	TMetrics struct {
		Lookups uint32
		Hits    uint32
		Misses  uint32
		Retries uint32
		Errors  uint32
		Peak    uint32
	}
)

var (
	// `gMetrics` is the global metrics instance that is used
	// to store current metrics data.
	//
	// This variable should be considered a R/O singleton.
	gMetrics = &TMetrics{}
)

// ---------------------------------------------------------------------------
// `TMetrics` methods:

// `check()` verifies the consistency of the metrics data.
//
// Parameters:
//   - `aCorrect`: `true` to correct any inconsistencies,
//     `false` to just check for them.
//
// Returns:
//   - `int`: `0` if the metrics data is consistent,
//     `1` if the metrics data is `nil`,
//     `2` if the number of lookups is less than the sum of hits and misses,
//     `3` if the peak number of cached entries is less than the number of hits.
func (m *TMetrics) check(aCorrect bool) int {
	if nil == m {
		return 1
	}

	if m.Lookups != (m.Hits + m.Misses) {
		if aCorrect {
			if m.Lookups < (m.Hits + m.Misses) {
				setMetricsFieldMax(&m.Lookups, m.Hits+m.Misses)
			} else {
				setMetricsFieldMax(&m.Hits, m.Lookups-m.Misses)
			}
		}
		return 2
	}

	if m.Peak < m.Hits {
		if aCorrect {
			setMetricsFieldMax(&m.Peak, m.Hits)
		}
		return 3
	}

	return 0
} // check()

// `clone()` returns the current metrics data.
//
// Returns:
//   - `*TMetrics`: Current metrics data.
func (m *TMetrics) clone() *TMetrics {
	return &TMetrics{
		Lookups: atomic.LoadUint32(&m.Lookups),
		Hits:    atomic.LoadUint32(&m.Hits),
		Misses:  atomic.LoadUint32(&m.Misses),
		Retries: atomic.LoadUint32(&m.Retries),
		Errors:  atomic.LoadUint32(&m.Errors),
		Peak:    atomic.LoadUint32(&m.Peak),
	}
} // clone()

// `String()` implements the `fmt.Stringer` interface for the metrics data.
//
// Returns:
//   - `string`: String representation of the metrics data.
func (m *TMetrics) String() string {
	if nil == m {
		return ""
	}
	if err := m.check(true); 0 != err {
		return fmt.Sprintf("invalid metrics data (err %d)", err)
	}

	var builder strings.Builder

	fmt.Fprintf(&builder, "Lookups: %d\n", m.Lookups)
	fmt.Fprintf(&builder, "Hits: %d\n", m.Hits)
	fmt.Fprintf(&builder, "Misses: %d\n", m.Misses)
	fmt.Fprintf(&builder, "Retries: %d\n", m.Retries)
	fmt.Fprintf(&builder, "Errors: %d\n", m.Errors)
	fmt.Fprintf(&builder, "Peak: %d\n", m.Peak)

	return builder.String()
} // String()

// ---------------------------------------------------------------------------
// Helper functions:

// `incMetricsFields()` increments the given metrics fields by one.
//
// Parameters:
//   - `aFields`: List of pointers to the metrics fields to increment.
func incMetricsFields(aFields ...*uint32) {
	for _, field := range aFields {
		atomic.AddUint32(field, 1)
	}
} // incMetricsFields()

// `setMetricsFieldMax()` sets the value of a metrics field to
// the maximum of its current value and the given value.
//
// Parameters:
//   - `aField`: Pointer to the metrics field to update.
//   - `aValue`: New value to compare with the current value.
func setMetricsFieldMax(aField *uint32, aValue uint32) {
	atomic.StoreUint32(aField, max(atomic.LoadUint32(aField), aValue))
} // setMetricsFieldMax()

/* _EoF_ */
