/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_tTrie_Equal(t *testing.T) {
	tc1 := newTrie()
	tc2 := newTrie()
	tc2.root.node.add(context.TODO(), tPartsList{"tld"})

	tests := []struct {
		name  string
		trie  *tTrie
		oTrie *tTrie
		want  bool
	}{
		{
			name:  "01 - nil trie",
			trie:  nil,
			oTrie: tc1,
			want:  false,
		},
		{
			name:  "02 - nil other trie",
			trie:  tc1,
			oTrie: nil,
			want:  false,
		},
		{
			name:  "03 - nil trie and other trie",
			trie:  nil,
			oTrie: nil,
			want:  true,
		},
		{
			name:  "04 - trie w/o root",
			trie:  &tTrie{},
			oTrie: tc2,
			want:  false,
		},
		{
			name:  "05 - other trie w/o root",
			trie:  tc1,
			oTrie: &tTrie{},
			want:  false,
		},
		{
			name:  "06 - same properties",
			trie:  tc1,
			oTrie: tc1,
			want:  true,
		},
		{
			name:  "07 - different properties",
			trie:  tc1,
			oTrie: tc2,
			want:  false,
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.trie.Equal(tc.oTrie); got != tc.want {
				t.Errorf("tTrie.Equal() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_tTrie_Equal()

func Test_tTrie_Add(t *testing.T) {
	tests := []struct {
		name     string
		trie     *tTrie
		pattern  string
		wantBool bool
	}{
		{
			name:     "01 - nil trie",
			trie:     nil,
			pattern:  "tld",
			wantBool: false,
		},
		{
			name:     "02 - nil root",
			trie:     &tTrie{},
			pattern:  "tld",
			wantBool: false,
		},
		{
			name:     "03 - empty pattern",
			trie:     newTrie(),
			pattern:  "",
			wantBool: false,
		},
		{
			name:     "04 - add tld",
			trie:     newTrie(),
			pattern:  "tld",
			wantBool: true,
		},
		{
			name:     "05 - add domain.tld",
			trie:     newTrie(),
			pattern:  "domain.tld",
			wantBool: true,
		},
		{
			name:     "06 - add sub.domain.tld",
			trie:     newTrie(),
			pattern:  "sub.domain.tld",
			wantBool: true,
		},
		{
			name:     "07 - add host.sub.domain.tld",
			trie:     newTrie(),
			pattern:  "host.sub.domain.tld",
			wantBool: true,
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotBool := tc.trie.Add(context.TODO(), tc.pattern)

			if gotBool != tc.wantBool {
				t.Errorf("tTrie.Add() gotBool = '%v', want '%v'",
					gotBool, tc.wantBool)
			}
		})
	}
} // Test_tTrie_Add()

func Test_tTrie_AllPatterns(t *testing.T) {
	tests := []struct {
		name     string
		trie     *tTrie
		wantList tPartsList
	}{
		/* */
		{
			name:     "01 - nil trie",
			trie:     nil,
			wantList: nil,
		},
		{
			name:     "02 - nil root",
			trie:     &tTrie{},
			wantList: nil,
		},
		{
			name:     "03 - empty trie",
			trie:     newTrie(),
			wantList: nil,
		},
		{
			name: "04 - one pattern",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				return t
			}(),
			wantList: tPartsList{"tld"},
		},
		{
			name: "05 - two patterns",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain"})
				return t
			}(),
			wantList: tPartsList{"tld", "domain.tld"},
		},
		{
			name: "06 - three patterns",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "sub"})
				return t
			}(),
			wantList: tPartsList{"tld", "domain.tld", "sub.domain.tld"},
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotList := tc.trie.AllPatterns(context.TODO())

			if nil == gotList {
				if nil != tc.wantList {
					t.Error("tTrie.AllPatterns() = nil, want non-nil")
				}
				return
			}
			if nil == tc.wantList {
				t.Errorf("tTrie.AllPatterns() =\n%q\nwant 'nil'",
					gotList.String())
				return
			}
			if !tc.wantList.Equal(gotList) {
				t.Errorf("tTrie.AllPatterns() =\n%q\nwant\n%q",
					gotList.String(), tc.wantList.String())
			}
		})
	}
} // Test_tTrie_AllPatterns()

