/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// `entryPoolDropMask` is the bit mask to use for limiting the pool returns.
	entryPoolDropMask = 7 // 0111

	// `entryPoolInitSize` is the number of items to pre-allocate for the
	// pool during initialisation.
	// Four times this value is used as the pool's maximum size.
	entryPoolInitSize = 1 << 9 // 512
)

type (
	// `tEntryPoolMetrics` contains the metrics data for the pool.
	//
	// These are the fields providing the metrics data:
	//
	//   - `Created`: Number of items created by the pool.
	//   - `Returned`: Number of items returned to the pool.
	//   - `Size`: Current number of items in the pool.
	tEntryPoolMetrics struct {
		Created  uint32
		Returned uint32
		Size     int
	}

	// `tMapPool` is a bounded pool of Map entries.
	//
	// The pool is inherently thread-safe. Its size is fixed and can't
	// be changed after creation.
	//
	// The pool's `get()` and `new()` methods return an unspecified item
	// (`any`) from the pool and the internal channel accepts `any` item.
	// That way the constructor function (calling the pool's `get()`
	// method) can differentiate between an item returned by the pool
	// and a newly created one.
	//
	tMapPool struct {
		new      func() any    // Factory function
		entries  chan any      // Bounded channel
		created  atomic.Uint32 // Number of items created
		returned atomic.Uint32 // Number of items returned
	}
)

var (
	// `mapPool` is the active pool of `tMapEntry` instances.
	mapPool *tMapPool

	// Make sure, the map entry pool is only initialised once.
	mapPoolInit sync.Once
)

// ---------------------------------------------------------------------------
// Initialise the map entry pool:

// `init()` pre-allocates some entries for the pool.
func init() {
	initEntryPool()
} // init()

// `initEntryPool()` pre-allocates some entries for the pool.
//
// This function is called only once during package initialisation.
//
// During unit testing, this function could be called manually.
func initEntryPool() {
	mapPoolInit.Do(func() {
		mapPool = &tMapPool{
			entries: make(chan any, int(entryPoolInitSize<<2)),
			new: func() any {
				entry := &tMapEntry{}
				mapPool.created.Add(1)
				//TODO: Go 1.24:
				// runtime.AddCleanup(entry, func() {
				// 	pool.put(entry)
				// })
				return entry
			},
		}

		for range entryPoolInitSize {
			// Pre-allocate some entries for the pool:
			mapPool.put(&tMapEntry{})
		}
	}) // mapPoolInit.Do()
} // initEntryPool()

// ---------------------------------------------------------------------------
// `tMapPool` methods:

// `get()` returns an item from the pool.
//
// If the pool is empty, a new item is created.
//
// Returns:
//   - `any`: An item from the pool.
func (ep *tMapPool) get() any {
	if nil == ep {
		initEntryPool() // initialise the entry pool
		ep = mapPool
	}

	select {
	case entry := <-ep.entries:
		return entry
	default:
		return ep.new()
	}
} // get()

// `put()` returns an item to the pool.
//
// If the pool is full, the item is dropped.
//
// Parameters:
//   - `aEntry`: The entry to return to the pool.
func (ep *tMapPool) put(aEntry *tMapEntry) {
	if nil == ep {
		initEntryPool() // initialise the entry pool
		ep = mapPool
	}
	if (ep.returned.Add(1) & entryPoolDropMask) == entryPoolDropMask {
		// Drop the entry if the drop mask matches.
		// This leaves the given `aEntry` for GC.
		// With a drop mask of `7` (0111) we drop 1 in 8 entries.
		runtime.GC()
		return
	}

	select {
	case ep.entries <- aEntry:
		// Item was added to pool
		runtime.Gosched()

	default:
		// Drop if pool is full
		runtime.Gosched()
	}
} // put()

// ---------------------------------------------------------------------------
// `tMapEntry` constructor:

// `newMapEntry()` returns a new `tMapEntry` instance.
//
// Returns:
//   - `*tMapEntry`: A new `tMapEntry` instance.
func newMapEntry() *tMapEntry {
	entry, ok := mapPool.get().(*tMapEntry)
	if ok {
		// Clear/reset the old field values
		entry.ips = tIpList{}
		entry.bestBefore = time.Time{}
	} else {
		entry = &tMapEntry{}
	}

	return entry
} // newMapEntry()

// ---------------------------------------------------------------------------
// Helper functions:

// `entryPoolMetrics()` returns the current pool metrics.
//
// Returns:
//   - `*tEntryPoolMetrics`: Current pool metrics.
func entryPoolMetrics() *tEntryPoolMetrics {
	if nil == mapPool {
		initEntryPool() // initialise the entry pool
	}
	ep := mapPool

	return &tEntryPoolMetrics{
		Created:  ep.created.Load(),
		Returned: ep.returned.Load(),
		Size:     len(ep.entries),
	}
} // entryPoolMetrics()

// `putEntry()` returns a entry to the pool.
//
// Parameters:
//   - `aEntry`: The entry to return to the pool.
func putEntry(aEntry *tMapEntry) {
	// We can't clear the entry's fields yet since it might
	// still be used by another list or goroutine.
	mapPool.put(aEntry)
} // putEntry()

/* _EoF_ */
