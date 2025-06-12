/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_newMapEntry(t *testing.T) {
	initEntryPool() // initialise the entry pool
	tests := []struct {
		name      string
		wantEntry *tMapEntry
	}{
		{
			name:      "01 - empty entry",
			wantEntry: &tMapEntry{},
		},
		{
			name:      "02 - new entry",
			wantEntry: newMapEntry(),
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotEntry := newMapEntry()

			if nil == gotEntry {
				if nil != tc.wantEntry {
					t.Error("newMapEntry() = nil, want non-nil")
				}
				return
			}
			if nil == tc.wantEntry {
				t.Errorf("newMapEntry() =\n%v\nwant 'nil'",
					gotEntry)
				return
			}
			if !tc.wantEntry.Equal(gotEntry) {
				t.Errorf("newMapEntry() =\n%q\nwant\n%q",
					gotEntry, tc.wantEntry)
			}
		})
	}
} // Test_newMapEntry()

func Test_entryPoolMetrics(t *testing.T) {
	clear := func() {
		// These tests would succeed only if the test was run as part
		// of only this file's tests, but would fail when run as part
		// of the package's whole test suite, as the pool is
		// initialised only once and the pool metric's numbers will be
		// influenced by other tests. To circumvent this, we reset the
		// pool's metrics to a known state: empty pool, no creations
		// or returns.
		ep := mapPool
		for range len(ep.entries) {
			_ = ep.get()
		}
		ep.created.Store(0)
		ep.returned.Store(0)
	}

	tests := []struct {
		name    string
		prepare func()
		want    *tEntryPoolMetrics
	}{
		{
			name:    "01 - empty pool",
			prepare: clear,
			want: &tEntryPoolMetrics{
				Created:  0,
				Returned: 0,
				Size:     0,
			},
		},
		{
			name: "02 - pool with one entry",
			prepare: func() {
				clear()
				_ = newMapEntry()
			},
			want: &tEntryPoolMetrics{
				Created:  1,
				Returned: 0,
				Size:     0,
			},
		},
		{
			name: "03 - pool with two entries",
			prepare: func() {
				clear()
				_ = newMapEntry()
				e2 := newMapEntry()
				mapPool.put(e2)
			},
			want: &tEntryPoolMetrics{
				Created:  2,
				Returned: 1,
				Size:     1,
			},
		},
		{
			name: "04 - pool with two entries",
			prepare: func() {
				clear()
				e1 := newMapEntry()
				e2 := newMapEntry()
				mapPool.put(e1)
				mapPool.put(e2)
			},
			want: &tEntryPoolMetrics{
				Created:  2,
				Returned: 2,
				Size:     2,
			},
		},
		{
			name: "05 - pool with three entries",
			prepare: func() {
				clear()
				e1 := newMapEntry()
				_ = newMapEntry()
				e2 := newMapEntry()
				_ = newMapEntry()
				mapPool.put(e1)
				mapPool.put(e2)
			},
			want: &tEntryPoolMetrics{
				Created:  4,
				Returned: 2,
				Size:     2,
			},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if nil != tc.prepare {
				tc.prepare()
			}
			got := entryPoolMetrics()

			if nil == got {
				if nil != tc.want {
					t.Error("entryPoolMetrics() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("entryPoolMetrics() =\n%v\nwant 'nil'",
					got)
				return
			}
			if got.Created != tc.want.Created {
				t.Errorf("entryPoolMetrics() Created = %d, want %d",
					got.Created, tc.want.Created)
			}
			if got.Returned != tc.want.Returned {
				t.Errorf("entryPoolMetrics() Returned = %d, want %d",
					got.Returned, tc.want.Returned)
			}
			if got.Size != tc.want.Size {
				t.Errorf("entryPoolMetrics() Size = %d, want %d",
					got.Size, tc.want.Size)
			}
		})
	}
} // Test_entryPoolMetrics()

/* _EoF_ */
