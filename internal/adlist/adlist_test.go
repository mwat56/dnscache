/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"path/filepath"
	"reflect"
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_NewADlist(t *testing.T) {
	tests := []struct {
		name string
		want *TADlist
	}{
		{
			name: "01 - new list",
			want: &TADlist{
				allow: newTrie(),
				deny:  newTrie(),
			},
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NewADlist()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("NewADlist() = %v, want %v", got, tc.want)
			}
		})
	}
} // Test_NewADlist()

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
			adl:     NewADlist(),
			pattern: "",
			wantOK:  false,
		},
		{
			name:    "03 - add tld",
			adl:     NewADlist(),
			pattern: "tld",
			wantOK:  true,
		},
		/* */
		// More tests are done with the trie's method.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotROK := tc.adl.AddAllow(tc.pattern)

			if gotROK != tc.wantOK {
				t.Errorf("TADlist.AddAllow() = '%v', want '%v'",
					gotROK, tc.wantOK)
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
			adl:     NewADlist(),
			pattern: "",
			wantOK:  false,
		},
		{
			name:    "03 - add tld",
			adl:     NewADlist(),
			pattern: "tld",
			wantOK:  true,
		},
		/* */
		// More tests are done with the trie's method.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := tc.adl.AddDeny(tc.pattern)

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
			adl:     NewADlist(),
			pattern: "",
			wantOK:  false,
		},
		{
			name:    "03 - delete missing tld",
			adl:     NewADlist(),
			pattern: "tld",
			wantOK:  false,
		},
		{
			name: "04 - delete added tld",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddAllow("tld")
				return a
			}(),
			pattern: "tld",
			wantOK:  true,
		},
		{
			name: "05 - delete added domain.tld",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddAllow("domain.tld")
				return a
			}(),
			pattern: "domain.tld",
			wantOK:  true,
		},
		{
			name: "06 - delete added sub.domain.tld",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddAllow("sub.domain.tld")
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
			gotOK := tc.adl.DeleteAllow(tc.pattern)

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
			adl:     NewADlist(),
			pattern: "",
			wantOK:  false,
		},
		{
			name:    "03 - delete missing tld",
			adl:     NewADlist(),
			pattern: "tld",
			wantOK:  false,
		},
		{
			name: "04 - delete added tld",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddDeny("tld")
				return a
			}(),
			pattern: "tld",
			wantOK:  true,
		},
		{
			name: "05 - delete added domain.tld",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddDeny("domain.tld")
				return a
			}(),
			pattern: "domain.tld",
			wantOK:  true,
		},
		{
			name: "06 - delete added sub.domain.tld",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddDeny("sub.domain.tld")
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
			gotOK := tc.adl.DeleteDeny(tc.pattern)

			if gotOK != tc.wantOK {
				t.Errorf("TADlist.DeleteDeny() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_TADlist_DeleteDeny()

func Test_TADlist_LoadAllow(t *testing.T) {
	tmpDir := t.TempDir()
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
			fName:   filepath.Join(tmpDir, "allow.txt"),
			wantErr: true,
		},
		{
			name:    "02 - non-existent file",
			adl:     NewADlist(),
			fName:   filepath.Join(tmpDir, "doesnotexist.txt"),
			wantErr: true,
		},
		{
			name:    "03 - empty file",
			adl:     NewADlist(),
			fName:   filepath.Join(tmpDir, "empty.txt"),
			wantErr: true,
		},
		{
			name:    "04 - valid file",
			adl:     NewADlist(),
			fName:   filepath.Join(tmpDir, "allow.txt"),
			wantErr: true,
		},
		/* */
		// TODO: Add test cases.
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.adl.LoadAllow(tc.fName)

			if (nil != err) != tc.wantErr {
				t.Errorf("TADlist.LoadAllow() error = '%v', wantErr '%v'",
					err, tc.wantErr)
			}
		})
	}
} // Test_TADlist_LoadAllow()