func Test_tTrie_Delete(t *testing.T) {
	tests := []struct {
		name     string
		trie     *tTrie
		pattern  string
		wantBool bool
	}{
		/* */
		{
			name:     "01 - nil trie",
			trie:     nil,
			pattern:  "tld",
			wantBool: false,
		},
		{
			name:     "02 - nil root",
			trie:     &tTrie{},
			pattern:  "tld",
			wantBool: false,
		},
		{
			name:     "03 - empty pattern",
			trie:     newTrie(),
			pattern:  "",
			wantBool: false,
		},
		{
			name: "04 - delete tld",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				return t
			}(),
			pattern:  "tld",
			wantBool: true,
		},
		/* */
		{
			name: "05 - delete domain.tld",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain"})
				return t
			}(),
			pattern:  "domain.tld",
			wantBool: true,
		},
		/* */
		{
			name: "06 - delete host.domain.tld",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "host"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "*"})
				return t
			}(),
			pattern:  "host.domain.tld",
			wantBool: true,
		},
		{
			name: "07 - delete *.domain.tld",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "host"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "*"})
				return t
			}(),
			pattern:  "*.domain.tld",
			wantBool: true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotBool := tc.trie.Delete(context.TODO(), tc.pattern)

			if gotBool != tc.wantBool {
				t.Errorf("tTrie.Delete() gotBool = '%v', want '%v'",
					gotBool, tc.wantBool)
			}
		})
	}
} // Test_tTrie_Delete()

func Test_tTrie_Count(t *testing.T) {
	tests := []struct {
		name         string
		trie         *tTrie
		wantNodes    int
		wantPatterns int
	}{
		/* */
		{
			name:         "01 - nil trie",
			trie:         nil,
			wantNodes:    0,
			wantPatterns: 0,
		},
		{
			name:         "02 - nil root",
			trie:         &tTrie{},
			wantNodes:    0,
			wantPatterns: 0,
		},
		{
			name:         "03 - empty trie",
			trie:         newTrie(),
			wantNodes:    0,
			wantPatterns: 0,
		},
		{
			name: "04 - one pattern",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				return t
			}(),
			wantNodes:    1,
			wantPatterns: 1,
		},
		{
			name: "05 - two patterns",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain"})
				return t
			}(),
			wantNodes:    2,
			wantPatterns: 2,
		},
		{
			name: "06 - three patterns",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "sub"})
				return t
			}(),
			wantNodes:    3,
			wantPatterns: 3,
		},
		{
			name: "07 - two patterns incl. wildcard",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "sub"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "sub", "*"})
				return t
			}(),
			wantNodes:    4,
			wantPatterns: 2,
		},
		{
			name: "08 - two patterns incl. wildcard and one more",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "sub"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "sub", "*"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "sub", "host"})
				return t
			}(),
			wantNodes:    5,
			wantPatterns: 3,
		},
		{
			name: "09 - two patterns incl. wildcard and one more and one more",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "sub"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "sub", "*"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "sub", "host"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "sub", "host", "grand"})
				return t
			}(),
			wantNodes:    6,
			wantPatterns: 4,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotNodes, gotPatterns := tc.trie.Count(context.TODO())

			if gotNodes != tc.wantNodes {
				t.Errorf("tTrie.Count() Nodes = %d, want %d",
					gotNodes, tc.wantNodes)
			}
			if gotPatterns != tc.wantPatterns {
				t.Errorf("tTrie.Count() Patterns = %d, want %d",
					gotPatterns, tc.wantPatterns)
			}
		})
	}
} // Test_tTrie_Count()

func Test_tTrie_ForEach(t *testing.T) {
	tests := []struct {
		name  string
		trie  *tTrie
		aFunc func(aNode *tNode)
	}{
		/* */
		{
			name:  "01 - nil trie",
			trie:  nil,
			aFunc: func(aNode *tNode) {},
		},
		{
			name:  "02 - nil function",
			trie:  newTrie(),
			aFunc: nil,
		},
		{
			name:  "03 - empty trie",
			trie:  newTrie(),
			aFunc: func(aNode *tNode) {},
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.trie.ForEach(context.TODO(), tc.aFunc)
		})
	}
} // Test_tTrie_ForEach()

