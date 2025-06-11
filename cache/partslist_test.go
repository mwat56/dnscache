/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"slices"
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_pattern2parts(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    tPartsList
	}{
		/* */
		{
			name:    "01 - empty pattern",
			pattern: "",
			want:    nil,
		},
		{
			name:    "02 - tld",
			pattern: "tld",
			want:    tPartsList{"tld"},
		},
		{
			name:    "03 - domain.tld",
			pattern: "domain.tld",
			want:    tPartsList{"tld", "domain"},
		},
		{
			name:    "04 - sub.domain.tld",
			pattern: "sub.domain.tld",
			want:    tPartsList{"tld", "domain", "sub"},
		},
		{
			name:    "05 - host.sub.domain.tld",
			pattern: "host.sub.domain.tld",
			want:    tPartsList{"tld", "domain", "sub", "host"},
		},
		/* */
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
				t.Errorf("pattern2parts() =\n'%v'\nwant 'nil'",
					got)
				return
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("pattern2parts() =\n'%v'\nwant\n'%v'",
					got, tc.want)
			}
		})
	}
} // Test_pattern2parts()

func Test_sortHostnames(t *testing.T) {
	tests := []struct {
		name      string
		hosts     []string
		wantHosts []string
	}{
		/* */
		{
			name:      "01 - nil list",
			hosts:     nil,
			wantHosts: nil,
		},
		{
			name:      "02 - empty list",
			hosts:     []string{},
			wantHosts: []string{},
		},
		{
			name:      "03 - one element",
			hosts:     []string{"tld"},
			wantHosts: []string{"tld"},
		},
		{
			name:      "04 - two elements",
			hosts:     []string{"domain.tld", "tld"},
			wantHosts: []string{"tld", "domain.tld"},
		},
		{
			name:      "05 - three elements",
			hosts:     []string{"domain.tld", "tld", "sub.domain.tld"},
			wantHosts: []string{"tld", "domain.tld", "sub.domain.tld"},
		},
		{
			name:      "06 - four elements",
			hosts:     []string{"domain.tld", "tld", "host.sub.domain.tld", "sub.domain.tld"},
			wantHosts: []string{"tld", "domain.tld", "sub.domain.tld", "host.sub.domain.tld"},
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.hosts
			sortHostnames(got)

			if nil == got {
				if nil != tc.wantHosts {
					t.Error("sortHostnames() = nil, want non-nil")
				}
				return
			}
			if nil == tc.wantHosts {
				t.Errorf("sortHostnames() =\n%v\nwant 'nil'",
					got)
				return
			}
			if !slices.Equal(got, tc.wantHosts) {
				t.Errorf("sortHostnames() =\n%v\nwant\n%v",
					got, tc.wantHosts)
			}
		})
	}
} // Test_sortHostnames()

func Test_tPartsList_Equal(t *testing.T) {
	tests := []struct {
		name  string
		pl    tPartsList
		other tPartsList
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
			other: tPartsList{"tld"},
			want:  false,
		},
		{
			name:  "03 - non-nil list and nil list",
			pl:    tPartsList{"tld"},
			other: nil,
			want:  false,
		},
		{
			name:  "04 - different length",
			pl:    tPartsList{"tld"},
			other: tPartsList{"tld", "domain"},
			want:  false,
		},
		{
			name:  "05 - same length but different content",
			pl:    tPartsList{"tld"},
			other: tPartsList{"domain"},
			want:  false,
		},
		{
			name:  "06 - same length and same content",
			pl:    tPartsList{"tld"},
			other: tPartsList{"tld"},
			want:  true,
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.pl.Equal(tc.other); got != tc.want {
				t.Errorf("tPartsList.Equal() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_tPartsList_Equal()

func Test_tPartsList_Len(t *testing.T) {
	tests := []struct {
		name string
		pl   tPartsList
		want int
	}{
		{
			name: "01 - nil list",
			pl:   nil,
			want: 0,
		},
		{
			name: "02 - empty list",
			pl:   tPartsList{},
			want: 0,
		},
		{
			name: "03 - one element",
			pl:   tPartsList{"tld"},
			want: 1,
		},
		{
			name: "04 - two elements",
			pl:   tPartsList{"tld", "domain"},
			want: 2,
		},
		{
			name: "05 - three elements",
			pl:   tPartsList{"tld", "domain", "sub"},
			want: 3,
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.pl.Len(); got != tc.want {
				t.Errorf("tPartsList.Len() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_tPartsList_Len()

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
			want: "tld",
		},
		{
			name: "04 - two elements",
			pl:   tPartsList{"tld", "domain"},
			want: "tld.domain",
		},
		{
			name: "05 - three elements",
			pl:   tPartsList{"tld", "domain", "sub"},
			want: "tld.domain.sub",
		},
		{
			name: "06 - four elements",
			pl:   tPartsList{"tld", "domain", "sub", "host"},
			want: "tld.domain.sub.host",
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.pl.String(); got != tc.want {
				t.Errorf("tPartsList.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tPartsList_String()

/* _EoF_ */