func Test_TADlist_LoadDeny(t *testing.T) {
	tmpDir := t.TempDir()

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
			fName:   filepath.Join(tmpDir, "deny.txt"),
			wantErr: true,
		},
		{
			name:    "02 - non-existent file",
			adl:     NewADlist(),
			fName:   filepath.Join(tmpDir, "doesnotexist.txt"),
			wantErr: true,
		},
		{
			name:    "03 - empty file",
			adl:     NewADlist(),
			fName:   filepath.Join(tmpDir, "empty.txt"),
			wantErr: true,
		},
		{
			name:    "04 - valid file",
			adl:     NewADlist(),
			fName:   filepath.Join(tmpDir, "deny.txt"),
			wantErr: true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.adl.LoadDeny(tc.fName)

			if (nil != err) != tc.wantErr {
				t.Errorf("TADlist.LoadDeny() error = '%v', wantErr '%v'",
					err, tc.wantErr)
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
			adl:      NewADlist(),
			hostname: "",
			want:     ADneutral,
		},
		{
			name:     "03 - non-matching hostname",
			adl:      NewADlist(),
			hostname: "nothing.will.be.matched.in.an.empty.tree",
			want:     ADneutral,
		},
		{
			name: "04 - match allow tld",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddAllow("tld")
				return a
			}(),
			hostname: "tld",
			want:     ADallow,
		},
		{
			name: "05 - match allow domain.tld",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddAllow("domain.tld")
				return a
			}(),
			hostname: "domain.tld",
			want:     ADallow,
		},
		{
			name: "06 - match allow sub.domain.tld",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddAllow("sub.domain.tld")
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
			got := tc.adl.Match(tc.hostname)
			if got != tc.want {
				t.Errorf("TADlist.Match() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_TADlist_Match()

func Test_TADlist_StoreAllow(t *testing.T) {
	tmpDir := t.TempDir()
	tests := []struct {
		name      string
		adl       *TADlist
		aFilename string
		wantErr   bool
	}{
		/* */
		{
			name:      "01 - nil list",
			adl:       nil,
			aFilename: filepath.Join(tmpDir, "allow.txt"),
			wantErr:   true,
		},
		{
			name:      "02 - empty filename",
			adl:       NewADlist(),
			aFilename: "",
			wantErr:   true,
		},
		{
			name:      "03 - valid filename",
			adl:       NewADlist(),
			aFilename: filepath.Join(tmpDir, "allow.txt"),
			wantErr:   false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.adl.StoreAllow(tc.aFilename)

			if (nil != err) != tc.wantErr {
				t.Errorf("TADlist.StoreAllow() error = '%v', wantErr '%v'",
					err, tc.wantErr)
			}
		})
	}
} // Test_TADlist_StoreAllow()

func Test_TADlist_StoreDeny(t *testing.T) {
	tmpDir := t.TempDir()
	tests := []struct {
		name      string
		adl       *TADlist
		aFilename string
		wantErr   bool
	}{
		/* */
		{
			name:      "01 - nil list",
			adl:       nil,
			aFilename: filepath.Join(tmpDir, "deny.txt"),
			wantErr:   true,
		},
		{
			name:      "02 - empty filename",
			adl:       NewADlist(),
			aFilename: "",
			wantErr:   true,
		},
		{
			name:      "03 - valid filename",
			adl:       NewADlist(),
			aFilename: filepath.Join(tmpDir, "deny.txt"),
			wantErr:   false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.adl.StoreDeny(tc.aFilename)

			if (nil != err) != tc.wantErr {
				t.Errorf("TADlist.StoreDeny() error = '%v', wantErr '%v'",
					err, tc.wantErr)
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
			adl:  NewADlist(),
			want: "Allow:\n\"Trie\":\n  isEnd: false\n  isWild: false\n\nDeny:\n\"Trie\":\n  isEnd: false\n  isWild: false\n",
		},
		{
			name: "03 - one allow pattern",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddAllow("tld")
				return a
			}(),
			want: "Allow:\n\"Trie\":\n  isEnd: false\n  isWild: false\n  \"tld\":\n      isEnd: true\n      isWild: false\n\nDeny:\n\"Trie\":\n  isEnd: false\n  isWild: false\n",
		},
		{
			name: "04 - one deny pattern",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddDeny("tld")
				return a
			}(),
			want: "Allow:\n\"Trie\":\n  isEnd: false\n  isWild: false\n\nDeny:\n\"Trie\":\n  isEnd: false\n  isWild: false\n  \"tld\":\n      isEnd: true\n      isWild: false\n",
		},
		{
			name: "05 - two allow patterns",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddAllow("tld")
				a.AddAllow("domain.tld")
				return a
			}(),
			want: "Allow:\n\"Trie\":\n  isEnd: false\n  isWild: false\n  \"tld\":\n      isEnd: true\n      isWild: false\n      \"domain\":\n          isEnd: true\n          isWild: false\n\nDeny:\n\"Trie\":\n  isEnd: false\n  isWild: false\n",
		},
		{
			name: "06 - two deny patterns",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddDeny("tld")
				a.AddDeny("domain.tld")
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
			adl:        NewADlist(),
			oldPattern: "",
			newPattern: "tld",
			wantOK:     false,
		},
		{
			name:       "03 - empty new pattern",
			adl:        NewADlist(),
			oldPattern: "tld",
			newPattern: "",
			wantOK:     false,
		},
		{
			name:       "04 - equal old and new pattern",
			adl:        NewADlist(),
			oldPattern: "tld",
			newPattern: "tld",
			wantOK:     false,
		},
		{
			name: "05 - update non-existent old pattern",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddAllow("tld")
				return a
			}(),
			oldPattern: "domain.tld",
			newPattern: "sub.domain.tld",
			wantOK:     true,
		},
		{
			name: "06 - update existing old pattern",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddAllow("tld")
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
			gotOK := tc.adl.UpdateAllow(tc.oldPattern, tc.newPattern)
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
			adl:        NewADlist(),
			oldPattern: "",
			newPattern: "tld",
			wantOK:     false,
		},
		{
			name:       "03 - empty new pattern",
			adl:        NewADlist(),
			oldPattern: "tld",
			newPattern: "",
			wantOK:     false,
		},
		{
			name:       "04 - equal old and new pattern",
			adl:        NewADlist(),
			oldPattern: "tld",
			newPattern: "tld",
			wantOK:     false,
		},
		{
			name: "05 - update non-existent old pattern",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddDeny("tld")
				return a
			}(),
			oldPattern: "domain.tld",
			newPattern: "sub.domain.tld",
			wantOK:     true,
		},
		{
			name: "06 - update existing old pattern",
			adl: func() *TADlist {
				a := NewADlist()
				a.AddDeny("tld")
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
			gotOK := tc.adl.UpdateDeny(tc.oldPattern, tc.newPattern)
			if gotOK != tc.wantOK {
				t.Errorf("TADlist.UpdateDeny() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_TADlist_UpdateDeny()

/* _EoF_ */