func Test_tTrie_loadLocal(t *testing.T) {
	tests := []struct {
		name    string
		trie    *tTrie
		fName   string
		wantErr bool
	}{
		/* */
		{
			name:    "01 - nil trie",
			trie:    nil,
			fName:   filepath.Join(t.TempDir(), "test.txt"),
			wantErr: true,
		},
		{
			name:    "02 - nil root",
			trie:    &tTrie{},
			fName:   filepath.Join(t.TempDir(), "test.txt"),
			wantErr: true,
		},
		{
			name:    "03 - non-existent file",
			trie:    newTrie(),
			fName:   filepath.Join(t.TempDir(), "doesnotexist.txt"),
			wantErr: true,
		},
		{
			name: "04 - empty file",
			trie: newTrie(),
			fName: func() string {
				fName := filepath.Join(t.TempDir(), "empty-04.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("\n\n\n")
				_ = f.Close()
				return fName
			}(),
			wantErr: false,
		},
		{
			name: "05 - valid file",
			trie: newTrie(),
			fName: func() string {
				fName := filepath.Join(t.TempDir(), "valid.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("\n# this file contains only hostnames\nwww.example.com\n")
				_ = f.Close()
				return fName
			}(),
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.trie.loadLocal(context.TODO(), tc.fName)

			if (nil != err) != tc.wantErr {
				t.Errorf("tTrie.loadLocal() error = '%v', wantErr '%v'",
					err, tc.wantErr)
			}
		})
	}
} // Test_tTrie_loadLocal()

func Test_tTrie_loadRemote(t *testing.T) {
	tests := []struct {
		name    string
		trie    *tTrie
		url     string
		fName   string
		wantErr bool
	}{
		/* */
		{
			name:    "01 - nil trie",
			trie:    nil,
			fName:   filepath.Join(t.TempDir(), "test.txt"),
			wantErr: true,
		},
		{
			name:    "02 - nil root",
			trie:    &tTrie{},
			fName:   filepath.Join(t.TempDir(), "test.txt"),
			wantErr: true,
		},
		{
			name:    "03 - non-existent file",
			trie:    newTrie(),
			fName:   filepath.Join(t.TempDir(), "doesnotexist.txt"),
			wantErr: true,
		},
		{
			name: "04 - empty file",
			trie: newTrie(),
			url:  "http://example.com/empty.txt",
			fName: func() string {
				fName := filepath.Join(t.TempDir(), "empty.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("\n\n\n")
				_ = f.Close()
				return fName
			}(),
			wantErr: true,
		},
		{
			name: "05 - valid file",
			trie: newTrie(),
			url:  "http://example.com/empty.txt",
			fName: func() string {
				fName := filepath.Join(t.TempDir(), "valid.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("\n# this file contains only hostnames\nwww.example.com\n")
				_ = f.Close()
				return fName
			}(),
			wantErr: true,
		},
		/* */
		{
			name:    "06 - valid file",
			trie:    newTrie(),
			url:     "https://adaway.org/hosts.txt",
			fName:   filepath.Join(t.TempDir(), "hosts-06.txt"),
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := tc.trie.loadRemote(context.TODO(), tc.url, tc.fName)

			if (nil != gotErr) != tc.wantErr {
				t.Errorf("tTrie.loadRemote() error = '%v', wantErr '%v'",
					gotErr, tc.wantErr)
			}
		})
	}
} // Test_tTrie_loadRemote()

func Test_tTrie_Match(t *testing.T) {
	tests := []struct {
		name     string
		trie     *tTrie
		pattern  string
		wantBool bool
	}{
		/* */
		{
			name:     "01 - nil trie",
			trie:     nil,
			pattern:  "tld",
			wantBool: false,
		},
		{
			name:     "02 - nil root",
			trie:     &tTrie{},
			pattern:  "tld",
			wantBool: false,
		},
		{
			name:     "03 - empty pattern",
			trie:     newTrie(),
			pattern:  "",
			wantBool: false,
		},
		{
			name:     "04 - non-matching pattern",
			trie:     newTrie(),
			pattern:  "nothing.will.be.matched.in.an.empty.tree",
			wantBool: false,
		},
		{
			name: "05 - match tld",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				return t
			}(),
			pattern:  "tld",
			wantBool: true,
		},
		/* */
		// More tests are done on the node's method.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotBool := tc.trie.Match(context.TODO(), tc.pattern)

			if gotBool != tc.wantBool {
				t.Errorf("tTrie.Match() gotBool = '%v', want '%v'",
					gotBool, tc.wantBool)
			}
		})
	}
} // Test_tTrie_Match()

