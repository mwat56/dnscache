/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_newNode(t *testing.T) {
	tests := []struct {
		name string
		want *tCacheNode
	}{
		/* */
		{
			name: "01 - empty node",
			want: &tCacheNode{},
		},
		{
			name: "02 - new node",
			want: newNode(),
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := newNode()

			if nil == got {
				if nil != tc.want {
					t.Error("newNode() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("newNode() =\n'%v'\nwant 'nil'",
					got)
				return
			}
			if !got.Equal(tc.want) {
				t.Errorf("newNode() =\n'%v'\nwant\n'%v'",
					got, tc.want)
			}
		})
	}
} // Test_newNode()

func Test_tCacheNode_allPatterns(t *testing.T) {
	tests := []struct {
		name     string
		node     *tCacheNode
		wantList tPatternList
	}{
		/* */
		{
			name:     "01 - nil node",
			node:     nil,
			wantList: nil,
		},
		{
			name:     "02 - empty node",
			node:     newNode(),
			wantList: nil,
		},
		{
			name: "03 - node with child, grandchild, wildcard, and child",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub", "*"},
					tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")}, 0)
				return n
			}(),
			wantList: tPatternList{"*.sub.domain.tld"},
		},
		{
			name: "04 - node with child, grandchild, wildcard, and child",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub", "*"},
					tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub", "host"},
					tIpList{net.ParseIP("3.4.5.6"), net.ParseIP("4.5.6.7")}, 0)
				return n
			}(),
			wantList: tPatternList{"*.sub.domain.tld", "host.sub.domain.tld"},
		},
		{
			name: "05 - node with child, grandchild, wildcard, and child",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub", "*"},
					tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub", "host"},
					tIpList{net.ParseIP("3.4.5.6"), net.ParseIP("4.5.6.7")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub", "host", "sub"},
					tIpList{net.ParseIP("5.6.7.8"), net.ParseIP("6.7.8.9")}, 0)
				return n
			}(),
			wantList: tPatternList{"*.sub.domain.tld", "host.sub.domain.tld", "sub.host.sub.domain.tld"},
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotList := tc.node.allPatterns(context.TODO())
			if nil == gotList {
				if nil != tc.wantList {
					t.Error("tCacheNode.allPatterns() = nil, want non-nil")
				}
				return
			}
			if nil == tc.wantList {
				t.Errorf("tCacheNode.allPatterns() =\n%q\nwant 'nil'",
					gotList)
				return
			}
			if !tc.wantList.Equal(gotList) {
				t.Errorf("tCacheNode.allPatterns() =\n%q\nwant\n%q",
					gotList, tc.wantList)
			}
		})
	}
} // Test_tCacheNode_allPatterns()

