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

func Test_tCacheNode_add(t *testing.T) {
	tests := []struct {
		name     string
		node     *tCacheNode
		partList tPartsList
		ips      TIpList
		wantOK   bool
	}{
		/* */
		{
			name:     "01 - nil node",
			node:     &tCacheNode{},
			partList: tPartsList{"tld"},
			ips:      TIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")},
			wantOK:   true,
		},
		{
			name:     "02 - nil parts",
			node:     newNode(),
			partList: nil,
			ips:      TIpList{net.ParseIP("1.2.3.4")},
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
			ips:      TIpList{net.ParseIP("1.2.3.4")},
			wantOK:   false,
		},
		{
			name:     "05 - empty IPs",
			node:     newNode(),
			partList: tPartsList{"tld"},
			ips:      TIpList{},
			wantOK:   true,
		},
		{
			name:     "06 - add single part",
			node:     newNode(),
			partList: tPartsList{"tld"},
			ips:      TIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")},
			wantOK:   true,
		},
		{
			name:     "07 - add FQDN",
			node:     newNode(),
			partList: tPartsList{"tld", "domain", "host"},
			ips:      TIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")},
			wantOK:   true,
		},
		{
			name: "08 - add existing FQDN",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain", "host"},
					TIpList{}, 0)
				return n
			}(),
			partList: tPartsList{"tld", "domain", "host"},
			ips:      TIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")},
			wantOK:   true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := tc.node.add(context.TODO(), tc.partList, tc.ips, 0)

			if gotOK != tc.wantOK {
				t.Errorf("tCacheNode.add() = %v, want %v",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_tCacheNode_add()

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
				n.add(context.TODO(), tPartsList{"tld", "domain", "sub", "*"},
					TIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")}, 0)
				return n
			}(),
			wantList: tPatternList{"*.sub.domain.tld"},
		},
		{
			name: "04 - node with child, grandchild, wildcard, and child",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain", "sub", "*"},
					TIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain", "sub", "host"},
					TIpList{net.ParseIP("3.4.5.6"), net.ParseIP("4.5.6.7")}, 0)
				return n
			}(),
			wantList: tPatternList{"*.sub.domain.tld", "host.sub.domain.tld"},
		},
		{
			name: "05 - node with child, grandchild, wildcard, and child",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain", "sub", "*"},
					TIpList{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain", "sub", "host"},
					TIpList{net.ParseIP("3.4.5.6"), net.ParseIP("4.5.6.7")}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain", "sub", "host", "sub"},
					TIpList{net.ParseIP("5.6.7.8"), net.ParseIP("6.7.8.9")}, 0)
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
				n.add(context.TODO(), tPartsList{"tld"},
					TIpList{net.ParseIP("1.2.3.4"), net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			wantNodes:    1,
			wantPatterns: 1,
		},
		{
			name: "04 - two patterns",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld"},
					TIpList{net.ParseIP("1.2.3.4"),
						net.ParseIP("2.3.4.5")}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("3.4.5.6"),
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
				n.add(context.TODO(), tPartsList{"tld"},
					TIpList{net.ParseIP("1.2.3.4"),
						net.ParseIP("1.2.3.4")}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("1.2.3.4"),
						net.ParseIP("1.2.3.4")}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain", "sub"},
					TIpList{net.ParseIP("1.2.3.4"),
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

func Test_tCacheNode_delete(t *testing.T) {
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
				n.add(context.TODO(), tPartsList{"tld"},
					TIpList{net.ParseIP("1.2.3.4"),
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
				n.add(context.TODO(), tPartsList{"tld"}, TIpList{}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("1.2.3.4"),
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
				n.add(context.TODO(), tPartsList{"tld"},
					TIpList{}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("1.2.3.4"),
						net.ParseIP("1.2.3.4")}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain", "host"},
					TIpList{net.ParseIP("2.3.4.5"),
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
				n.add(context.TODO(), tPartsList{"tld"},
					TIpList{}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("1.2.3.4"),
						net.ParseIP("2.3.4.5")}, 0)
				return n
			}(),
			partList: tPartsList{"tld", "domain"},
			wantOK:   true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := tc.node.delete(context.TODO(), tc.partList)
			if gotOK != tc.wantOK {
				t.Errorf("tCacheNode.delete() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_tCacheNode_delete()

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
			other:  &tCacheNode{tCachedIP: tCachedIP{TIpList: TIpList{net.ParseIP("1.2.3.4")}}},
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
			other:  &tCacheNode{tCachedIP: tCachedIP{TIpList: TIpList{net.ParseIP("1.2.3.4")}}},
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
							TIpList: TIpList{net.ParseIP("1.2.3.4")},
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
				n.add(context.TODO(), tPartsList{"tld"},
					TIpList{net.ParseIP("1.2.3.4")}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("2.3.4.5")}, 0)
				return n
			}(),
			other: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("2.3.4.5")}, 0)
				n.add(context.TODO(), tPartsList{"tld"},
					TIpList{net.ParseIP("1.2.3.4")}, 0)
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
		/* */
		{
			name: "03 - expired node",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld"},
					TIpList{net.ParseIP("1.2.3.4")}, -time.Hour)
				return n
			}(),
			wantOK: true,
		},
		/* */
		{
			name: "04 - non-expired node",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld"},
					TIpList{net.ParseIP("1.2.3.4")}, time.Hour)
				return n
			}(),
			wantOK: false,
		},
		{
			name: "05 - expired child",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("1.2.3.4")}, -time.Hour)
				return n
			}(),
			wantOK: true,
		},
		{
			name: "06 - non-expired child",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain", "host"},
					TIpList{net.ParseIP("1.2.3.4")}, time.Hour)
				return n
			}(),
			wantOK: false,
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

