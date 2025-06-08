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

func Test_tCacheNode_Add(t *testing.T) {
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
				n.Add(context.TODO(), tPartsList{"tld", "domain", "host"},
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
			gotOK := tc.node.Add(context.TODO(), tc.partList, tc.ips, 0)

			if gotOK != tc.wantOK {
				t.Errorf("tCacheNode.Add() = %v, want %v",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_tCacheNode_Add()

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
				n.Add(context.TODO(), tPartsList{"tld", "domain", "sub", "*"},
					tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")}, 0)
				return n
			}(),
			wantList: tPatternList{"*.sub.domain.tld"},
		},
		{
			name: "04 - node with child, grandchild, wildcard, and child",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld", "domain", "sub", "*"},
					tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain", "sub", "host"},
					tIpList{net.ParseIP("3.4.5.6"), net.ParseIP("4.5.6.7")}, 0)
				return n
			}(),
			wantList: tPatternList{"*.sub.domain.tld", "host.sub.domain.tld"},
		},
		{
			name: "05 - node with child, grandchild, wildcard, and child",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld", "domain", "sub", "*"},
					tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain", "sub", "host"},
					tIpList{net.ParseIP("3.4.5.6"), net.ParseIP("4.5.6.7")}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain", "sub", "host", "sub"},
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
				n.Add(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			wantNodes:    1,
			wantPatterns: 1,
		},
		{
			name: "04 - two patterns",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4"),
						net.ParseIP("2.3.4.5")}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("3.4.5.6"),
						net.ParseIP("4.5.6.7")}, 0)
				return n
			}(),
			wantNodes:    2,
			wantPatterns: 2,
		},
		{
			name: "05 - three patterns",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4"),
						net.ParseIP("1.2.3.4")}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4"),
						net.ParseIP("1.2.3.4")}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain", "sub"},
					tIpList{net.ParseIP("1.2.3.4"),
						net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			wantNodes:    3,
			wantPatterns: 3,
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
				n.Add(context.TODO(), tPartsList{"tld"},
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
				n.Add(context.TODO(), tPartsList{"tld"}, tIpList{}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
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
				n.Add(context.TODO(), tPartsList{"tld"},
					tIpList{}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4"),
						net.ParseIP("1.2.3.4")}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain", "host"},
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
				n.Add(context.TODO(), tPartsList{"tld"},
					tIpList{}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
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
				n.Add(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain", "host"},
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
				n.Add(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, 0)
				return n
			}(),
			other: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, 0)
				n.Add(context.TODO(), tPartsList{"tld"},
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
				n.Add(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, -time.Hour)
				return n
			}(),
			wantOK: true,
		},
		{
			name: "04 - non-expired node",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, time.Hour)
				return n
			}(),
			wantOK: false,
		},
		{
			name: "05 - expired child",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, -time.Hour)
				return n
			}(),
			wantOK: true,
		},
		{
			name: "06 - non-expired child",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("1.2.3.4")}, time.Hour)
				return n
			}(),
			wantOK: false,
		},
		{
			name: "07 - expired child, non-expired grandchild",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, time.Hour)
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, -time.Hour)
				n.Add(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("3.4.5.6")}, time.Hour)
				return n
			}(),
			wantOK: true,
		},
		{
			name: "08 - expired parent, non-expired child & grandchild",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, -time.Hour)
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, time.Hour)
				n.Add(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("3.4.5.6")}, time.Hour)
				return n
			}(),
			wantOK: true,
		},
		{
			name: "09 - expired child, non-expired parent & grandchild",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, time.Hour)
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, -time.Hour)
				n.Add(context.TODO(), tPartsList{"tld", "domain", "host"},
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

func Test_tCacheNode_Match(t *testing.T) {
	tests := []struct {
		name     string
		node     *tCacheNode
		partList tPartsList
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
		{
			name: "05 - match existing part",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			partList: tPartsList{"tld"},
			wantOK:   true,
		},
		{
			name: "06 - match existing part with child",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld"}, tIpList{}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("2.3.4.5")}, 0)
				return n
			}(),
			partList: tPartsList{"tld", "domain"},
			wantOK:   true,
		},
		{
			name: "07 - match existing part with child and grandchild",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld"}, tIpList{}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain"}, tIpList{}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain", "host"},
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
			gotOK := tc.node.Match(context.TODO(), tc.partList)

			if gotOK != tc.wantOK {
				t.Errorf("tCacheNode.Match() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_tCacheNode_Match()

func Test_tCacheNode_Read(t *testing.T) {
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
				n.Add(context.TODO(), tPartsList{"tld"},
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
			gotIPs := tc.node.Read(context.TODO(), tc.partList)
			if nil == gotIPs {
				if nil != tc.wantIPs {
					t.Error("tCacheNode.Read() = nil, want non-nil")
				}
				return
			}
			if nil == tc.wantIPs {
				t.Errorf("tCacheNode.Read() = %v, want 'nil'",
					gotIPs)
				return
			}
			if !tc.wantIPs.Equal(gotIPs) {
				t.Errorf("tCacheNode.Read() = %v, want %v",
					gotIPs, tc.wantIPs)
			}
		})
	}
} // Test_tCacheNode_Read()

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
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
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
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain", "host"},
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
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4"), net.ParseIP("5.6.7.8")}, 0)
				n.Add(context.TODO(), tPartsList{"tld", "domain", "host"},
					tIpList{net.ParseIP("2.3.4.5"), net.ParseIP("6.7.8.9")}, 0)
				return n
			}(),
			wantText: "1.2.3.4 domain.tld\n5.6.7.8 domain.tld\n2.3.4.5 host.domain.tld\n6.7.8.9 host.domain.tld\n",
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
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			ips: tIpList{},
			ttl: time.Minute,
			want: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{}, 0)
				return n
			}(),
		},
		{
			name: "04 - update with nil IPs",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			ips: nil,
			ttl: time.Minute,
			want: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{}, 0)
				return n
			}(),
		},
		{
			name: "05 - update with different IPs",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			ips: tIpList{net.ParseIP("5.6.7.8")},
			ttl: time.Minute,
			want: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("5.6.7.8")}, 0)
				return n
			}(),
		},
		{
			name: "06 - update with different TTL",
			node: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			ips: tIpList{net.ParseIP("1.2.3.4")},
			ttl: 2 * time.Minute,
			want: func() *tCacheNode {
				n := newNode()
				n.Add(context.TODO(), tPartsList{"tld", "domain"},
					tIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.node.Update(tc.ips, tc.ttl)
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
