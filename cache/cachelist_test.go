/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"fmt"
	"net"
	"testing"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_New(t *testing.T) {
	tests := []struct {
		name string
		size uint
		want TCacheList
	}{
		{
			name: "zero size",
			size: 0,
			want: TCacheList{},
		},
		{
			name: "positive size",
			size: 1,
			want: TCacheList{},
		},
		{
			name: "default size",
			size: DefaultCacheSize,
			want: TCacheList{},
		},
		{
			name: "large size",
			size: 1024,
			want: TCacheList{},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := New(tc.size)

			if !tc.want.Equal(&got) {
				t.Errorf("New() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_New()

func Test_tCacheList_Clone(t *testing.T) {
	tests := []struct {
		name string
		cl   *TCacheList
		want *TCacheList
	}{
		{
			name: "01 - clone",
			cl: &TCacheList{
				"example.com": &TCacheEntry{
					ips: TIpList{
						net.ParseIP("192.168.1.1"),
						net.ParseIP("192.168.1.2"),
					},
				},
			},
			want: &TCacheList{
				"example.com": &TCacheEntry{
					ips: TIpList{
						net.ParseIP("192.168.1.1"),
						net.ParseIP("192.168.1.2"),
					},
				},
			},
		},
		{
			name: "02 - clone nil",
			cl:   nil,
			want: nil,
		},
		{
			name: "03 - clone empty",
			cl:   &TCacheList{},
			want: &TCacheList{},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cl.Clone()
			if nil == got {
				if nil != tc.want {
					t.Error("tCacheList.clone() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("tCacheList.clone() = %v, want 'nil'",
					got)
				return
			}
			if !tc.want.Equal(got) {
				t.Errorf("tCacheList.clone() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_tCacheList_Clone()

func Test_tCacheList_Delete(t *testing.T) {
	h1 := "example.com"
	h2 := "example.org"
	tc1 := TIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}

	tests := []struct {
		name string
		cl   *TCacheList
		host string
		want *TCacheList
	}{
		{
			name: "01 - delete",
			cl: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
			host: h1,
			want: &TCacheList{},
		},
		{
			name: "02 - delete nil",
			cl:   nil,
			host: h1,
			want: nil,
		},
		{
			name: "03 - delete empty",
			cl:   &TCacheList{},
			host: h1,
			want: &TCacheList{},
		},
		{
			name: "04 - delete non-existent",
			cl: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
			host: h2,
			want: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cl.Delete(tc.host)
			if nil == got {
				if nil != tc.want {
					t.Error("tCacheList.delete() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("tCacheList.delete() = %v, want 'nil'",
					got)
				return
			}
			if !tc.want.Equal(got) {
				t.Errorf("tCacheList.delete() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_tCacheList_Delete()

func Test_tCacheList_Equal(t *testing.T) {
	h1 := "example.com"
	h2 := "example.org"
	tc1 := TIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}
	tc2 := TIpList{
		net.ParseIP("192.168.1.3"),
		net.ParseIP("192.168.1.4"),
	}

	tests := []struct {
		name string
		cl   *TCacheList
		ol   *TCacheList
		want bool
	}{
		{
			name: "01 - equal",
			cl: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
			ol: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
			want: true,
		},
		{
			name: "02 - not equal",
			cl: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
			ol: &TCacheList{
				h1: &TCacheEntry{
					ips: tc2,
				},
			},
			want: false,
		},
		{
			name: "03 - nil cl",
			cl:   nil,
			ol: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
			want: false,
		},
		{
			name: "04 - nil ol",
			cl: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
			ol:   nil,
			want: false,
		},
		{
			name: "05 - nil cl and ol",
			cl:   nil,
			ol:   nil,
			want: true,
		},
		{
			name: "06 - different length",
			cl: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
				h2: &TCacheEntry{
					ips: tc2,
				},
			},
			ol: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
			want: false,
		},
		{
			name: "07 - different hostnames",
			cl: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
			ol: &TCacheList{
				h2: &TCacheEntry{
					ips: tc1,
				},
			},
			want: false,
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cl.Equal(tc.ol); got != tc.want {
				t.Errorf("tCacheList.Equal() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_tCacheList_Equal()

func Test_tCacheList_expireEntries(t *testing.T) {
	h1 := "example.com"
	h2 := "example.org"
	h3 := "example.net"
	t1 := time.Now()

	tests := []struct {
		name string
		cl   *TCacheList
		want *TCacheList
	}{
		{
			name: "01 - expire entries",
			cl: &TCacheList{
				h1: &TCacheEntry{
					ips: TIpList{
						net.ParseIP("192.168.1.1"),
						net.ParseIP("192.168.1.2"),
					},
					bestBefore: t1.Add(-time.Hour).Add(time.Minute),
				},
				h2: &TCacheEntry{
					ips: TIpList{
						net.ParseIP("192.168.1.3"),
						net.ParseIP("192.168.1.4"),
					},
					bestBefore: t1.Add(time.Hour),
				},
				h3: &TCacheEntry{
					ips: TIpList{
						net.ParseIP("192.168.1.5"),
						net.ParseIP("192.168.1.6"),
					},
					bestBefore: t1.Add(-time.Hour).Add(time.Minute),
				},
			},
			want: &TCacheList{
				h2: &TCacheEntry{
					ips: TIpList{
						net.ParseIP("192.168.1.3"),
						net.ParseIP("192.168.1.4"),
					},
					bestBefore: t1.Add(time.Hour),
				},
			},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.cl.expireEntries()
			if !tc.cl.Equal(tc.want) {
				t.Errorf("tCacheList.expireEntries() =\n'%v'\nwant\n'%v'",
					tc.cl, tc.want)
			}
		})
	}
} // Test_tCacheList_expireEntries()

func Test_tCacheList_GetEntry(t *testing.T) {
	tests := []struct {
		name   string
		cl     *TCacheList
		host   string
		want   *TCacheEntry
		wantOK bool
	}{
		{
			name: "01 - found",
			cl: &TCacheList{
				"example.com": &TCacheEntry{
					ips: TIpList{
						net.ParseIP("192.168.1.1"),
						net.ParseIP("192.168.1.2"),
					},
				},
			},
			host: "example.com",
			want: &TCacheEntry{
				ips: TIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			wantOK: true,
		},
		{
			name: "02 - not found",
			cl: &TCacheList{
				"example.com": &TCacheEntry{
					ips: TIpList{
						net.ParseIP("192.168.1.1"),
						net.ParseIP("192.168.1.2"),
					},
				},
			},
			host:   "example.org",
			want:   nil,
			wantOK: false,
		},
		{
			name:   "03 - nil cl",
			cl:     nil,
			host:   "example.com",
			want:   nil,
			wantOK: false,
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, gotOK := tc.cl.GetEntry(tc.host)

			if nil == got {
				if nil != tc.want {
					t.Error("tCacheList.getEntry() got = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("tCacheList.getEntry() got = %v, want 'nil'",
					got)
				return
			}
			if gotOK != tc.wantOK {
				t.Errorf("tCacheList.getEntry() got1 = %v, want %v",
					gotOK, tc.wantOK)
			}
			if !tc.want.Equal(got) {
				t.Errorf("tCacheList.getEntry() got = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_tCacheList_GetEntry()

func Test_tCacheList_IPs(t *testing.T) {
	h1 := "example.com"
	h2 := "example.org"
	tc1 := TIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}

	tests := []struct {
		name    string
		cl      *TCacheList
		host    string
		wantIPs TIpList
		wantOK  bool
	}{
		{
			name: "01 - found",
			cl: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
			host:    h1,
			wantIPs: tc1,
			wantOK:  true,
		},
		{
			name: "02 - not found",
			cl: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
			host:    h2,
			wantIPs: nil,
			wantOK:  false,
		},
		{
			name:    "03 - nil cl",
			cl:      nil,
			host:    h2,
			wantIPs: nil,
			wantOK:  false,
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, gotOK := tc.cl.IPs(tc.host)

			if !tc.wantIPs.Equal(got) {
				t.Errorf("tCacheList.ips() got = %q, want %q",
					got, tc.wantIPs)
			}

			if gotOK != tc.wantOK {
				t.Errorf("tCacheList.ips() got1 = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_tCacheList_IPs()

func Test_tCacheList_SetEntry(t *testing.T) {
	type tArgs struct {
		aHostname string
		aIPs      TIpList
	}

	t1 := time.Now()
	h1 := "example.com"
	tc1 := TIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}
	tc2 := TIpList{
		net.ParseIP("192.168.1.3"),
		net.ParseIP("192.168.1.4"),
	}

	tests := []struct {
		name string
		cl   *TCacheList
		args tArgs
		want *TCacheList
	}{
		/* */
		{
			name: "01 - set new entry",
			cl:   &TCacheList{},
			args: tArgs{
				aHostname: h1,
				aIPs:      tc1,
			},
			want: &TCacheList{
				h1: &TCacheEntry{
					ips:        tc1,
					bestBefore: t1.Add(DefaultTTL),
				},
			},
		},
		{
			name: "02 - update existing entry",
			cl: &TCacheList{
				h1: &TCacheEntry{
					ips:        tc1,
					bestBefore: t1.Add(DefaultTTL),
				},
			},
			args: tArgs{
				aHostname: h1,
				aIPs:      tc2,
			},
			want: &TCacheList{
				h1: &TCacheEntry{
					ips:        tc2,
					bestBefore: t1.Add(DefaultTTL),
				},
			},
		},
		{
			name: "03 - set nil IPs",
			cl: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
			args: tArgs{
				aHostname: h1,
				aIPs:      nil,
			},
			want: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
		},
		{
			name: "04 - set empty IPs",
			cl: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
			args: tArgs{
				aHostname: h1,
				aIPs:      TIpList{},
			},
			want: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
		},
		{
			name: "05 - set nil cache list",
			cl:   nil,
			args: tArgs{
				aHostname: h1,
				aIPs:      tc1,
			},
			want: nil,
		},
		/* */

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cl.SetEntry(tc.args.aHostname, tc.args.aIPs, DefaultTTL)

			if !tc.want.Equal(got) {
				t.Errorf("tCacheList.setEntry() =\n'%v'\nwant\n'%v'",
					got, tc.want)
			}
		})
	}
} // Test_tCacheList_SetEntry()

func Test_tCacheList_String(t *testing.T) {
	h1 := "example.com"
	tc1 := TIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
		net.ParseIP("192.168.1.3"),
	}
	t2 := time.Time{}

	tests := []struct {
		name string
		cl   *TCacheList
		want string
	}{
		{
			name: "00 - nil cache list",
			cl:   nil,
			want: "",
		},
		{
			name: "01 - empty cache list",
			cl:   &TCacheList{},
			want: "",
		},
		{
			name: "02 - one entry",
			cl: &TCacheList{
				h1: &TCacheEntry{
					ips: tc1,
				},
			},
			want: fmt.Sprintf("%s: %s\n%s\n",
				h1, tc1.String(), t2.Format(defTimeFormat)),
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cl.String(); got != tc.want {
				t.Errorf("tCacheList.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tCacheList_String()

/* _EoF_ */
