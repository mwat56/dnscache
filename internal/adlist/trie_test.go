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

func Test_tTrie_Equal(t *testing.T) {
	tc1 := newTrie()
	tc2 := newTrie()
	tc2.root.add(tPartsList{"tld"})

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
			gotBool := tc.trie.Add(tc.pattern)

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
				t.root.add(tPartsList{"tld"})
				return t
			}(),
			wantList: tPartsList{"tld"},
		},
		{
			name: "05 - two patterns",
			trie: func() *tTrie {
				t := newTrie()
				t.root.add(tPartsList{"tld"})
				t.root.add(tPartsList{"tld", "domain"})
				return t
			}(),
			wantList: tPartsList{"tld", "domain.tld"},
		},
		{
			name: "06 - three patterns",
			trie: func() *tTrie {
				t := newTrie()
				t.root.add(tPartsList{"tld"})
				t.root.add(tPartsList{"tld", "domain"})
				t.root.add(tPartsList{"tld", "domain", "sub"})
				return t
			}(),
			wantList: tPartsList{"tld", "domain.tld", "sub.domain.tld"},
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotList := tc.trie.AllPatterns()
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
				t.root.add(tPartsList{"tld"})
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
				t.root.add(tPartsList{"tld", "domain"})
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
				t.root.add(tPartsList{"tld", "domain", "host"})
				t.root.add(tPartsList{"tld", "domain", "*"})
				return t
			}(),
			pattern:  "host.domain.tld",
			wantBool: true,
		},
		{
			name: "07 - delete *.domain.tld",
			trie: func() *tTrie {
				t := newTrie()
				t.root.add(tPartsList{"tld", "domain", "host"})
				t.root.add(tPartsList{"tld", "domain", "*"})
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
			gotBool := tc.trie.Delete(tc.pattern)
			if gotBool != tc.wantBool {
				t.Errorf("tTrie.Delete() gotBool = '%v', want '%v'",
					gotBool, tc.wantBool)
			}
		})
	}
} // Test_tTrie_Delete()

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
			tc.trie.ForEach(tc.aFunc)
		})
	}
} // Test_tTrie_ForEach()

func Test_tTrie_Load(t *testing.T) {
	tests := []struct {
		name    string
		trie    *tTrie
		reader  io.Reader
		wantErr bool
	}{
		/* */
		{
			name:    "01 - nil trie",
			trie:    nil,
			reader:  strings.NewReader("tld"),
			wantErr: true,
		},
		{
			name:    "02 - nil reader",
			trie:    newTrie(),
			reader:  nil,
			wantErr: true,
		},
		{
			name:    "03 - empty reader",
			trie:    newTrie(),
			reader:  strings.NewReader(""),
			wantErr: false,
		},
		{
			name:    "04 - reader with comments",
			trie:    newTrie(),
			reader:  strings.NewReader("# comment\n; comment\n# the next line is no comment\n comment"),
			wantErr: false,
		},
		{
			name:    "05 - reader with empty lines",
			trie:    newTrie(),
			reader:  strings.NewReader("\n\n\n"),
			wantErr: false,
		},
		{
			name:    "06 - reader with valid data",
			trie:    newTrie(),
			reader:  strings.NewReader("tld\ndomain.tld\nhost.domain.tld\ninvalid\n*.domain.tld"),
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.trie.Load(tc.reader)
			if (nil != err) != tc.wantErr {
				t.Errorf("tTrie.Load() error = '%v', wantErr '%v'",
					err, tc.wantErr)
				return
			}
		})
	}
} // Test_tTrie_Load()

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
				t.root.add(tPartsList{"tld"})
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
			gotBool := tc.trie.Match(tc.pattern)
			if gotBool != tc.wantBool {
				t.Errorf("tTrie.Match() gotBool = '%v', want '%v'",
					gotBool, tc.wantBool)
			}
		})
	}
} // Test_tTrie_Match()

