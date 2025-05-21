/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_tNode_Equal(t *testing.T) {
	tests := []struct {
		name  string
		node  *tNode
		other *tNode
		want  bool
	}{
		/* */
		{
			name:  "01 - equal",
			node:  newNode(),
			other: newNode(),
			want:  true,
		},
		{
			name:  "02 - not equal",
			node:  newNode(),
			other: &tNode{terminator: endMask},
			want:  false,
		},
		{
			name:  "03 - nil node",
			node:  nil,
			other: newNode(),
			want:  false,
		},
		{
			name:  "04 - nil other",
			node:  newNode(),
			other: nil,
			want:  false,
		},
		{
			name:  "05 - nil node and other",
			node:  nil,
			other: nil,
			want:  true,
		},
		{
			name:  "06 - same object",
			node:  newNode(),
			other: newNode(),
			want:  true,
		},
		{
			name: "07 - different children",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain"})
				return n
			}(),
			other: func() *tNode {
				n := newNode()
				n.add(tPartsList{"domain"})
				return n
			}(),
			want: false,
		},
		{
			name: "08 - different number of children",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld"})
				return n
			}(),
			other: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain"})
				return n
			}(),
			want: false,
		},
		{
			name: "09 - different children order",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain"})
				return n
			}(),
			other: func() *tNode {
				n := newNode()
				n.add(tPartsList{"domain", "tld"})
				return n
			}(),
			want: false,
		},
		{
			name: "10 - different children values",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld"})
				n.tChildren["tld"].terminator = endMask
				return n
			}(),
			other: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld"})
				n.tChildren["tld"].terminator = 0
				return n
			}(),
			want: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.node.Equal(tc.other); got != tc.want {
				t.Errorf("tNode.Equal() = `%v`, want `%v`",
					got, tc.want)
			}
		})
	}
} // Test_tNode_Equal()