func Test_tCacheNode_ips(t *testing.T) {
	tests := []struct {
		name     string
		node     *tCacheNode
		partList tPartsList
		wantIPs  TIpList
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
				n.add(context.TODO(), tPartsList{"tld"},
					TIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			partList: tPartsList{"tld"},
			wantIPs:  TIpList{net.ParseIP("1.2.3.4")},
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotIPs := tc.node.ips(context.TODO(), tc.partList)
			if nil == gotIPs {
				if nil != tc.wantIPs {
					t.Error("tCacheNode.ips() = nil, want non-nil")
				}
				return
			}
			if nil == tc.wantIPs {
				t.Errorf("tCacheNode.ips() = %v, want 'nil'",
					gotIPs)
				return
			}
			if !tc.wantIPs.Equal(gotIPs) {
				t.Errorf("tCacheNode.ips() = %v, want %v",
					gotIPs, tc.wantIPs)
			}
		})
	}
} // Test_tCacheNode_ips()

func Test_tCacheNode_match(t *testing.T) {
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
				n.add(context.TODO(), tPartsList{"tld"},
					TIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			partList: tPartsList{"tld"},
			wantOK:   true,
		},
		{
			name: "06 - match existing part with child",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld"}, TIpList{}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("2.3.4.5")}, 0)
				return n
			}(),
			partList: tPartsList{"tld", "domain"},
			wantOK:   true,
		},
		{
			name: "07 - match existing part with child and grandchild",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld"}, TIpList{}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain"}, TIpList{}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain", "host"},
					TIpList{net.ParseIP("3.4.5.6")}, 0)
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
			gotOK := tc.node.match(context.TODO(), tc.partList)

			if gotOK != tc.wantOK {
				t.Errorf("tCacheNode.match() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_tCacheNode_match()

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
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			wantText: "1.2.3.4 domain.tld\n",
			wantErr:  false,
		},
		{
			name: "04 - node with child and grandchild",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("1.2.3.4")}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain", "host"},
					TIpList{net.ParseIP("2.3.4.5")}, 0)
				return n
			}(),
			wantText: "1.2.3.4 domain.tld\n2.3.4.5 host.domain.tld\n",
			wantErr:  false,
		},
		{
			name: "05 - node with child and grandchild with multiple IPs",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("1.2.3.4"), net.ParseIP("5.6.7.8")}, 0)
				n.add(context.TODO(), tPartsList{"tld", "domain", "host"},
					TIpList{net.ParseIP("2.3.4.5"), net.ParseIP("6.7.8.9")}, 0)
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

func Test_tCacheNode_update(t *testing.T) {
	tests := []struct {
		name string
		node *tCacheNode
		ips  TIpList
		ttl  time.Duration
		want *tCacheNode
	}{
		/* */
		{
			name: "01 - nil node",
			node: nil,
			ips:  TIpList{net.ParseIP("1.2.3.4")},
			ttl:  time.Minute,
			want: nil,
		},
		{
			name: "02 - empty node",
			node: newNode(),
			ips:  TIpList{net.ParseIP("1.2.3.4")},
			ttl:  time.Minute,
			want: func() *tCacheNode {
				n := newNode()
				n.tCachedIP.TIpList = TIpList{net.ParseIP("1.2.3.4")}
				return n
			}(),
		},
		{
			name: "03 - update empty IPs",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			ips: TIpList{},
			ttl: time.Minute,
			want: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{}, 0)
				return n
			}(),
		},
		{
			name: "04 - update with nil IPs",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			ips: nil,
			ttl: time.Minute,
			want: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{}, 0)
				return n
			}(),
		},
		{
			name: "05 - update with different IPs",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			ips: TIpList{net.ParseIP("5.6.7.8")},
			ttl: time.Minute,
			want: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("5.6.7.8")}, 0)
				return n
			}(),
		},
		{
			name: "06 - update with different TTL",
			node: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
			ips: TIpList{net.ParseIP("1.2.3.4")},
			ttl: 2 * time.Minute,
			want: func() *tCacheNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain"},
					TIpList{net.ParseIP("1.2.3.4")}, 0)
				return n
			}(),
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.node.update(tc.ips, tc.ttl)
			if nil == got {
				if nil != tc.want {
					t.Error("tCacheNode.update() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("tCacheNode.update() = '%v', want 'nil'",
					got)
				return
			}
			if !tc.want.Equal(got) {
				t.Errorf("tCacheNode.update() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tCacheNode_update()

/* _EoF_ */
