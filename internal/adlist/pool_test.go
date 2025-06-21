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

			if 0 != gotNode.terminator {
				t.Errorf("newNode() = %v, want `0` `terminator`",
					gotNode.terminator)
			}

			isEnd := ((gotNode.terminator & endMask) == endMask)
			if isEnd {
				t.Errorf("newNode() = %v, want `false` `isEnd`",
					isEnd)
			}
			if (gotNode.terminator & wildMask) == wildMask {
				t.Errorf("newNode() = %v, want `false` `isWild`",
					true)
			}
			if !tc.wantNode.Equal(gotNode) {
				t.Errorf("newNode() =\n%q\nwant\n%q",
					gotNode.String(), tc.wantNode.String())
			}
		})
	}
} // Test_newNode()

func Test_putNode(t *testing.T) {
	tests := []struct {
		name  string
		aNode *tNode
	}{
		{
			name:  "01 - nil node",
			aNode: nil,
		},
		{
			name:  "02 - empty node",
			aNode: &tNode{tChildren: make(tChildren)},
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			putNode(tc.aNode)
		})
	}
} // Test_putNode()

/* _EoF_ */