func Test_tTrie_Store(t *testing.T) {
	tests := []struct {
		name     string
		trie     *tTrie
		wantText string
		wantErr  bool
	}{
		/* */
		{
			name:     "01 - nil trie",
			trie:     nil,
			wantText: "",
			wantErr:  true,
		},
		{
			name:     "02 - nil root node",
			trie:     &tTrie{},
			wantText: "",
			wantErr:  true,
		},
		{
			name:     "03 - empty trie",
			trie:     newTrie(),
			wantText: "",
			wantErr:  false,
		},
		{
			name: "04 - one pattern",
			trie: func() *tTrie {
				t := newTrie()
				t.root.add(tPartsList{"tld"})
				return t
			}(),
			wantText: "tld\n",
			wantErr:  false,
		},
		/* */
		// More tests are done on the node's method.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			aWriter := &bytes.Buffer{}
			err := tc.trie.Store(aWriter)

			if (nil != err) != tc.wantErr {
				t.Errorf("tTrie.Store() error = '%v', wantErr '%v'",
					err, tc.wantErr)
				return
			}
			if gotText := aWriter.String(); gotText != tc.wantText {
				t.Errorf("tTrie.Store() =\n%q\nwant\n%q",
					gotText, tc.wantText)
			}
		})
	}
} // Test_tTrie_Store()

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
			want: "\"Trie\":\n  isEnd: false\n  isWild: false\n  hits: 0\n",
		},
		{
			name: "03 - trie with root",
			trie: func() *tTrie {
				t := newTrie()
				t.root.add(tPartsList{"tld"})
				return t
			}(),
			want: "\"Trie\":\n  isEnd: false\n  isWild: false\n  hits: 0\n  \"tld\":\n      isEnd: true\n      isWild: false\n      hits: 0\n",
		},
		{
			name: "04 - trie with root and children",
			trie: func() *tTrie {
				t := newTrie()
				t.root.add(tPartsList{"tld", "domain"})
				return t
			}(),
			want: "\"Trie\":\n  isEnd: false\n  isWild: false\n  hits: 0\n  \"tld\":\n      isEnd: false\n      isWild: false\n      hits: 0\n      \"domain\":\n          isEnd: true\n          isWild: false\n          hits: 0\n",
		},
		{
			name: "05 - trie with root, child and wildcard",
			trie: func() *tTrie {
				t := newTrie()
				t.root.add(tPartsList{"*", "domain"})
				return t
			}(),
			want: "\"Trie\":\n  isEnd: false\n  isWild: false\n  hits: 0\n  \"*\":\n      isEnd: false\n      isWild: true\n      hits: 0\n      \"domain\":\n          isEnd: true\n          isWild: false\n          hits: 0\n",
		},
		{
			name: "06 - trie with root and children and wildcard",
			trie: func() *tTrie {
				t := newTrie()
				t.root.add(tPartsList{"tld", "domain", "*"})
				return t
			}(),
			want: "\"Trie\":\n  isEnd: false\n  isWild: false\n  hits: 0\n  \"tld\":\n      isEnd: false\n      isWild: false\n      hits: 0\n      \"domain\":\n          isEnd: false\n          isWild: false\n          hits: 0\n          \"*\":\n              isEnd: false\n              isWild: true\n              hits: 0\n",
		},
		{
			name: "07 - trie with root and child and wildcard and child",
			trie: func() *tTrie {
				t := newTrie()
				t.root.add(tPartsList{"tld", "domain", "*", "sub"})
				return t
			}(),
			want: "\"Trie\":\n  isEnd: false\n  isWild: false\n  hits: 0\n  \"tld\":\n      isEnd: false\n      isWild: false\n      hits: 0\n      \"domain\":\n          isEnd: false\n          isWild: false\n          hits: 0\n          \"*\":\n              isEnd: false\n              isWild: true\n              hits: 0\n              \"sub\":\n                  isEnd: true\n                  isWild: false\n                  hits: 0\n",
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
				t.root.add(tPartsList{"tld"})
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
			gotBool := tc.trie.Update(tc.oldPattern, tc.newPattern)

			if gotBool != tc.wantBool {
				t.Errorf("tTrie.Update() gotBool = '%v', want '%v'",
					gotBool, tc.wantBool)
			}
		})
	}
} // Test_tTrie_Update()

/* _EoF_ */
