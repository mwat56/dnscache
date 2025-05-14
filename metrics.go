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
	// `gMetrics` is the global metrics instance that
	// is used to store the current metrics data.
	//
	// NOTE: This variable should be considered R/O.
	gMetrics = new(TMetrics)
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
//   - `bool`: `true` if the metrics data is consistent, `false` otherwise.
func (m *TMetrics) check(aCorrect bool) bool {
	if nil == m {
		if aCorrect {
			gMetrics = new(TMetrics)
		}
		return true
	}

	if m.Lookups != (m.Hits + m.Misses) {
		if aCorrect {
			if m.Lookups < (m.Hits + m.Misses) {
				setMetricsFieldMax(&m.Lookups, m.Hits+m.Misses)
			} else {
				setMetricsFieldMax(&m.Hits, m.Lookups-m.Misses)
			}
		}
		return false
	}

	return true
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

	return (m.Lookups == aMetrics.Lookups) &&
		(m.Hits == aMetrics.Hits) &&
		(m.Misses == aMetrics.Misses) &&
		(m.Retries == aMetrics.Retries) &&
		(m.Errors == aMetrics.Errors) &&
		(m.Peak == aMetrics.Peak)
} // Equal()

// `String()` implements the `fmt.Stringer` interface for the metrics data.
//
// Returns:
//   - `string`: String representation of the metrics data.
func (m *TMetrics) String() string {
	if nil == m {
		return ""
	}
	if ok := m.check(true); !ok {
		return "invalid metrics data"
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