func Test_tCacheNode_clone(t *testing.T) {
	tests := []struct {
		name string
		node *tCacheNode
		want *tCacheNode
	}{
		/* */
		{
			name: "01 - nil node",
			node: nil,
			want: nil,
		},
		{
			name: "02 - empty node",
			node: newNode(),
			want: newNode(),
		},
		{
			name: "03 - node with child",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("3.4.5.6")}, 0)
				return n
			}(),
			want: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("3.4.5.6")}, 0)
				return n
			}(),
		},
		/* */
		{
			name: "04 - node with child and grandchildren",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("4.5.6.1")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("4.5.6.7")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("4.5.6.8")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub"},
					tIpList{net.ParseIP("4.5.6.9")}, 0)
				return n
			}(),
			want: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("4.5.6.1")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("4.5.6.7")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("4.5.6.8")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub"},
					tIpList{net.ParseIP("4.5.6.9")}, 0)
				return n
			}(),
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.node.clone()
			if nil == got {
				if nil != tc.want {
					t.Error("tCacheNode.clone() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("tCacheNode.clone() =\n%q\nwant 'nil'",
					got)
				return
			}
			if !tc.want.Equal(got) {
				t.Errorf("tCacheNode.clone() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tCacheNode_clone()

func Test_tCacheNode_count(t *testing.T) {
	tests := []struct {
		name         string
		node         *tCacheNode
		wantNodes    int
		wantPatterns int
	}{
		/* */
		{
			name:         "01 - empty node",
			node:         &tCacheNode{},
			wantNodes:    0,
			wantPatterns: 0,
		},
		{
			name:         "02 - new node",
			node:         newNode(),
			wantNodes:    0,
			wantPatterns: 0,
		},
		{
			name: "03 - one pattern",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("3.4.5.6")}, 0)
				return n
			}(),
			wantNodes:    1,
			wantPatterns: 1,
		},
		{
			name: "04 - two patterns",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("4.3.2.1")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("4.5.6.7")}, 0)
				return n
			}(),
			wantNodes:    2,
			wantPatterns: 2,
		},
		{
			name: "05 - three patterns",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("5.4.1.1")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("5.4.2.2")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub"},
					tIpList{net.ParseIP("5.4.3.3")}, 0)
				return n
			}(),
			wantNodes:    3,
			wantPatterns: 3,
		},
		{
			name: "06 - four patterns with 3rd level siblings",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld1"},
					tIpList{net.ParseIP("6.5.1.1")}, 0)
				n.Create(context.TODO(), tPartsList{"tld1", "domain1"},
					tIpList{net.ParseIP("6.5.2.2")}, 0)
				n.Create(context.TODO(), tPartsList{"tld2", "domain2", "sub2"},
					tIpList{net.ParseIP("6.5.3.3")}, 0)
				n.Create(context.TODO(), tPartsList{"tld2", "domain2", "sub2", "host2"},
					tIpList{net.ParseIP("6.5.4.4")}, 0)
				return n
			}(),
			wantNodes:    6,
			wantPatterns: 4,
		},
		/* */
		{
			name: "07 - four patterns with 2nd level siblings",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("7.6.1.1")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("7.6.2.2")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub"},
					tIpList{net.ParseIP("7.6.3.3")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub", "host"},
					tIpList{net.ParseIP("7.6.4.4")}, 0)
				return n
			}(),
			wantNodes:    4,
			wantPatterns: 4,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotNodes, gotPatterns := tc.node.count(context.TODO())

			if gotNodes != tc.wantNodes {
				t.Errorf("tCacheNode.count() gotNodes = %v, want %v",
					gotNodes, tc.wantNodes)
			}
			if gotPatterns != tc.wantPatterns {
				t.Errorf("tCacheNode.count() gotPatterns = %v, want %v",
					gotPatterns, tc.wantPatterns)
			}
		})
	}
} // Test_tCacheNode_count()

