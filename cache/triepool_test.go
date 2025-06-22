/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"context"
	"net"
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_newTrieNode(t *testing.T) {
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

func Test_putNode(t *testing.T) {
	tests := []struct {
		name string
		node *tTrieNode
	}{
		{
			name: "01 - nil node",
			node: nil,
		},
		{
			name: "02 - empty node",
			node: &tTrieNode{tChildren: make(tChildren)},
		},
		{
			name: "03 - node with child",
			node: func() *tTrieNode {
				n := newTrieNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			putNode(tc.node)
		})
	}
} // Test_putNode()

/* _EoF_ */
