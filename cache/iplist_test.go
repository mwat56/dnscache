/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"net"
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_TIpList_Equal(t *testing.T) {
	tests := []struct {
		name  string
		il    tIpList
		other tIpList
		want  bool
	}{
		{
			name:  "01 - nil lists",
			il:    nil,
			other: nil,
			want:  true,
		},
		{
			name:  "02 - nil list and non-nil list",
			il:    nil,
			other: tIpList{net.ParseIP("192.168.1.1")},
			want:  false,
		},
		{
			name:  "03 - non-nil list and nil list",
			il:    tIpList{net.ParseIP("192.168.1.1")},
			other: nil,
			want:  false,
		},
		{
			name: "04 - different length",
			il:   tIpList{net.ParseIP("192.168.1.1")},
			other: tIpList{net.ParseIP("192.168.1.1"),
				net.ParseIP("192.168.1.2")},
			want: false,
		},
		{
			name:  "05 - same length but different content",
			il:    tIpList{net.ParseIP("192.168.1.1")},
			other: tIpList{net.ParseIP("192.168.1.2")},
			want:  false,
		},
		{
			name:  "06 - same length and same content",
			il:    tIpList{net.ParseIP("192.168.1.1")},
			other: tIpList{net.ParseIP("192.168.1.1")},
			want:  true,
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.il.Equal(tc.other); got != tc.want {
				t.Errorf("TIpList.Equal() = %v, want %v", got, tc.want)
			}
		})
	}
} // Test_TIpList_Equal()

func Test_TIpList_First(t *testing.T) {
	tests := []struct {
		name string
		il   tIpList
		want net.IP
	}{
		{
			name: "01 - nil list",
			il:   nil,
			want: nil,
		},
		{
			name: "02 - empty list",
			il:   tIpList{},
			want: nil,
		},
		{
			name: "03 - one element",
			il:   tIpList{net.ParseIP("192.168.1.1")},
			want: net.ParseIP("192.168.1.1"),
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.il.First()
			if nil == got {
				if nil != tc.want {
					t.Error("TIpList.First() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("TIpList.First() = %v, want 'nil'",
					got)
				return
			}
			if !got.Equal(tc.want) {
				t.Errorf("TIpList.First() = %v, want %v", got, tc.want)
			}
		})
	}
} // Test_TIpList_First()

func Test_TIpList_String(t *testing.T) {
	tests := []struct {
		name string
		il   tIpList
		want string
	}{
		{
			name: "01 - nil list",
			il:   nil,
			want: "",
		},
		{
			name: "02 - empty list",
			il:   tIpList{},
			want: "",
		},
		{
			name: "03 - one element",
			il:   tIpList{net.ParseIP("192.168.1.1")},
			want: "192.168.1.1",
		},
		{
			name: "04 - two elements",
			il:   tIpList{net.ParseIP("192.168.1.1"), net.ParseIP("192.168.1.2")},
			want: "192.168.1.1 - 192.168.1.2",
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.il.String(); got != tc.want {
				t.Errorf("TIpList.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_TIpList_String()

/* _EoF_ */
