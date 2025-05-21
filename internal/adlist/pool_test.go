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

func Test_tPool_Metrics(t *testing.T) {
	tests := []struct {
		name         string
		pool         *tPool
		wantCreated  uint32
		wantReturned uint32
		wantSize     int
	}{
		{
			name:         "01 - empty pool",
			pool:         newPool(10, nil),
			wantCreated:  0,
			wantReturned: 0,
			wantSize:     0,
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotCreated, gotReturned, gotSize := tc.pool.Metrics()

			if gotCreated != tc.wantCreated {
				t.Errorf("tPool.Metrics() gotCreated = %d, want %d",
					gotCreated, tc.wantCreated)
			}
			if gotReturned != tc.wantReturned {
				t.Errorf("tPool.Metrics() gotCreated = %d, want %d",
					gotReturned, tc.wantReturned)
			}
			if gotSize != tc.wantSize {
				t.Errorf("tPool.Metrics() gotSize = %d, want %d",
					gotSize, tc.wantSize)
			}
		})
	}
} // Test_tPool_Metrics()

/* _EoF_ */
