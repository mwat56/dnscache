/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_tPartsList_Equal(t *testing.T) {
	tests := []struct {
		name string
		pl   tPartsList
		list tPartsList
		want bool
	}{
		{
			name: "01 - nil list",
			pl:   nil,
			list: nil,
			want: true,
		},
		{
			name: "02 - nil list and non-nil list",
			pl:   nil,
			list: tPartsList{"tld"},
			want: false,
		},
		{
			name: "03 - non-nil list and nil list",
			pl:   tPartsList{"tld"},
			list: nil,
			want: false,
		},
		{
			name: "04 - different length",
			pl:   tPartsList{"tld"},
			list: tPartsList{"tld", "domain"},
			want: false,
		},
		{
			name: "05 - same length but different content",
			pl:   tPartsList{"tld"},
			list: tPartsList{"domain"},
			want: false,
		},
		{
			name: "06 - same length and same content",
			pl:   tPartsList{"tld"},
			list: tPartsList{"tld"},
			want: true,
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.pl.Equal(tc.list)
			if got != tc.want {
				t.Errorf("tPartsList.Equal() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_tPartsList_Equal()

func Test_tPartsList_String(t *testing.T) {
	tests := []struct {
		name string
		pl   tPartsList
		want string
	}{
		{
			name: "01 - nil list",
			pl:   nil,
			want: "",
		},
		{
			name: "02 - empty list",
			pl:   tPartsList{},
			want: "",
		},
		{
			name: "03 - one element",
			pl:   tPartsList{"tld"},
			want: "[ 'tld' ]",
		},
		{
			name: "04 - two elements",
			pl:   tPartsList{"tld", "domain"},
			want: "[ 'tld' 'domain' ]",
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.pl.String()

			if got != tc.want {
				t.Errorf("tPartsList.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tPartsList_String()

/* _EoF_ */