func Test_tNode_String(t *testing.T) {
	tests := []struct {
		name string
		node *tNode
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
			want: "\"Node\":\n  isEnd: false\n  isWild: false\n",
		},
		{
			name: "03 - node with children",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld"})
				return n
			}(),
			want: "\"Node\":\n  isEnd: false\n  isWild: false\n  \"tld\":\n      isEnd: true\n      isWild: false\n",
		},
		{
			name: "04 - node with wildcard",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"*"})
				return n
			}(),
			want: "\"Node\":\n  isEnd: false\n  isWild: false\n  \"*\":\n      isEnd: false\n      isWild: true\n",
		},
		{
			name: "05 - node with multiple children",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld"})
				n.add(tPartsList{"tld2"})
				n.add(tPartsList{"tld3"})
				return n
			}(),
			want: "\"Node\":\n  isEnd: false\n  isWild: false\n  \"tld\":\n      isEnd: true\n      isWild: false\n  \"tld2\":\n      isEnd: true\n      isWild: false\n  \"tld3\":\n      isEnd: true\n      isWild: false\n",
		},
		{
			name: "06 - node with multiple levels",
			node: &tNode{
				tChildren: tChildren{"tld": &tNode{
					tChildren: tChildren{"domain": &tNode{terminator: endMask}},
				}},
			},
			want: "\"Node\":\n  isEnd: false\n  isWild: false\n  \"tld\":\n      isEnd: false\n      isWild: false\n      \"domain\":\n          isEnd: true\n          isWild: false\n",
		},
		{
			name: "07 - node with multiple children&grandchildren",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld1", "domain1", "sub1", "host1"})
				n.add(tPartsList{"tld2", "domain2", "sub2", "*"})
				n.add(tPartsList{"tld3", "domain3", "*"})
				n.add(tPartsList{"tld4", "*"})
				return n
			}(),
			want: "\"Node\":\n  isEnd: false\n  isWild: false\n  \"tld1\":\n      isEnd: false\n      isWild: false\n      \"domain1\":\n          isEnd: false\n          isWild: false\n          \"sub1\":\n              isEnd: false\n              isWild: false\n              \"host1\":\n                  isEnd: true\n                  isWild: false\n  \"tld2\":\n      isEnd: false\n      isWild: false\n      \"domain2\":\n          isEnd: false\n          isWild: false\n          \"sub2\":\n              isEnd: false\n              isWild: false\n              \"*\":\n                  isEnd: false\n                  isWild: true\n  \"tld3\":\n      isEnd: false\n      isWild: false\n      \"domain3\":\n          isEnd: false\n          isWild: false\n          \"*\":\n              isEnd: false\n              isWild: true\n  \"tld4\":\n      isEnd: false\n      isWild: false\n      \"*\":\n          isEnd: false\n          isWild: true\n",
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.node.String(); got != tc.want {
				t.Errorf("tNode.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tNode_String()

func Test_tNode_add(t *testing.T) {
	tests := []struct {
		name  string
		node  *tNode
		parts tPartsList
		want  bool
	}{
		/* */
		{
			name:  "01 - add nil node",
			node:  nil,
			parts: tPartsList{"tld"},
			want:  false,
		},
		{
			name:  "02 - add empty parts",
			node:  newNode(),
			parts: nil,
			want:  false,
		},
		{
			name:  "03 - add single part",
			node:  newNode(),
			parts: tPartsList{"tld"},
			want:  true,
		},
		{
			name:  "04 - add wildcard",
			node:  newNode(),
			parts: tPartsList{"*"},
			want:  true,
		},
		{
			name:  "05 - add FQDN",
			node:  newNode(),
			parts: tPartsList{"tld", "domain", "host"},
			want:  true,
		},
		{
			name:  "06 - add wildcard host",
			node:  newNode(),
			parts: tPartsList{"tld", "domain", "*"},
			want:  true,
		},
		{
			name:  "07 - add existing part with wildcard",
			node:  &tNode{tChildren: tChildren{"tld": newNode()}},
			parts: tPartsList{"tld", "*"},
			want:  true,
		},
		/* */
		{
			name:  "08 - add existing parts",
			node:  &tNode{tChildren: tChildren{"tld": &tNode{tChildren: tChildren{"domain": newNode()}}}},
			parts: tPartsList{"tld", "domain"},
			want:  true,
		},
		{
			name:  "09 - add existing wildcard",
			node:  &tNode{tChildren: tChildren{"*": newNode()}},
			parts: tPartsList{"*"},
			want:  true,
		},
		/* */
		{
			name:  "10 - add wildcard after part",
			node:  &tNode{tChildren: tChildren{"tld": newNode()}},
			parts: tPartsList{"*"},
			want:  true,
		},
		{
			name: "11 - node with child, grandchild, wildcard, and child",
			node: func() *tNode {
				n := newNode()
				n.tChildren["tld"] = &tNode{
					tChildren: tChildren{
						"domain": &tNode{
							tChildren: tChildren{
								"*": &tNode{
									tChildren: tChildren{
										"sub": newNode(),
									},
									terminator: wildMask,
								},
							},
						},
					},
				}

				return n
			}(),
			parts: tPartsList{"tld", "domain", "*", "sub", "host"},
			want:  true,
		},
		{
			name: "12 - node with children, and grandchildren",
			node: func() *tNode {
				n := newNode()
				n.tChildren["tld"] = &tNode{
					tChildren: tChildren{
						"domain": &tNode{
							tChildren: tChildren{
								"sub": newNode(),
							},
						},
					},
				}

				return n
			}(),
			parts: tPartsList{"tld", "domain", "sub", "host"},
			want:  true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.node.add(tc.parts)
			if got != tc.want {
				t.Errorf("tNode.add() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_tNode_add()

func Test_tNode_allPatterns(t *testing.T) {
	tests := []struct {
		name string
		node *tNode
		want tPartsList
	}{
		/* */
		{
			name: "01 - nil node",
			node: nil,
			want: nil,
		},
		{
			name: "02 - nil patterns",
			node: func() *tNode {
				n := newNode()
				n.terminator = endMask
				return n
			}(),
			want: nil, //tPartsList{},
		},
		{
			name: "03 - empty patterns",
			node: func() *tNode {
				n := newNode()
				n.terminator = endMask
				return n
			}(),
			want: nil, //tPartsList{},
		},
		{
			name: "04 - node with child, grandchild, wildcard, and child",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "sub", "*"})
				return n
			}(),
			want: func() tPartsList {
				p := strings.Join(tPartsList{"*", "sub", "domain", "tld"}, ".")
				pl := tPartsList{p}
				return pl
			}(),
		},
		{
			name: "05 - node with multiple children",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld1", "domain1", "sub1", "host1"})
				n.add(tPartsList{"tld2", "domain2", "sub2", "*"})
				n.add(tPartsList{"tld3", "domain3", "host3", "*"})
				return n
			}(),
			want: func() tPartsList {
				p1 := strings.Join(
					tPartsList{"host1", "sub1", "domain1", "tld1"}, ".")
				p2 := strings.Join(
					tPartsList{"*", "sub2", "domain2", "tld2"}, ".")
				p3 := strings.Join(
					tPartsList{"*", "host3", "domain3", "tld3"}, ".")
				pl := tPartsList{p1, p2, p3}
				return pl
			}(),
		},
		{
			name: "06 - node with multiple children and grandchildren",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld1", "domain1", "sub1", "host1"})
				n.add(tPartsList{"tld1", "domain1", "sub2", "*"})
				n.add(tPartsList{"tld2", "domain2", "host3", "*"})
				n.add(tPartsList{"tld2", "domain2", "sub4", "host4"})
				return n
			}(),
			want: func() tPartsList {
				p1 := strings.Join(
					tPartsList{"host1", "sub1", "domain1", "tld1"}, ".")
				p2 := strings.Join(
					tPartsList{"*", "sub2", "domain1", "tld1"}, ".")
				p3 := strings.Join(
					tPartsList{"*", "host3", "domain2", "tld2"}, ".")
				p4 := strings.Join(
					tPartsList{"host4", "sub4", "domain2", "tld2"}, ".")
				pl := tPartsList{p1, p2, p3, p4}
				return pl
			}(),
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.node.allPatterns()

			if nil == got {
				if nil != tc.want {
					t.Error("\ntNode.allPatterns() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("\ntNode.allPatterns() = %q, want 'nil'",
					got.String())
				return
			}
			if !tc.want.Equal(got) {
				t.Errorf("\ntNode.allPatterns() =\n%q\nwant\n%q",
					got.String(), tc.want.String())
			}
		})
	}
} // Test_tNode_allPatterns()

func Test_tNode_clone(t *testing.T) {
	km1a := tChildren{"tld": &tNode{terminator: endMask}}
	km1b := tChildren{"domain": &tNode{terminator: endMask}}

	tests := []struct {
		name string
		node *tNode
		want *tNode
	}{
		/* */
		{
			name: "01 - clone nil",
			node: nil,
			want: nil,
		},
		{
			name: "02 - clone",
			node: &tNode{
				tChildren:  km1a,
				terminator: endMask | wildMask,
			},
			want: &tNode{
				tChildren:  km1a,
				terminator: endMask | wildMask,
			},
		},
		{
			name: "03 - clone with children",
			node: &tNode{
				tChildren:  km1a,
				terminator: endMask | wildMask,
			},
			want: &tNode{
				tChildren:  km1a,
				terminator: endMask | wildMask,
			},
		},
		{
			name: "04 - clone with multiple children",
			node: &tNode{
				tChildren:  tChildren{"tld": &tNode{terminator: endMask}, "domain": &tNode{terminator: endMask}},
				terminator: endMask | wildMask,
			},
			want: &tNode{
				tChildren:  tChildren{"tld": &tNode{terminator: endMask}, "domain": &tNode{terminator: endMask}},
				terminator: endMask | wildMask,
			},
		},
		{
			name: "05 - clone with multiple levels",
			node: &tNode{
				tChildren: tChildren{"tld": &tNode{
					tChildren: km1b,
				}},
				terminator: endMask | wildMask,
			},
			want: &tNode{
				tChildren: tChildren{"tld": &tNode{
					tChildren: km1b,
				}},
				terminator: endMask | wildMask,
			},
		},
		{
			name: "06 - clone with wildcard",
			node: &tNode{
				tChildren:  tChildren{"*": &tNode{terminator: wildMask}},
				terminator: endMask | wildMask,
			},
			want: &tNode{
				tChildren:  tChildren{"*": &tNode{terminator: wildMask}},
				terminator: endMask | wildMask,
			},
		},
		{
			name: "07 - clone with wildcard and children",
			node: &tNode{
				tChildren:  tChildren{"*": &tNode{terminator: wildMask}, "tld": &tNode{terminator: endMask}},
				terminator: endMask | wildMask,
			},
			want: &tNode{
				tChildren:  tChildren{"*": &tNode{terminator: wildMask}, "tld": &tNode{terminator: endMask}},
				terminator: endMask | wildMask,
			},
		},
		{
			name: "08 - clone with wildcard and multiple children",
			node: &tNode{
				tChildren:  tChildren{"*": &tNode{terminator: wildMask}, "tld": &tNode{terminator: endMask}, "domain": &tNode{terminator: endMask}},
				terminator: endMask | wildMask,
			},
			want: &tNode{
				tChildren:  tChildren{"*": &tNode{terminator: wildMask}, "tld": &tNode{terminator: endMask}, "domain": &tNode{terminator: endMask}},
				terminator: endMask | wildMask,
			},
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.node.clone()

			if nil == got {
				if nil != tc.want {
					t.Error("tNode.clone() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("tNode.clone() =\n%q\nwant\nnil",
					got.String())
				return
			}
			if !got.Equal(tc.want) {
				t.Errorf("tNode.clone() =\n%q\nwant\n%q",
					got.String(), tc.want.String())
				return
			}
		})
	}
} // Test_tNode_clone()

func Test_tNode_count(t *testing.T) {
	tests := []struct {
		name         string
		node         *tNode
		wantNodes    int
		wantPatterns int
	}{
		/* */
		{
			name:         "01 - nil node",
			node:         nil,
			wantNodes:    0,
			wantPatterns: 0,
		},
		{
			name:         "02 - empty node",
			node:         newNode(),
			wantNodes:    1,
			wantPatterns: 0,
		},
		{
			name: "03 - one pattern",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld"})
				return n
			}(),
			wantNodes:    1,
			wantPatterns: 1,
		},
		{
			name: "04 - two patterns",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld"})
				n.add(tPartsList{"tld", "domain"})
				return n
			}(),
			wantNodes:    2,
			wantPatterns: 2,
		},
		{
			name: "05 - three patterns",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld"})
				n.add(tPartsList{"tld", "domain"})
				n.add(tPartsList{"tld", "domain", "sub"})
				return n
			}(),
			wantNodes:    4,
			wantPatterns: 3,
		},
		{
			name: "06 - two patterns incl. wildcard",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "sub"})
				n.add(tPartsList{"tld", "domain", "sub", "*"})
				return n
			}(),
			wantNodes:    4,
			wantPatterns: 2,
		},
		{
			name: "07 - two patterns incl. wildcard and one more",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "sub"})
				n.add(tPartsList{"tld", "domain", "sub", "host", "*"})
				return n
			}(),
			wantNodes:    5,
			wantPatterns: 2,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotNodes, gotPatterns := tc.node.count()

			if nil == tc.node {
				if 0 != gotNodes {
					t.Errorf("tNode.count() Nodes = %d, want %d",
						gotNodes, 0)
				}
				if 0 != gotPatterns {
					t.Errorf("tNode.count() Patterns = %d, want %d", gotPatterns, 0)
				}
				return
			}
			if 0 == gotNodes {
				if 0 != tc.wantNodes {
					t.Errorf("tNode.count() Nodes = %d, want %d",
						gotNodes, tc.wantNodes)
				}
				return
			}
			if 0 == gotPatterns {
				if 0 != tc.wantPatterns {
					t.Errorf("tNode.count() Patterns = %d, want %d", gotPatterns, tc.wantPatterns)
				}
				return
			}
			if 0 == tc.wantNodes {
				if 0 != gotNodes {
					t.Errorf("tNode.count() Nodes = %d, want %d",
						gotNodes, 0)
				}
				return
			}
			if 0 == tc.wantPatterns {
				if 0 != gotPatterns {
					t.Errorf("tNode.count() Patterns = %d, want %d", gotPatterns, 0)
				}
				return
			}
			if gotPatterns >= gotNodes {
				t.Errorf("tNode.count() Patterns -ge Nodes: %d >= %d",
					gotPatterns, gotNodes)
			}
		})
	}
} // Test_tNode_count()