func Test_tTrie_Merge(t *testing.T) {
	tests := []struct {
		name     string
		trie     *tTrie
		other    *tTrie
		wantBool bool
	}{
		/* */
		{
			name:     "01 - nil trie",
			trie:     nil,
			other:    newTrie(),
			wantBool: false,
		},
		{
			name:     "02 - nil other",
			trie:     newTrie(),
			other:    nil,
			wantBool: false,
		},
		{
			name:     "03 - nil trie and other",
			trie:     nil,
			other:    nil,
			wantBool: false,
		},
		{
			name:     "04 - empty tries",
			trie:     newTrie(),
			other:    newTrie(),
			wantBool: true,
		},
		{
			name: "05 - merge into empty trie",
			trie: newTrie(),
			other: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				return t
			}(),
			wantBool: true,
		},
		{
			name: "06 - merge into non-empty trie",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				return t
			}(),
			other: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld2"})
				return t
			}(),
			wantBool: true,
		},
		{
			name: "07 - merge multi-level nodes",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "*"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "host"})
				return t
			}(),
			other: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "host"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "sub", "host"})
				return t
			}(),
			wantBool: true,
		},
		/* */
		// More tests are done on the node's method.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotBool := tc.trie.Merge(context.TODO(), tc.other)

			if gotBool != tc.wantBool {
				t.Errorf("tTrie.Merge() gotBool = '%v', want '%v'",
					gotBool, tc.wantBool)
			}
		})
	}
} // Test_tTrie_Merge()

func Test_tTrie_Metrics(t *testing.T) {
	// np, _ := nodepool.Init(func() any {
	// 	return &tNode{tChildren: make(tChildren)}
	// }, 0)
	tests := []struct {
		name    string
		prepare func()
		trie    *tTrie
		want    *TMetrics
	}{
		/* */
		{
			name: "01 - nil trie",
			trie: nil,
			want: nil,
		},
		{
			name: "02 - nil root",
			trie: func() *tTrie {
				return &tTrie{}
			}(),
			want: nil,
		},
		/* * /
		{
			name: "03 - initialised trie",
			prepare: func() {
				np.Clear()
			},
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				return t
			}(),
			want: &TMetrics{
				PoolCreations:  1,
				PoolReturns:    0,
				PoolSize:       0,
				Nodes:          1,
				Patterns:       1,
				Hits:           0,
				Misses:         0,
				Reloads:        0,
				Retries:        0,
				HeapAllocs:     0,
				HeapFrees:      0,
				GCPauseTotalNs: 0,
			},
		},
		{
			name: "04 - initialised trie with no patterns",
			prepare: func() {
				np.Clear()
			},
			trie: func() *tTrie {
				return newTrie()
			}(),
			want: &TMetrics{
				PoolCreations:  1,
				PoolReturns:    0,
				PoolSize:       0,
				Nodes:          0,
				Patterns:       0,
				Hits:           0,
				Misses:         0,
				Reloads:        0,
				Retries:        0,
				HeapAllocs:     0,
				HeapFrees:      0,
				GCPauseTotalNs: 0,
			},
		},
		{
			name: "05 - initialised trie with patterns",
			prepare: func() {
				np.Clear()
			},
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "sub"})
				return t
			}(),
			want: &TMetrics{
				PoolCreations:  3,
				PoolReturns:    0,
				PoolSize:       0,
				Nodes:          3,
				Patterns:       3,
				Hits:           0,
				Misses:         0,
				Reloads:        0,
				Retries:        0,
				HeapAllocs:     0,
				HeapFrees:      0,
				GCPauseTotalNs: 0,
			},
		},
		/* */
		// More tests are done on the metric's method.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if nil != tc.prepare {
				tc.prepare()
			}
			got := tc.trie.Metrics()

			if nil == got {
				if nil != tc.want {
					t.Error("tTrie.Metrics() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("tTrie.Metrics() =\n%q\nwant 'nil'",
					got.String())
				return
			}
			if !tc.want.Equal(got) {
				t.Errorf("tTrie.Metrics() =\n%q\nwant\n%q",
					got.String(), tc.want.String())
			}
		})
	}
} // Test_tTrie_Metrics()

func Test_tTrie_storeFile(t *testing.T) {
	tests := []struct {
		name      string
		trie      *tTrie
		aFilename string
		wantErr   bool
	}{
		/* */
		{
			name:      "01 - nil trie",
			trie:      nil,
			aFilename: filepath.Join(t.TempDir(), "test-01.txt"),
			wantErr:   true,
		},
		{
			name:      "02 - nil root",
			trie:      &tTrie{},
			aFilename: filepath.Join(t.TempDir(), "test-02.txt"),
			wantErr:   true,
		},
		{
			name:      "03 - empty filename",
			trie:      newTrie(),
			aFilename: "",
			wantErr:   true,
		},
		{
			name:      "04 - valid filename",
			trie:      newTrie(),
			aFilename: filepath.Join(t.TempDir(), "test-04.txt"),
			wantErr:   false,
		},
		{
			name: "05 - valid filename, with patterns",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "host"})
				return t
			}(),
			aFilename: filepath.Join(t.TempDir(), "test-05.txt"),
			wantErr:   false,
		},
		{
			name: "06 - valid filename, with multiple patterns",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				t.root.node.add(context.TODO(), tPartsList{"tld2", "domain"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "host"})
				t.root.node.add(context.TODO(), tPartsList{"tld2"})
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain"})
				t.root.node.add(context.TODO(), tPartsList{"tld2", "domain", "host"})
				return t
			}(),
			aFilename: filepath.Join(t.TempDir(), "test-06.txt"),
			wantErr:   false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := tc.trie.storeFile(context.TODO(), tc.aFilename)

			if (nil != gotErr) != tc.wantErr {
				t.Errorf("tTrie.storeFile() error = '%v', wantErr '%v'",
					gotErr, tc.wantErr)
			}
		})
	}
} // Test_tTrie_storeFile()