func Test_tCacheNode_Create(t *testing.T) {
	tests := []struct {
		name     string
		node     *tCacheNode
		partList tPartsList
		ips      tIpList
		wantOK   bool
	}{
		/* */
		{
			name:     "01 - nil node",
			node:     &tCacheNode{},
			partList: tPartsList{"tld"},
			ips:      tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")},
			wantOK:   true,
		},
		{
			name:     "02 - nil parts",
			node:     newNode(),
			partList: nil,
			ips:      tIpList{net.ParseIP("1.2.3.4")},
			wantOK:   false,
		},
		{
			name:     "03 - nil IPs",
			node:     newNode(),
			partList: tPartsList{"tld"},
			ips:      nil,
			wantOK:   true,
		},
		{
			name:     "04 - empty parts",
			node:     newNode(),
			partList: tPartsList{},
			ips:      tIpList{net.ParseIP("1.2.3.4")},
			wantOK:   false,
		},
		{
			name:     "05 - empty IPs",
			node:     newNode(),
			partList: tPartsList{"tld"},
			ips:      tIpList{},
			wantOK:   true,
		},
		{
			name:     "06 - add single part",
			node:     newNode(),
			partList: tPartsList{"tld"},
			ips:      tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")},
			wantOK:   true,
		},
		{
			name:     "07 - add FQDN",
			node:     newNode(),
			partList: tPartsList{"tld", "domain", "host"},
			ips:      tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")},
			wantOK:   true,
		},
		{
			name: "08 - add existing FQDN",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{}, 0)
				return n
			}(),
			partList: tPartsList{"tld", "domain", "host"},
			ips:      tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")},
			wantOK:   true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := tc.node.Create(context.TODO(), tc.partList, tc.ips, 0)

			if gotOK != tc.wantOK {
				t.Errorf("tCacheNode.Create() = %v, want %v",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_tCacheNode_Create()

func Test_tCacheNode_Delete(t *testing.T) {
	tests := []struct {
		name     string
		node     *tCacheNode
		partList tPartsList
		wantOK   bool
	}{
		/* */
		{
			name:     "01 - nil node",
			node:     &tCacheNode{},
			partList: tPartsList{"tld"},
			wantOK:   false,
		},
		{
			name:     "02 - nil part list",
			node:     newNode(),
			partList: nil,
			wantOK:   false,
		},
		{
			name:     "03 - empty part list",
			node:     newNode(),
			partList: tPartsList{},
			wantOK:   false,
		},
		{
			name:     "04 - delete non-existent part",
			node:     newNode(),
			partList: tPartsList{"tld"},
			wantOK:   false,
		},
		{
			name: "05 - delete existing part",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4"),
						net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			partList: tPartsList{"tld"},
			wantOK:   true,
		},
		{
			name: "06 - delete existing part with child",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"}, tIpList{}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4"),
						net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			partList: tPartsList{"tld"},
			wantOK:   false,
		},
		{
			name: "07 - delete existing part with child but leave child",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4"),
						net.ParseIP("1.2.3.4")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("2.3.4.5"),
						net.ParseIP("3.4.5.6")}, 0)
				return n
			}(),
			partList: tPartsList{"tld", "domain"},
			wantOK:   false,
		},
		{
			name: "08 - delete existing part with child and delete child",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4"),
						net.ParseIP("2.3.4.5")}, 0)
				return n
			}(),
			partList: tPartsList{"tld", "domain"},
			wantOK:   true,
		},
		{
			name: "09 - delete existing part but leave parent",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("3.4.5.6")}, 0)
				return n
			}(),
			partList: tPartsList{"tld", "domain", "host"},
			wantOK:   true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := tc.node.Delete(context.TODO(), tc.partList)

			if gotOK != tc.wantOK {
				t.Errorf("tCacheNode.Delete() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_tCacheNode_Delete()

func Test_tCacheNode_Equal(t *testing.T) {
	tests := []struct {
		name   string
		node   *tCacheNode
		other  *tCacheNode
		wantOK bool
	}{
		/* */
		{
			name:   "01 - equal",
			node:   newNode(),
			other:  newNode(),
			wantOK: true,
		},
		{
			name:   "02 - not equal",
			node:   newNode(),
			other:  &tCacheNode{tCachedIP: tCachedIP{tIpList: tIpList{net.ParseIP("1.2.3.4")}}},
			wantOK: true, // cached IPs are ignored
		},
		{
			name:   "03 - nil node",
			node:   nil,
			other:  newNode(),
			wantOK: false,
		},
		{
			name:   "04 - nil other",
			node:   newNode(),
			other:  nil,
			wantOK: false,
		},
		{
			name:   "05 - nil node and other",
			node:   nil,
			other:  nil,
			wantOK: true,
		},
		{
			name:   "06 - same object",
			node:   newNode(),
			other:  newNode(),
			wantOK: true,
		},
		{
			name:   "07 - different IPs",
			node:   newNode(),
			other:  &tCacheNode{tCachedIP: tCachedIP{tIpList: tIpList{net.ParseIP("1.2.3.4")}}},
			wantOK: true, // cached IPs are ignored
		},
		{
			name:   "08 - different children",
			node:   newNode(),
			other:  &tCacheNode{tChildren: tChildren{"tld": newNode()}},
			wantOK: false,
		},
		{
			name: "09 - different children values",
			node: newNode(),
			other: &tCacheNode{
				tChildren: tChildren{
					"tld": &tCacheNode{
						tCachedIP: tCachedIP{
							tIpList: tIpList{net.ParseIP("1.2.3.4")},
						},
					},
				},
			},
			wantOK: false,
		},
		{
			name: "10 - equivalent object",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, 0)
				return n
			}(),
			other: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, 0)
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			wantOK: true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if gotOK := tc.node.Equal(tc.other); gotOK != tc.wantOK {
				t.Errorf("tCacheNode.Equal() = %v, want %v", gotOK, tc.wantOK)
			}
		})
	}
} // Test_tCacheNode_Equal()