func Test_tNode_delete(t *testing.T) {
	tests := []struct {
		name  string
		aNode *tNode
		parts tPartsList
		want  bool
	}{
		/* */
		{
			name:  "01 - delete nil node",
			aNode: nil,
			parts: nil,
			want:  false,
		},
		{
			name:  "02 - delete empty node",
			aNode: newNode(),
			parts: nil,
			want:  false,
		},
		{
			name: "03 - delete non-existent part",
			aNode: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld"})
				return n
			}(),
			parts: tPartsList{"domain"},
			want:  false,
		},
		{
			name: "04 - delete existing part",
			aNode: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld"})
				return n
			}(),
			parts: tPartsList{"tld"},
			want:  true,
		},
		{
			name: "05 - delete existing part with children",
			aNode: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain"})
				return n
			}(),
			parts: tPartsList{"tld"},
			want:  false,
		},
		{
			name: "06 - delete existing part with children and delete child",
			aNode: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain"})
				return n
			}(),
			parts: tPartsList{"tld", "domain"},
			want:  true,
		},
		{
			name: "07 - delete FQDN part with children from wildcard",
			aNode: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "*"})
				return n
			}(),
			parts: tPartsList{"tld", "domain", "host"},
			want:  false, // no matching pattern
		},
		{
			name: "08 - delete wildcard from FQDN",
			aNode: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "host"})
				return n
			}(),
			parts: tPartsList{"tld", "domain", "*"},
			want:  false, // no matching pattern
		},
		{
			name: "09 - delete wildcard from FQDN with children",
			aNode: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "*"})
				n.add(tPartsList{"tld", "domain", "host", "sub"})
				return n
			}(),
			parts: tPartsList{"tld", "domain", "*"},
			want:  true, // matching pattern
		},
		{
			name: "10 - delete wildcard from FQDN with children and delete child",
			aNode: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "*"})
				n.add(tPartsList{"tld", "domain", "host", "sub"})
				return n
			}(),
			parts: tPartsList{"tld", "domain", "host", "sub"},
			want:  true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.aNode.delete(tc.parts)

			if got != tc.want {
				t.Errorf("tNode.delete() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_tNode_delete()

func Test_tNode_forEach(t *testing.T) {
	tests := []struct {
		name  string
		node  *tNode
		aFunc func(aNode *tNode)
	}{
		/* */
		{
			name:  "01 - nil node",
			node:  nil,
			aFunc: func(aNode *tNode) {},
		},
		{
			name:  "02 - nil function",
			node:  newNode(),
			aFunc: nil,
		},
		{
			name:  "03 - empty node",
			node:  newNode(),
			aFunc: func(aNode *tNode) {},
		},
		{
			name: "04 - node with child",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain"})
				return n
			}(),
			aFunc: func(aNode *tNode) {},
		},
		{
			name: "05 - node with wildcard",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "*"})
				return n
			}(),
			aFunc: func(aNode *tNode) {},
		},
		{
			name: "06 - node with children",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "host"})
				return n
			}(),
			aFunc: func(aNode *tNode) {},
		},
		/* */
		// TODO: Add test cases.
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.node.forEach(tc.aFunc)
		})
	}
} // Test_tNode_forEach()