func Test_tTrie_String(t *testing.T) {
	tests := []struct {
		name string
		trie *tTrie
		want string
	}{
		/* */
		{
			name: "01 - nil trie",
			trie: nil,
			want: "node or argument is nil",
		},
		{
			name: "02 - empty trie",
			trie: newTrie(),
			want: "\"Trie\":\n  isEnd: false\n  isWild: false\n",
		},
		{
			name: "03 - trie with root",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				return t
			}(),
			want: "\"Trie\":\n  isEnd: false\n  isWild: false\n  \"tld\":\n      isEnd: true\n      isWild: false\n",
		},
		{
			name: "04 - trie with root and children",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain"})
				return t
			}(),
			want: "\"Trie\":\n  isEnd: false\n  isWild: false\n  \"tld\":\n      isEnd: false\n      isWild: false\n      \"domain\":\n          isEnd: true\n          isWild: false\n",
		},
		{
			name: "05 - trie with root, child and wildcard",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"*", "domain"})
				return t
			}(),
			want: "\"Trie\":\n  isEnd: false\n  isWild: false\n  \"*\":\n      isEnd: false\n      isWild: true\n      \"domain\":\n          isEnd: true\n          isWild: false\n",
		},
		{
			name: "06 - trie with root and children and wildcard",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "*"})
				return t
			}(),
			want: "\"Trie\":\n  isEnd: false\n  isWild: false\n  \"tld\":\n      isEnd: false\n      isWild: false\n      \"domain\":\n          isEnd: false\n          isWild: false\n          \"*\":\n              isEnd: false\n              isWild: true\n",
		},
		{
			name: "07 - trie with root and child and wildcard and child",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld", "domain", "*", "sub"})
				return t
			}(),
			want: "\"Trie\":\n  isEnd: false\n  isWild: false\n  \"tld\":\n      isEnd: false\n      isWild: false\n      \"domain\":\n          isEnd: false\n          isWild: false\n          \"*\":\n              isEnd: false\n              isWild: true\n              \"sub\":\n                  isEnd: true\n                  isWild: false\n",
		},
		/* */
		// More tests are done on the node's method.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.trie.String(); got != tc.want {
				t.Errorf("tTrie.String() =\n%q\n\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tTrie_String()

func Test_tTrie_Update(t *testing.T) {
	tests := []struct {
		name       string
		trie       *tTrie
		oldPattern string
		newPattern string
		wantBool   bool
	}{
		/* */
		{
			name:       "01 - nil trie",
			trie:       nil,
			oldPattern: "tld",
			newPattern: "tld",
			wantBool:   false,
		},
		{
			name:       "02 - nil root",
			trie:       &tTrie{},
			oldPattern: "tld",
			newPattern: "tld",
			wantBool:   false,
		},
		{
			name:       "03 - empty old pattern",
			trie:       newTrie(),
			oldPattern: "",
			newPattern: "tld",
			wantBool:   false,
		},
		{
			name:       "04 - empty new pattern",
			trie:       newTrie(),
			oldPattern: "tld",
			newPattern: "",
			wantBool:   false,
		},
		{
			name:       "05 - equal old and new pattern",
			trie:       newTrie(),
			oldPattern: "tld",
			newPattern: "tld",
			wantBool:   false,
		},
		{
			name: "06 - update tld",
			trie: func() *tTrie {
				t := newTrie()
				t.root.node.add(context.TODO(), tPartsList{"tld"})
				return t
			}(),
			oldPattern: "tld",
			newPattern: "new.tld",
			wantBool:   true,
		},
		/* */
		// More tests are done with the node's method.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotBool := tc.trie.Update(context.TODO(), tc.oldPattern, tc.newPattern)

			if gotBool != tc.wantBool {
				t.Errorf("tTrie.Update() gotBool = '%v', want '%v'",
					gotBool, tc.wantBool)
			}
		})
	}
} // Test_tTrie_Update()

/* _EoF_ */
