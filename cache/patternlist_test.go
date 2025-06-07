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

func Test_tPatternList_Equal(t *testing.T) {
	tests := []struct {
		name  string
		pl    tPatternList
		other tPatternList
		want  bool
	}{
		{
			name:  "01 - nil lists",
			pl:    nil,
			other: nil,
			want:  true,
		},
		{
			name:  "02 - nil list and non-nil list",
			pl:    nil,
			other: tPatternList{"tld"},
			want:  false,
		},
		{
			name:  "03 - non-nil list and nil list",
			pl:    tPatternList{"tld"},
			other: nil,
			want:  false,
		},
		{
			name:  "04 - different length",
			pl:    tPatternList{"tld"},
			other: tPatternList{"tld", "domain"},
			want:  false,
		},
		{
			name:  "05 - same length but different content",
			pl:    tPatternList{"tld"},
			other: tPatternList{"domain"},
			want:  false,
		},
		{
			name:  "06 - same length and same content",
			pl:    tPatternList{"tld"},
			other: tPatternList{"tld"},
			want:  true,
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.pl.Equal(tc.other); got != tc.want {
				t.Errorf("tPatternList.Equal() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_tPatternList_Equal()

func Test_tPatternList_Len(t *testing.T) {
	tests := []struct {
		name string
		pl   tPatternList
		want int
	}{
		{
			name: "01 - nil list",
			pl:   nil,
			want: 0,
		},
		{
			name: "02 - empty list",
			pl:   tPatternList{},
			want: 0,
		},
		{
			name: "03 - one element",
			pl:   tPatternList{"tld"},
			want: 1,
		},
		{
			name: "04 - two elements",
			pl:   tPatternList{"tld", "domain"},
			want: 2,
		},
		{
			name: "05 - three elements",
			pl:   tPatternList{"tld", "domain", "sub"},
			want: 3,
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.pl.Len(); got != tc.want {
				t.Errorf("tPatternList.Len() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_tPatternList_Len()

func Test_tPatternList_String(t *testing.T) {
	tests := []struct {
		name string
		pl   tPatternList
		want string
	}{
		/* */
		{
			name: "01 - nil list",
			pl:   nil,
			want: "",
		},
		{
			name: "02 - empty list",
			pl:   tPatternList{},
			want: "",
		},
		{
			name: "03 - one element",
			pl:   tPatternList{"tld"},
			want: "tld",
		},
		{
			name: "04 - two elements",
			pl:   tPatternList{"tld", "domain"},
			want: "tld\ndomain",
		},
		{
			name: "05 - three elements",
			pl:   tPatternList{"tld", "domain", "sub"},
			want: "tld\ndomain\nsub",
		},
		{
			name: "06 - four elements",
			pl:   tPatternList{"tld", "domain", "sub", "host"},
			want: "tld\ndomain\nsub\nhost",
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.pl.String(); got != tc.want {
				t.Errorf("tPatternList.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tPatternList_String()

/* _EoF_ */