func Test_tNode_load(t *testing.T) {
	tests := []struct {
		name    string
		node    *tNode
		reader  io.Reader
		wantErr bool
	}{
		/* */
		{
			name:    "01 - nil node",
			node:    nil,
			reader:  strings.NewReader("tld"),
			wantErr: true,
		},
		{
			name:    "02 - nil reader",
			node:    newNode(),
			reader:  nil,
			wantErr: true,
		},
		{
			name:    "03 - empty reader",
			node:    newNode(),
			reader:  strings.NewReader(""),
			wantErr: false,
		},
		{
			name:    "04 - reader with comments",
			node:    newNode(),
			reader:  strings.NewReader("# comment\n; comment\n# the next line is no comment\n comment"),
			wantErr: false,
		},
		{
			name:    "05 - reader with empty lines",
			node:    newNode(),
			reader:  strings.NewReader("\n\n\n"),
			wantErr: false,
		},
		{
			name:    "06 - reader with valid data",
			node:    newNode(),
			reader:  strings.NewReader("tld\ndomain.tld\nhost.domain.tld\ninvalid\n*.domain.tld"),
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.node.load(tc.reader)
			if (nil != err) != tc.wantErr {
				t.Errorf("tNode.load() error = '%v', wantErr '%v'",
					err, tc.wantErr)
			}
		})
	}
} // Test_tNode_load()