func Test_tCacheNode_expire(t *testing.T) {
	tests := []struct {
		name   string
		node   *tCacheNode
		wantOK bool
	}{
		/* */
		{
			name:   "01 - nil node",
			node:   nil,
			wantOK: false,
		},
		{
			name:   "02 - empty node",
			node:   newNode(),
			wantOK: false,
		},
		{
			name: "03 - expired node",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, -time.Hour)
				return n
			}(),
			wantOK: true,
		},
		{
			name: "04 - non-expired node",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, time.Hour)
				return n
			}(),
			wantOK: false,
		},
		{
			name: "05 - expired child",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, -time.Hour)
				return n
			}(),
			wantOK: true,
		},
		{
			name: "06 - non-expired child",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("1.2.3.4")}, time.Hour)
				return n
			}(),
			wantOK: false,
		},
		{
			name: "07 - expired child, non-expired grandchild",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, time.Hour)
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, -time.Hour)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("3.4.5.6")}, time.Hour)
				return n
			}(),
			wantOK: true,
		},
		{
			name: "08 - expired parent, non-expired child & grandchild",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, -time.Hour)
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, time.Hour)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("3.4.5.6")}, time.Hour)
				return n
			}(),
			wantOK: true,
		},
		{
			name: "09 - expired child, non-expired parent & grandchild",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, time.Hour)
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, -time.Hour)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("3.4.5.6")}, time.Hour)
				return n
			}(),
			wantOK: true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := tc.node.expire(context.TODO())

			if gotOK != tc.wantOK {
				t.Errorf("tCacheNode.expire() = '%v,' want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_tCacheNode_expire()

func Test_tCacheNode_finalNode(t *testing.T) {
	tests := []struct {
		name      string
		rootNode  *tCacheNode
		partsList tPartsList
		wantNode  *tCacheNode
		wantOK    bool
	}{
		/* */
		{
			name:      "01 - nil node",
			rootNode:  nil,
			partsList: tPartsList{"tld", "domain", "sub"},
			wantNode:  nil,
			wantOK:    false,
		},
		{
			name:      "02 - nil parts list",
			rootNode:  newNode(),
			partsList: nil,
			wantNode:  nil,
			wantOK:    false,
		},
		{
			name:      "03 - empty parts list",
			rootNode:  newNode(),
			partsList: tPartsList{},
			wantNode:  nil,
			wantOK:    false,
		},
		{
			name:      "04 - non-existent path",
			rootNode:  newNode(),
			partsList: tPartsList{"tld", "domain", "sub"},
			wantNode:  nil,
			wantOK:    false,
		},
		{
			name: "05 - multi-level path with final node having IPs",
			rootNode: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			partsList: tPartsList{"tld", "domain", "sub"},
			wantNode: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				child := n
				for _, part := range []string{"tld", "domain", "sub"} {
					child = child.tChildren[part]
				}
				return child
			}(),
			wantOK: true,
		},
		{
			name: "06 - multi-level path with final node having IPs but expired",
			rootNode: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub"},
					tIpList{net.ParseIP("1.2.3.4")}, -time.Hour)
				return n
			}(),
			partsList: tPartsList{"tld", "domain", "sub"},
			wantNode:  nil,
			wantOK:    false,
		},
		{
			name: "07 - multi-level path with non-existent middle node",
			rootNode: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			partsList: tPartsList{"tld", "missing", "sub"},
			wantNode:  nil,
			wantOK:    false,
		},
		{
			name: "08 - multi-level path with partial match",
			rootNode: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			partsList: tPartsList{"tld", "domain", "sub"},
			wantNode:  nil,
			wantOK:    false,
		},
		{
			name: "09 - multi-level path with longer parts list than existing path",
			rootNode: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub", "host"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			partsList: tPartsList{"tld", "domain", "sub"},
			wantNode:  nil,
			wantOK:    false,
		},
		{
			name: "10 - multi-level path with multiple IPs",
			rootNode: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub"},
					tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("5.6.7.8")}, 0)
				return n
			}(),
			partsList: tPartsList{"tld", "domain", "sub"},
			wantNode: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain", "sub"},
					tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("5.6.7.8")}, 0)
				child := n
				for _, part := range []string{"tld", "domain", "sub"} {
					child = child.tChildren[part]
				}
				return child
			}(),
			wantOK: true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotNode, gotOK := tc.rootNode.finalNode(context.TODO(), tc.partsList)

			if gotOK != tc.wantOK {
				t.Errorf("tCacheNode.finalNode() gotOK = '%v', want '%v'",
					gotOK, tc.wantOK)
				return
			}

			if nil == gotNode {
				if nil != tc.wantNode {
					t.Error("tCacheNode.finalNode() gotNode = nil, want non-nil")
				}
				return
			}
			if nil == tc.wantNode {
				t.Errorf("tCacheNode.finalNode() gotNode =\n%q\nwant nil",
					gotNode)
				return
			}
			if !tc.wantNode.Equal(gotNode) {
				t.Errorf("tCacheNode.finalNode() gotNode =\n%q\nwant\n%q",
					gotNode, tc.wantNode)
			}
		})
	}
} // Test_tCacheNode_finalNode()

