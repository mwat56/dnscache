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

func Test_newTrieNode(t *testing.T) {
	initReal() // initialise the node pool
	tests := []struct {
		name     string
		wantNode *tTrieNode
	}{
		{
			name:     "01 - empty node",
			wantNode: &tTrieNode{},
		},
		{
			name:     "02 - new node",
			wantNode: newTrieNode(),
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotNode := newTrieNode()

			if nil == gotNode {
				if nil != tc.wantNode {
					t.Error("newTrieNode() = nil, want non-nil")
				}
				return
			}
			if nil == tc.wantNode {
				t.Errorf("newTrieNode() =\n%v\nwant 'nil'",
					gotNode)
				return
			}
			if 0 != len(gotNode.tChildren) {
				t.Errorf("newTrieNode() = %v, want empty children",
					gotNode.tChildren)
			}

			if !tc.wantNode.Equal(gotNode) {
				t.Errorf("newTrieNode() =\n%q\nwant\n%q",
					gotNode.String(), tc.wantNode.String())
			}
		})
	}
} // Test_newTrieNode()

func Test_poolMetrics(t *testing.T) {
	clear := func() {
		// These tests would succeed only if the test was run as part
		// of only this file's tests, but would fail when run as part
		// of the package's whole test suite, as the pool is
		// initialised only once and the pool metric's numbers will be
		// influenced by other tests. To circumvent this, we reset the
		// pool's metrics to a known state: empty pool, no creations
		// or returns.
		np := triePool
		for range len(np.nodes) {
			_ = np.Get()
		}
		np.created.Store(0)
		np.returned.Store(0)
	}

	tests := []struct {
		name    string
		prepare func()
		want    *tPoolMetrics
	}{
		{
			name:    "01 - empty pool",
			prepare: clear,
			want: &tPoolMetrics{
				Created:  0,
				Returned: 0,
				Size:     0,
			},
		},
		{
			name: "02 - pool with one node",
			prepare: func() {
				clear()
				_ = newTrieNode()
			},
			want: &tPoolMetrics{
				Created:  1,
				Returned: 0,
				Size:     0,
			},
		},
		{
			name: "03 - pool with two nodes",
			prepare: func() {
				clear()
				_ = newTrieNode()
				n2 := newTrieNode()
				triePool.Put(n2)
			},
			want: &tPoolMetrics{
				Created:  2,
				Returned: 1,
				Size:     1,
			},
		},
		{
			name: "04 - pool with two nodes",
			prepare: func() {
				clear()
				n1 := newTrieNode()
				n2 := newTrieNode()
				triePool.Put(n1)
				triePool.Put(n2)
			},
			want: &tPoolMetrics{
				Created:  2,
				Returned: 2,
				Size:     2,
			},
		},
		{
			name: "05 - pool with three nodes",
			prepare: func() {
				clear()
				n1 := newTrieNode()
				_ = newTrieNode()
				n2 := newTrieNode()
				_ = newTrieNode()
				triePool.Put(n1)
				triePool.Put(n2)
			},
			want: &tPoolMetrics{
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
			got := poolMetrics()

			if nil == got {
				if nil != tc.want {
					t.Error("poolMetrics() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("poolMetrics() =\n%v\nwant 'nil'",
					got)
				return
			}
			if got.Created != tc.want.Created {
				t.Errorf("poolMetrics() Created = %d, want %d",
					got.Created, tc.want.Created)
			}
			if got.Returned != tc.want.Returned {
				t.Errorf("poolMetrics() Returned = %d, want %d",
					got.Returned, tc.want.Returned)
			}
			if got.Size != tc.want.Size {
				t.Errorf("poolMetrics() Size = %d, want %d",
					got.Size, tc.want.Size)
			}
		})
	}
} // Test_poolMetrics()

/* _EoF_ */