func Test_tNode_match(t *testing.T) {
	tests := []struct {
		name  string
		node  *tNode
		parts tPartsList
		want  bool
	}{
		/* */
		{
			name:  "01 - match nil node",
			node:  nil,
			parts: tPartsList{},
			want:  false,
		},
		{
			name:  "02 - match nil parts",
			node:  newNode(),
			parts: nil,
			want:  false,
		},
		{
			name:  "03 - match empty parts",
			node:  newNode(),
			parts: tPartsList{},
			want:  false,
		},
		{
			name: "04 - match non-existent part",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld"})
				return n
			}(),
			parts: tPartsList{"domain"},
			want:  false,
		},
		{
			name: "05 - match existing part",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain"})
				return n
			}(),
			parts: tPartsList{"tld", "domain"},
			want:  true,
		},
		{
			name: "06 - match wildcard against FQDN",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "host"})
				return n
			}(),
			parts: tPartsList{"tld", "domain", "*"},
			want:  false,
		},
		/* */
		{
			name: "07 - match FQDN against wildcard",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "*"})
				return n
			}(),
			parts: tPartsList{"tld", "domain", "host"},
			want:  true,
		},
		/* */
		{
			name: "08 - match FQDN against wildcard and FQDN",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "*"})
				n.add(tPartsList{"tld", "domain", "host"})
				return n
			}(),
			parts: tPartsList{"tld", "domain", "host"},
			want:  true,
		},
		{
			name: "09 - match FQDN against wildcards",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "*"})
				n.add(tPartsList{"tld", "domain", "host", "*"})
				return n
			}(),
			parts: tPartsList{"tld", "domain", "host"},
			want:  true,
		},
		{
			name: "10 - match FQDN against wildcard and FQDN and wildcards",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "*"})
				n.add(tPartsList{"tld", "domain", "host"})
				n.add(tPartsList{"tld", "domain", "host", "*"})
				return n
			}(),
			parts: tPartsList{"tld", "domain", "host"},
			want:  true,
		},
		{
			name: "11 - match FQDN against wildcard and FQDN and wildcards and FQDN",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "*"})
				n.add(tPartsList{"tld", "domain", "host", "*"})
				n.add(tPartsList{"tld", "domain", "sub", "host"})
				return n
			}(),
			parts: tPartsList{"tld", "domain", "host"},
			want:  true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.node.match(tc.parts)

			if got != tc.want {
				t.Errorf("tNode.match() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_tNode_match()

func Test_tNode_store(t *testing.T) {
	tests := []struct {
		name     string
		node     *tNode
		ip       string
		wantText string
		wantErr  bool
	}{
		/* */
		{
			name:     "01 - save nil node",
			node:     nil,
			ip:       "",
			wantText: "",
			wantErr:  true,
		},
		{
			name:     "02 - save empty node",
			node:     newNode(),
			ip:       "",
			wantText: "",
			wantErr:  false,
		},
		{
			name: "03 - save node with child",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain"})
				return n
			}(),
			ip:       "",
			wantText: "domain.tld\n",
			wantErr:  false,
		},
		{
			name: "04 - save node with wildcard",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "*"})
				return n
			}(),
			ip:       "",
			wantText: "*.tld\n",
			wantErr:  false,
		},
		{
			name: "05 - save node with children",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "host"})
				return n
			}(),
			ip:       "",
			wantText: "host.domain.tld\n",
			wantErr:  false,
		},
		{
			name: "06 - save node with wildcard and children",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "*"})
				n.add(tPartsList{"tld", "domain", "host"})
				return n
			}(),
			ip:       "",
			wantText: "*.domain.tld\nhost.domain.tld\n",
			wantErr:  false,
		},
		{
			name: "07 - save node with child",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain"})
				return n
			}(),
			ip:       "0.0.0.0",
			wantText: "0.0.0.0 domain.tld\n",
			wantErr:  false,
		},
		{
			name: "08 - save node with wildcard",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "*"})
				return n
			}(),
			ip:       "0.0.0.0",
			wantText: "0.0.0.0 *.tld\n",
			wantErr:  false,
		},
		{
			name: "09 - save node with children",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "host"})
				return n
			}(),
			ip:       "0.0.0.0",
			wantText: "0.0.0.0 host.domain.tld\n",
			wantErr:  false,
		},
		{
			name: "10 - save node with wildcard and children",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain", "*"})
				n.add(tPartsList{"tld", "domain", "host"})
				return n
			}(),
			ip:       "0.0.0.0",
			wantText: "0.0.0.0 *.domain.tld\n0.0.0.0 host.domain.tld\n",
			wantErr:  false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			aWriter := &bytes.Buffer{}
			err := tc.node.store(aWriter, tc.ip)

			if (nil != err) != tc.wantErr {
				t.Errorf("tNode.store() error = %v, wantErr %v",
					err, tc.wantErr)
				return
			}

			gotText := aWriter.String()
			if gotText != tc.wantText {
				t.Errorf("tNode.store() =\n%q\nwant\n%q",
					gotText, tc.wantText)
			}
		})
	}
} // Test_tNode_store()

