/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_New(t *testing.T) {
	tests := []struct {
		name  string
		cType TCacheType
		cSize uint
		want  ICacheList
	}{
		{
			name:  "01 - nil",
			cType: CacheTypeMap,
			cSize: 0,
			want:  &tMapList{},
		},
		{
			name:  "02 - empty",
			cType: CacheTypeMap,
			cSize: 0,
			want:  &tMapList{},
		},
		{
			name:  "03 - default size",
			cType: CacheTypeMap,
			cSize: DefaultCacheSize,
			want:  &tMapList{},
		},
		{
			name:  "04 - custom size",
			cType: CacheTypeMap,
			cSize: 128,
			want:  &tMapList{},
		},
		{
			name:  "05 - trie",
			cType: CacheTypeTrie,
			cSize: 0,
			want: &tTrieList{
				tRoot: tRoot{
					node: newTrieNode(),
				},
			},
		},
		{
			name:  "06 - trie with size",
			cType: CacheTypeTrie,
			cSize: 128,
			want: &tTrieList{
				tRoot: tRoot{
					node: newTrieNode(),
				},
			},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := New(tc.cType, tc.cSize)

			if nil == got {
				if nil != tc.want {
					t.Error("New() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("New() =\n%v\nwant 'nil'",
					got)
				return
			}

			gotMap, Gok := got.(*tMapList)
			wantMap, Wok := tc.want.(*tMapList)
			if Gok && Wok {
				if !wantMap.Equal(gotMap) {
					t.Errorf("New() =\n%q\nwant\n%q",
						gotMap, wantMap)
				}
				return
			}

			gotTrie, Gok := got.(*tTrieList)
			wantTrie, Wok := tc.want.(*tTrieList)
			if Gok && Wok {
				if !wantTrie.Equal(gotTrie) {
					t.Errorf("New() =\n%q\nwant\n%q",
						gotTrie, wantTrie)
				}
				return
			}
		})
	}
} // Test_New()

/* _EoF_ */