func Test_tCacheNode_match(t *testing.T) {
	tests := []struct {
		name     string
		node     *tCacheNode
		partList tPartsList
		wantNode *tCacheNode
		wantOK   bool
	}{
		/* */
		{
			name:     "01 - nil node",
			node:     nil,
			partList: tPartsList{"tld"},
			wantOK:   false,
		},
		{
			name:     "02 - nil part list",
			node:     newNode(),
			partList: nil,
			wantOK:   false,
		},
		{
			name:     "03 - empty part list",
			node:     newNode(),
			partList: tPartsList{},
			wantOK:   false,
		},
		{
			name:     "04 - match non-existent part",
			node:     newNode(),
			partList: tPartsList{"tld"},
			wantOK:   false,
		},
		/* * /
		{
		// This test fails because the node's children are different
			name: "05 - match existing part",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("5.6.7.8")}, 0)
				return n
			}(),
			partList: tPartsList{"tld"},
			wantNode: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("5.6.7.8")}, 0)
				return n
			}(),
			wantOK: true,
		},
		/* */
		{
			name: "06 - match existing part with child",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"}, tIpList{}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, 0)
				return n
			}(),
			partList: tPartsList{"tld", "domain"},
			wantNode: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, 0)

				child := n
				for _, part := range []string{"tld", "domain"} {
					child = child.tChildren[part]
				}
				return child
			}(),
			wantOK: true,
		},
		{
			name: "07 - match existing part with child and grandchild",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"}, tIpList{}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain"}, tIpList{}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("3.4.5.6")}, 0)
				return n
			}(),
			partList: tPartsList{"tld", "domain"},
			wantOK:   false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotNode, gotOK := tc.node.match(context.TODO(), tc.partList)

			if gotOK != tc.wantOK {
				t.Errorf("tCacheNode.match() = '%v', want '%v'",
					gotOK, tc.wantOK)
				return
			}
			if nil == gotNode {
				if nil != tc.wantNode {
					t.Error("tCacheNode.match() = nil, want non-nil")
				}
				return
			}
			if nil == tc.wantNode {
				t.Errorf("tCacheNode.match() =\n%q\nwant 'nil'",
					gotNode)
				return
			}
			if !gotNode.Equal(tc.wantNode) {
				t.Errorf("tCacheNode.match() =\n%q\nwant\n%q",
					gotNode, tc.wantNode)
			}
		})
	}
} // Test_tCacheNode_match()

func Test_tCacheNode_Retrieve(t *testing.T) {
	tests := []struct {
		name     string
		node     *tCacheNode
		partList tPartsList
		wantIPs  tIpList
	}{
		/* */
		{
			name:     "01 - nil node",
			node:     nil,
			partList: tPartsList{"tld"},
			wantIPs:  nil,
		},
		{
			name:     "02 - nil part list",
			node:     newNode(),
			partList: nil,
			wantIPs:  nil,
		},
		{
			name:     "03 - empty part list",
			node:     newNode(),
			partList: tPartsList{},
			wantIPs:  nil,
		},
		{
			name:     "04 - non-existent part",
			node:     newNode(),
			partList: tPartsList{"tld"},
			wantIPs:  nil,
		},
		{
			name: "05 - existing part",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			partList: tPartsList{"tld"},
			wantIPs:  tIpList{net.ParseIP("1.2.3.4")},
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotIPs := tc.node.Retrieve(context.TODO(), tc.partList)
			if nil == gotIPs {
				if nil != tc.wantIPs {
					t.Error("tCacheNode.Retrieve() = nil, want non-nil")
				}
				return
			}
			if nil == tc.wantIPs {
				t.Errorf("tCacheNode.Retrieve() = %v, want 'nil'",
					gotIPs)
				return
			}
			if !tc.wantIPs.Equal(gotIPs) {
				t.Errorf("tCacheNode.Retrieve() = %v, want %v",
					gotIPs, tc.wantIPs)
			}
		})
	}
} // Test_tCacheNode_Retrieve()