func Test_tNode_update(t *testing.T) {
	tests := []struct {
		name     string
		node     *tNode
		oldParts tPartsList
		newParts tPartsList
		want     bool
	}{
		/* */
		{
			name:     "01 - update nil node",
			node:     nil,
			oldParts: tPartsList{"tld"},
			newParts: tPartsList{"tld"},
			want:     false,
		},
		{
			name:     "02 - update nil old parts",
			node:     newNode(),
			oldParts: nil,
			newParts: tPartsList{"tld"},
			want:     false,
		},
		{
			name:     "03 - update nil new parts",
			node:     newNode(),
			oldParts: tPartsList{"tld"},
			newParts: nil,
			want:     false,
		},
		{
			name:     "04 - update empty old parts",
			node:     newNode(),
			oldParts: tPartsList{},
			newParts: tPartsList{"tld"},
			want:     false,
		},
		{
			name:     "05 - update empty new parts",
			node:     newNode(),
			oldParts: tPartsList{"tld"},
			newParts: tPartsList{},
			want:     false,
		},
		{
			name:     "06 - update equal old and new parts",
			node:     newNode(),
			oldParts: tPartsList{"tld"},
			newParts: tPartsList{"tld"},
			want:     false,
		},
		/* */
		{
			name: "07 - update non-existent old parts",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld"})
				return n
			}(),
			oldParts: tPartsList{"domain"},
			newParts: tPartsList{"tld"},
			want:     true,
		},
		{
			name: "08 - update existing old parts",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain"})
				return n
			}(),
			oldParts: tPartsList{"tld", "domain"},
			newParts: tPartsList{"tld", "domain", "sub"},
			want:     true,
		},
		/* */
		{
			name: "09 - update existing old parts to existing new parts",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain"})
				n.add(tPartsList{"tld", "domain", "sub"})
				return n
			}(),
			oldParts: tPartsList{"tld", "domain"},
			newParts: tPartsList{"tld", "domain", "sub"},
			want:     true,
		},
		/* */
		{
			name: "10 - update/replace existing old parts to existing new parts with wildcard",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain"})
				n.add(tPartsList{"tld", "domain", "sub"})
				return n
			}(),
			oldParts: tPartsList{"tld", "domain"},
			newParts: tPartsList{"tld", "domain", "*"},
			want:     true,
		},
		/* */
		{
			name: "11 - update existing old parts to existing new parts with wildcard",
			node: func() *tNode {
				n := newNode()
				n.add(tPartsList{"tld", "domain"})
				n.add(tPartsList{"tld", "domain", "sub"})
				return n
			}(),
			oldParts: tPartsList{"tld", "domain"},
			newParts: tPartsList{"tld", "domain", "sub", "*"},
			want:     true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.node.update(tc.oldParts, tc.newParts); got != tc.want {
				t.Errorf("tNode.update() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_tNode_update()

func Test_pattern2parts(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    *tPartsList
	}{
		{
			name:    "empty pattern",
			pattern: "",
			want:    nil,
		},
		{
			name:    "tld",
			pattern: "tld",
			want:    &tPartsList{"tld"},
		},
		{
			name:    "domain.tld",
			pattern: "domain.tld",
			want:    &tPartsList{"tld", "domain"},
		},
		{
			name:    "sub.domain.tld",
			pattern: "sub.domain.tld",
			want:    &tPartsList{"tld", "domain", "sub"},
		},
		{
			name:    "host.sub.domain.tld",
			pattern: "host.sub.domain.tld",
			want:    &tPartsList{"tld", "domain", "sub", "host"},
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := pattern2parts(tc.pattern)
			if nil == got {
				if nil != tc.want {
					t.Error("pattern2parts() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("pattern2parts() =\n%q\nwant 'nil'",
					got.String())
				return
			}

			if !tc.want.Equal(got) {
				t.Errorf("pattern2parts() =\n%q\nwant\n%q",
					got.String(), tc.want.String())
			}
		})
	}
} // Test_pattern2parts()

/* _EoF_ */
