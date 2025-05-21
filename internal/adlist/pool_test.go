/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_newNode(t *testing.T) {
	initReal() // initialise the node pool
	tests := []struct {
		name     string
		wantNode *tNode
	}{
		{
			name:     "01 - empty node",
			wantNode: &tNode{tChildren: make(tChildren)},
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotNode := newNode()
			if nil == gotNode {
				if nil != tc.wantNode {
					t.Error("newNode() = nil, want non-nil")
				}
				return
			}
			if nil == tc.wantNode {
				t.Errorf("newNode() = %q, want 'nil'",
					gotNode.String())
				return
			}
			if 0 != len(gotNode.tChildren) {
				t.Errorf("newNode() = %v, want empty children",
					gotNode.tChildren)
			}
			if gotNode.isEnd {
				t.Errorf("newNode() = %v, want `false` `isEnd`",
					gotNode.isEnd)
			}
			if gotNode.isWild {
				t.Errorf("newNode() = %v, want `false` `isWild`",
					gotNode.isWild)
			}
			if !tc.wantNode.Equal(gotNode) {
				t.Errorf("newNode() =\n%q\nwant\n%q",
					gotNode.String(), tc.wantNode.String())
			}
		})
	}
} // Test_newNode()

func Test_poolMetrics(t *testing.T) {
	func() {
		// These tests would succeed only if the test was run as part
		// of only this file's tests, but would fail when run as part
		// of the package's whole test suite, as the pool is
		// initialised only once and the pool metric's numbers will be
		// influenced by other tests. To circumvent this, we reset the
		// pool's metrics to a known state: empty pool, no creations
		// or returns.
		np := nodePool
		for range len(np.nodes) {
			_ = np.Get()
		}
		np.created.Store(0)
		np.returned.Store(0)
	}()

	tests := []struct {
		name         string
		wantCreated  uint32
		wantReturned uint32
		wantSize     int
	}{
		/* */
		{
			name:         "01 - empty pool",
			wantCreated:  10, // item created by test 02
			wantReturned: 3,  // 256 during init plus items from test 02, 03
			wantSize:     2,  // items from tests 02 and 03
		},
		{
			name:        "02 - pool with one item",
			wantCreated: 10, // item created by test 02
			wantReturned: func() uint32 {
				_ = nodePool.Get()
				nodePool.Put(&tNode{tChildren: make(tChildren)})
				return 3 // 256 during init plus items from test 02, 03
			}(),
			wantSize: 2, // items from tests 02 and 03
		},
		{
			name:        "03 - pool with three items",
			wantCreated: 10, // item created by test 02
			wantReturned: func() uint32 {
				for range 10 {
					_ = nodePool.Get()
				}
				nodePool.Put(&tNode{tChildren: make(tChildren)})
				nodePool.Put(&tNode{tChildren: make(tChildren)})
				return 3 // 256 during init plus items from test 02, 03
			}(),
			wantSize: 2, // items from tests 02 and 03
		},
		{
			name:         "04 - pool with two items",
			wantCreated:  10, // item created by test 02
			wantReturned: 3,  // 256 during init plus items from test 02, 03
			wantSize:     2,  // items from tests 02 and 03
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotMetrics := poolMetrics()

			if nil == gotMetrics {
				t.Error("poolMetrics() = nil, want non-nil")
				return
			}
			if gotMetrics.Created != tc.wantCreated {
				t.Errorf("poolMetrics() gotCreated = %d, want %d",
					gotMetrics.Created, tc.wantCreated)
			}
			if gotMetrics.Returned != tc.wantReturned {
				t.Errorf("poolMetrics() gotReturned = %d, want %d",
					gotMetrics.Returned, tc.wantReturned)
			}
			if gotMetrics.Size != tc.wantSize {
				t.Errorf("poolMetrics() gotSize = %d, want %d",
					gotMetrics.Size, tc.wantSize)
			}
		})
	}
} // Test_poolMetrics()

/* _EoF_ */