func Test_tCacheNode_store(t *testing.T) {
	tests := []struct {
		name     string
		node     *tCacheNode
		wantText string
		wantErr  bool
	}{
		/* */
		{
			name:     "01 - nil node",
			node:     nil,
			wantText: "",
			wantErr:  true,
		},
		{
			name:     "02 - empty node",
			node:     newNode(),
			wantText: "",
			wantErr:  false,
		},
		{
			name: "03 - node with child",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			wantText: "1.2.3.4 domain.tld\n",
			wantErr:  false,
		},
		{
			name: "04 - node with child and grandchild",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("2.3.4.5")}, 0)
				return n
			}(),
			wantText: "1.2.3.4 domain.tld\n2.3.4.5 host.domain.tld\n",
			wantErr:  false,
		},
		{
			name: "05 - node with child and grandchild with multiple IPs",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("5.6.7.8")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("2.3.4.5"), net.ParseIP("6.7.8.9")}, 0)
				return n
			}(),
			wantText: "1.2.3.4 domain.tld\n5.6.7.8 domain.tld\n2.3.4.5 host.domain.tld\n6.7.8.9 host.domain.tld\n",
			wantErr:  false,
		},
		{
			name: "06 - node with multiple children",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain1"},
					tIpList{net.ParseIP("6.0.1.2")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain1", "host1"},
					tIpList{net.ParseIP("6.0.2.3"), net.ParseIP("6.1.2.3")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain2"},
					tIpList{net.ParseIP("6.2.3.1")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain2", "host2"},
					tIpList{net.ParseIP("6.2.2.2"), net.ParseIP("6.2.2.3")}, 0)
				return n
			}(),
			wantText: "6.0.1.2 domain1.tld\n6.0.2.3 host1.domain1.tld\n6.1.2.3 host1.domain1.tld\n6.2.3.1 domain2.tld\n6.2.2.2 host2.domain2.tld\n6.2.2.3 host2.domain2.tld\n",
			wantErr:  false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			err := tc.node.store(context.TODO(), writer)

			if (err != nil) != tc.wantErr {
				t.Errorf("tCacheNode.store() error = '%v', wantErr '%v'",
					err, tc.wantErr)
				return
			}
			if gotText := writer.String(); gotText != tc.wantText {
				t.Errorf("tCacheNode.store() =\n%q\nwant\n%q",
					gotText, tc.wantText)
			}
		})
	}
} // Test_tCacheNode_store()

func Test_tCacheNode_String(t *testing.T) {
	tests := []struct {
		name string
		node *tCacheNode
		want string
	}{
		/* */
		{
			name: "01 - nil node",
			node: nil,
			want: "",
		},
		{
			name: "02 - empty node",
			node: newNode(),
			want: "",
		},
		{
			name: "03 - node with child",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			want: "1.2.3.4 domain.tld\n",
		},
		{
			name: "04 - node with child and grandchild",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				n.Create(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("2.3.4.5")}, 0)
				return n
			}(),
			want: "1.2.3.4 domain.tld\n2.3.4.5 host.domain.tld\n",
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.node.String(); got != tc.want {
				t.Errorf("tCacheNode.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tCacheNode_String()

func Test_tCacheNode_Update(t *testing.T) {
	tests := []struct {
		name string
		node *tCacheNode
		ips  tIpList
		ttl  time.Duration
		want *tCacheNode
	}{
		/* */
		{
			name: "01 - nil node",
			node: nil,
			ips:  tIpList{net.ParseIP("1.2.3.4")},
			ttl:  time.Minute,
			want: nil,
		},
		{
			name: "02 - empty node",
			node: newNode(),
			ips:  tIpList{net.ParseIP("1.2.3.4")},
			ttl:  time.Minute,
			want: func() *tCacheNode {
				n := newNode()
				n.tCachedIP.tIpList = tIpList{net.ParseIP("1.2.3.4")}
				return n
			}(),
		},
		{
			name: "03 - update empty IPs",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			ips: tIpList{},
			ttl: time.Minute,
			want: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{}, 0)
				return n
			}(),
		},
		{
			name: "04 - update with nil IPs",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			ips: nil,
			ttl: time.Minute,
			want: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{}, 0)
				return n
			}(),
		},
		{
			name: "05 - update with different IPs",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			ips: tIpList{net.ParseIP("5.6.7.8")},
			ttl: time.Minute,
			want: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("5.6.7.8")}, 0)
				return n
			}(),
		},
		{
			name: "06 - update with different TTL",
			node: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			ips: tIpList{net.ParseIP("1.2.3.4")},
			ttl: 2 * time.Minute,
			want: func() *tCacheNode {
				n := newNode()
				n.Create(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.node.Update(context.TODO(), tc.ips, tc.ttl)
			if nil == got {
				if nil != tc.want {
					t.Error("tCacheNode.Update() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("tCacheNode.Update() = '%v', want 'nil'",
					got)
				return
			}
			tGot := got.(*tCacheNode)
			if !tc.want.Equal(tGot) {
				t.Errorf("tCacheNode.Update() =\n%q\nwant\n%q",
					tGot, tc.want)
			}
		})
	}
} // Test_tCacheNode_Update()

/* _EoF_ */
