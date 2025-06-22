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

func Test_New(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want *TADlist
	}{
		{
			name: "01 - new list",
			dir:  t.TempDir(),
			want: &TADlist{
				allow: newTrie(),
				deny:  newTrie(),
			},
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := New(tc.dir)

			if nil == got {
				t.Error("New() = nil, want non-nil")
				return
			}
			if nil == tc.want {
				t.Errorf("New() =\n%q\nwant 'nil'",
					got)
				return
			}
			if !tc.want.Equal(got) {
				t.Errorf("New() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_New()

func Test_TADlist_AddAllow(t *testing.T) {
	tests := []struct {
		name    string
		adl     *TADlist
		pattern string
		wantOK  bool
	}{
		/* */
		{
			name:    "01 - nil list",
			adl:     nil,
			pattern: "tld",
			wantOK:  false,
		},
		{
			name:    "02 - empty pattern",
			adl:     New(t.TempDir()),
			pattern: "",
			wantOK:  false,
		},
		{
			name:    "03 - add tld",
			adl:     New(t.TempDir()),
			pattern: "tld",
			wantOK:  true,
		},
		/* */
		// More tests are done with the trie's method.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := tc.adl.AddAllow(context.TODO(), tc.pattern)

			if gotOK != tc.wantOK {
				t.Errorf("TADlist.AddAllow() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_TADlist_AddAllow()

func Test_TADlist_AddDeny(t *testing.T) {
	tests := []struct {
		name    string
		adl     *TADlist
		pattern string
		wantOK  bool
	}{
		/* */
		{
			name:    "01 - nil list",
			adl:     nil,
			pattern: "tld",
			wantOK:  false,
		},
		{
			name:    "02 - empty pattern",
			adl:     New(t.TempDir()),
			pattern: "",
			wantOK:  false,
		},
		{
			name:    "03 - add tld",
			adl:     New(t.TempDir()),
			pattern: "tld",
			wantOK:  true,
		},
		/* */
		// More tests are done with the trie's method.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := tc.adl.AddDeny(context.TODO(), tc.pattern)

			if gotOK != tc.wantOK {
				t.Errorf("TADlist.AddDeny() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_TADlist_AddDeny()

func Test_TADlist_DeleteAllow(t *testing.T) {
	tests := []struct {
		name    string
		adl     *TADlist
		pattern string
		wantOK  bool
	}{
		/* */
		{
			name:    "01 - nil list",
			adl:     nil,
			pattern: "tld",
			wantOK:  false,
		},
		{
			name:    "02 - empty pattern",
			adl:     New(t.TempDir()),
			pattern: "",
			wantOK:  false,
		},
		{
			name:    "03 - delete missing tld",
			adl:     New(t.TempDir()),
			pattern: "tld",
			wantOK:  false,
		},
		{
			name: "04 - delete added tld",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddAllow(context.TODO(), "tld")
				return a
			}(),
			pattern: "tld",
			wantOK:  true,
		},
		{
			name: "05 - delete added domain.tld",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddAllow(context.TODO(), "domain.tld")
				return a
			}(),
			pattern: "domain.tld",
			wantOK:  true,
		},
		{
			name: "06 - delete added sub.domain.tld",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddAllow(context.TODO(), "sub.domain.tld")
				return a
			}(),
			pattern: "sub.domain.tld",
			wantOK:  true,
		},
		/* */
		// More tests are done with the trie's method.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := tc.adl.DeleteAllow(context.TODO(), tc.pattern)

			if gotOK != tc.wantOK {
				t.Errorf("TADlist.DeleteAllow() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_TADlist_DeleteAllow()

func Test_TADlist_DeleteDeny(t *testing.T) {
	tests := []struct {
		name    string
		adl     *TADlist
		pattern string
		wantOK  bool
	}{
		/* */
		{
			name:    "01 - nil list",
			adl:     nil,
			pattern: "tld",
			wantOK:  false,
		},
		{
			name:    "02 - empty pattern",
			adl:     New(t.TempDir()),
			pattern: "",
			wantOK:  false,
		},
		{
			name:    "03 - delete missing tld",
			adl:     New(t.TempDir()),
			pattern: "tld",
			wantOK:  false,
		},
		{
			name: "04 - delete added tld",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddDeny(context.TODO(), "tld")
				return a
			}(),
			pattern: "tld",
			wantOK:  true,
		},
		{
			name: "05 - delete added domain.tld",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddDeny(context.TODO(), "domain.tld")
				return a
			}(),
			pattern: "domain.tld",
			wantOK:  true,
		},
		{
			name: "06 - delete added sub.domain.tld",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddDeny(context.TODO(), "sub.domain.tld")
				return a
			}(),
			pattern: "sub.domain.tld",
			wantOK:  true,
		},
		/* */
		// More tests are done with the trie's method.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := tc.adl.DeleteDeny(context.TODO(), tc.pattern)

			if gotOK != tc.wantOK {
				t.Errorf("TADlist.DeleteDeny() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_TADlist_DeleteDeny()

func Test_TADlist_Equal(t *testing.T) {
	adlist := New(t.TempDir())
	tests := []struct {
		name   string
		adl    *TADlist
		other  *TADlist
		wantOK bool
	}{
		/* */
		{
			name:   "01 - nil list",
			adl:    nil,
			other:  nil,
			wantOK: true,
		},
		{
			name:   "02 - nil list and non-nil list",
			adl:    nil,
			other:  adlist,
			wantOK: false,
		},
		{
			name:   "03 - non-nil list and nil list",
			adl:    adlist,
			other:  nil,
			wantOK: false,
		},
		{
			name:   "04 - equal lists",
			adl:    adlist,
			other:  New(t.TempDir()),
			wantOK: true,
		},
		{
			name:   "05 - same lists",
			adl:    adlist,
			other:  adlist,
			wantOK: true,
		},
		{
			name: "05 - equal lists with allow patterns",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddAllow(context.TODO(), "tld")
				a.AddAllow(context.TODO(), "domain.tld")
				a.AddAllow(context.TODO(), "sub.domain.tld")
				return a
			}(),
			other: func() *TADlist {
				a := New(t.TempDir())
				a.AddAllow(context.TODO(), "tld")
				a.AddAllow(context.TODO(), "domain.tld")
				a.AddAllow(context.TODO(), "sub.domain.tld")
				return a
			}(),
			wantOK: true,
		},
		{
			name: "06 - equal lists with deny patterns",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddDeny(context.TODO(), "tld")
				a.AddDeny(context.TODO(), "domain.tld")
				a.AddDeny(context.TODO(), "sub.domain.tld")
				return a
			}(),
			other: func() *TADlist {
				a := New(t.TempDir())
				a.AddDeny(context.TODO(), "tld")
				a.AddDeny(context.TODO(), "domain.tld")
				a.AddDeny(context.TODO(), "sub.domain.tld")
				return a
			}(),
			wantOK: true,
		},
		{
			name: "07 - lists with different allow and deny patterns",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddAllow(context.TODO(), "tld")
				a.AddAllow(context.TODO(), "domain.tld")
				a.AddAllow(context.TODO(), "sub.domain.tld")
				a.AddDeny(context.TODO(), "tld")
				a.AddDeny(context.TODO(), "domain.tld")
				a.AddDeny(context.TODO(), "sub.domain.tld")
				return a
			}(),
			other: func() *TADlist {
				a := New(t.TempDir())
				a.AddAllow(context.TODO(), "domain.tld")
				a.AddAllow(context.TODO(), "sub.domain.tld")
				a.AddDeny(context.TODO(), "domain.tld")
				a.AddDeny(context.TODO(), "sub.domain.tld")
				return a
			}(),
			wantOK: false,
		},
		/* */
		// TODO: Add test cases.
		// More tests are done with the trie's method.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := tc.adl.Equal(tc.other)

			if gotOK != tc.wantOK {
				t.Errorf("TADlist.Equal() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_TADlist_Equal()

func Test_TADlist_LoadAllow(t *testing.T) {
	tests := []struct {
		name    string
		adl     *TADlist
		fName   string
		wantErr bool
	}{
		/* */
		{
			name:    "01 - nil list",
			adl:     nil,
			fName:   "allow-01.txt",
			wantErr: true,
		},
		{
			name:    "02 - non-existent file",
			adl:     New(t.TempDir()),
			fName:   "doesnotexist-02.txt",
			wantErr: true,
		},
		{
			name: "03 - valid file with empty lines",
			adl:  New(t.TempDir()),
			fName: func() string {
				fName := filepath.Join(t.TempDir(), "allow-03.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("\n\n\n")
				_ = f.Close()
				return fName
			}(),
			wantErr: false,
		},
		{
			name: "04 - valid file with comments",
			adl:  New(t.TempDir()),
			fName: func() string {
				fName := filepath.Join(t.TempDir(), "allow-04.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("\n# this file contains only hostnames\nwww.example.com\nwww.other.com\nwww.another.com\n")
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
			gotErr := tc.adl.LoadAllow(context.TODO(), tc.fName)

			if (nil != gotErr) != tc.wantErr {
				t.Errorf("TADlist.LoadAllow() error = '%v', wantErr '%v'",
					gotErr, tc.wantErr)
			}
		})
	}
} // Test_TADlist_LoadAllow()

func Test_TADlist_LoadDeny(t *testing.T) {
	tests := []struct {
		name    string
		adl     *TADlist
		urls    []string
		wantErr bool
	}{
		/* */
		{
			name:    "01 - nil list",
			adl:     nil,
			urls:    []string{"http://example.com/deny.txt"},
			wantErr: true,
		},
		{
			name:    "02 - non-existent file",
			adl:     New(t.TempDir()),
			urls:    []string{"http://example.com/deny.txt"},
			wantErr: true,
		},
		{
			name:    "03 - valid file with empty lines",
			adl:     New(t.TempDir()),
			urls:    []string{"http://example.com/deny.txt"},
			wantErr: true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := tc.adl.LoadDeny(context.TODO(), tc.urls)

			if (nil != gotErr) != tc.wantErr {
				t.Errorf("TADlist.LoadDeny() error = '%v', wantErr '%v'",
					gotErr, tc.wantErr)
			}
		})
	}
} // Test_TADlist_LoadDeny()

func Test_TADlist_Match(t *testing.T) {
	tests := []struct {
		name     string
		adl      *TADlist
		hostname string
		want     TADresult
	}{
		/* */
		{
			name:     "01 - nil list",
			adl:      nil,
			hostname: "tld",
			want:     ADneutral,
		},
		{
			name:     "02 - empty hostname",
			adl:      New(t.TempDir()),
			hostname: "",
			want:     ADdeny,
		},
		{
			name:     "03 - non-matching hostname",
			adl:      New(t.TempDir()),
			hostname: "nothing.will.be.matched.in.an.empty.tree",
			want:     ADneutral,
		},
		{
			name: "04 - match allow tld",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddAllow(context.TODO(), "tld")
				return a
			}(),
			hostname: "tld",
			want:     ADallow,
		},
		{
			name: "05 - match allow domain.tld",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddAllow(context.TODO(), "domain.tld")
				return a
			}(),
			hostname: "domain.tld",
			want:     ADallow,
		},
		{
			name: "06 - match allow sub.domain.tld",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddAllow(context.TODO(), "sub.domain.tld")
				return a
			}(),
			hostname: "sub.domain.tld",
			want:     ADallow,
		},
		/* */
		// More tests are done with the trie's method.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.adl.Match(context.TODO(), tc.hostname)

			if got != tc.want {
				t.Errorf("TADlist.Match() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_TADlist_Match()

func Test_TADlist_Shutdown(t *testing.T) {
	tests := []struct {
		name    string
		adl     *TADlist
		wantErr bool
	}{
		/* */
		{
			name:    "01 - nil list",
			adl:     nil,
			wantErr: true,
		},
		{
			name:    "02 - empty list",
			adl:     New(t.TempDir()),
			wantErr: true,
		},
		{
			name: "03 - valid names",
			adl: func() *TADlist {
				ad := New(t.TempDir())
				ad.AddAllow(context.TODO(), "host.domain.tld")
				ad.AddAllow(context.TODO(), "www.domain.tld")
				ad.AddDeny(context.TODO(), "sub.domain.tld")
				return ad
			}(),
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := tc.adl.Shutdown()

			if (nil != gotErr) != tc.wantErr {
				t.Errorf("TADlist.Shutdown() error = '%v', wantErr '%v'",
					gotErr, tc.wantErr)
			}
		})
	}
} // Test_TADlist_Shutdown()

func Test_TADlist_StoreAllow(t *testing.T) {
	tests := []struct {
		name    string
		adl     *TADlist
		wantErr bool
	}{
		/* */
		{
			name:    "01 - nil list",
			adl:     nil,
			wantErr: true,
		},
		{
			name:    "02 - empty trie",
			adl:     New(t.TempDir()),
			wantErr: true,
		},
		{
			name: "03 - valid filename",
			adl: func() *TADlist {
				ad := New(t.TempDir())
				ad.AddAllow(context.TODO(), "domain.tld")
				return ad
			}(),
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := tc.adl.StoreAllow(context.TODO())

			if (nil != gotErr) != tc.wantErr {
				t.Errorf("TADlist.StoreAllow() error = '%v', wantErr '%v'",
					gotErr, tc.wantErr)
			}
		})
	}
} // Test_TADlist_StoreAllow()

func Test_TADlist_StoreDeny(t *testing.T) {
	tests := []struct {
		name    string
		adl     *TADlist
		wantErr bool
	}{
		/* */
		{
			name:    "01 - nil list",
			adl:     nil,
			wantErr: true,
		},
		{
			name:    "02 - empty List",
			adl:     New(t.TempDir()),
			wantErr: true,
		},
		{
			name:    "03 - valid filename",
			adl:     New(t.TempDir()),
			wantErr: true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := tc.adl.StoreDeny(context.TODO())

			if (nil != gotErr) != tc.wantErr {
				t.Errorf("TADlist.StoreDeny() error = '%v', wantErr '%v'",
					gotErr, tc.wantErr)
			}
		})
	}
} // Test_TADlist_StoreDeny()

func Test_TADlist_String(t *testing.T) {
	tests := []struct {
		name string
		adl  *TADlist
		want string
	}{
		/* */
		{
			name: "01 - nil list",
			adl:  nil,
			want: "",
		},
		{
			name: "02 - empty list",
			adl:  New(t.TempDir()),
			want: "Allow:\n\"Trie\":\n  isEnd: false\n  isWild: false\n\nDeny:\n\"Trie\":\n  isEnd: false\n  isWild: false\n",
		},
		{
			name: "03 - one allow pattern",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddAllow(context.TODO(), "tld")
				return a
			}(),
			want: "Allow:\n\"Trie\":\n  isEnd: false\n  isWild: false\n  \"tld\":\n      isEnd: true\n      isWild: false\n\nDeny:\n\"Trie\":\n  isEnd: false\n  isWild: false\n",
		},
		{
			name: "04 - one deny pattern",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddDeny(context.TODO(), "tld")
				return a
			}(),
			want: "Allow:\n\"Trie\":\n  isEnd: false\n  isWild: false\n\nDeny:\n\"Trie\":\n  isEnd: false\n  isWild: false\n  \"tld\":\n      isEnd: true\n      isWild: false\n",
		},
		{
			name: "05 - two allow patterns",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddAllow(context.TODO(), "tld")
				a.AddAllow(context.TODO(), "domain.tld")
				return a
			}(),
			want: "Allow:\n\"Trie\":\n  isEnd: false\n  isWild: false\n  \"tld\":\n      isEnd: true\n      isWild: false\n      \"domain\":\n          isEnd: true\n          isWild: false\n\nDeny:\n\"Trie\":\n  isEnd: false\n  isWild: false\n",
		},
		{
			name: "06 - two deny patterns",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddDeny(context.TODO(), "tld")
				a.AddDeny(context.TODO(), "domain.tld")
				return a
			}(),
			want: "Allow:\n\"Trie\":\n  isEnd: false\n  isWild: false\n\nDeny:\n\"Trie\":\n  isEnd: false\n  isWild: false\n  \"tld\":\n      isEnd: true\n      isWild: false\n      \"domain\":\n          isEnd: true\n          isWild: false\n",
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.adl.String()

			if got != tc.want {
				t.Errorf("TADlist.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_TADlist_String()

func Test_TADlist_UpdateAllow(t *testing.T) {
	tests := []struct {
		name       string
		adl        *TADlist
		oldPattern string
		newPattern string
		wantOK     bool
	}{
		/* */
		{
			name:       "01 - nil list",
			adl:        nil,
			oldPattern: "tld",
			newPattern: "tld",
			wantOK:     false,
		},
		{
			name:       "02 - empty old pattern",
			adl:        New(t.TempDir()),
			oldPattern: "",
			newPattern: "tld",
			wantOK:     false,
		},
		{
			name:       "03 - empty new pattern",
			adl:        New(t.TempDir()),
			oldPattern: "tld",
			newPattern: "",
			wantOK:     false,
		},
		{
			name:       "04 - equal old and new pattern",
			adl:        New(t.TempDir()),
			oldPattern: "tld",
			newPattern: "tld",
			wantOK:     false,
		},
		{
			name: "05 - update non-existent old pattern",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddAllow(context.TODO(), "tld")
				return a
			}(),
			oldPattern: "domain.tld",
			newPattern: "sub.domain.tld",
			wantOK:     true,
		},
		{
			name: "06 - update existing old pattern",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddAllow(context.TODO(), "tld")
				return a
			}(),
			oldPattern: "tld",
			newPattern: "domain.tld",
			wantOK:     true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := tc.adl.UpdateAllow(context.TODO(), tc.oldPattern, tc.newPattern)

			if gotOK != tc.wantOK {
				t.Errorf("TADlist.UpdateAllow() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_TADlist_UpdateAllow()

func Test_TADlist_UpdateDeny(t *testing.T) {
	tests := []struct {
		name       string
		adl        *TADlist
		oldPattern string
		newPattern string
		wantOK     bool
	}{
		/* */
		{
			name:       "01 - nil list",
			adl:        nil,
			oldPattern: "tld",
			newPattern: "tld",
			wantOK:     false,
		},
		{
			name:       "02 - empty old pattern",
			adl:        New(t.TempDir()),
			oldPattern: "",
			newPattern: "tld",
			wantOK:     false,
		},
		{
			name:       "03 - empty new pattern",
			adl:        New(t.TempDir()),
			oldPattern: "tld",
			newPattern: "",
			wantOK:     false,
		},
		{
			name:       "04 - equal old and new pattern",
			adl:        New(t.TempDir()),
			oldPattern: "tld",
			newPattern: "tld",
			wantOK:     false,
		},
		{
			name: "05 - update non-existent old pattern",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddDeny(context.TODO(), "tld")
				return a
			}(),
			oldPattern: "domain.tld",
			newPattern: "sub.domain.tld",
			wantOK:     true,
		},
		{
			name: "06 - update existing old pattern",
			adl: func() *TADlist {
				a := New(t.TempDir())
				a.AddDeny(context.TODO(), "tld")
				return a
			}(),
			oldPattern: "tld",
			newPattern: "domain.tld",
			wantOK:     true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := tc.adl.UpdateDeny(context.TODO(), tc.oldPattern, tc.newPattern)

			if gotOK != tc.wantOK {
				t.Errorf("TADlist.UpdateDeny() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_TADlist_UpdateDeny()

func Test_urlPath2Filename(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		/* */
		{
			name:    "01 - empty url",
			url:     "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "02 - invalid url",
			url:     "://example.com/",
			want:    "",
			wantErr: true,
		},
		{
			name:    "03 - incomplete url, no path",
			url:     "http://example.com",
			want:    "",
			wantErr: true,
		},
		{
			name:    "04 - valid url, no path",
			url:     "http://example.com/",
			want:    "",
			wantErr: true,
		},
		{
			name:    "05 - valid url, with path",
			url:     "http://example.com/path/to/file",
			want:    "path_to_file",
			wantErr: false,
		},
		/* */
		{
			name:    "06 - valid url, with path and query",
			url:     "http://example.com/path/to/file.txt?query=string",
			want:    "path_to_file.txt",
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, gotErr := urlPath2Filename(tc.url)

			if (gotErr != nil) != tc.wantErr {
				t.Errorf("urlPath2Filename() error = '%v', wantErr '%v'",
					gotErr, tc.wantErr)
				return
			}
			if got != tc.want {
				t.Errorf("urlPath2Filename() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_urlPath2Filename()

/* _EoF_ */
