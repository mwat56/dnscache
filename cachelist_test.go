/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package dnscache

import (
	"fmt"
	"net"
	"slices"
	"testing"
	"time"
)

func Test_tCacheEntry_clone(t *testing.T) {
	tc1 := tIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}
	t1 := time.Now()

	tests := []struct {
		name string
		ce   *tCacheEntry
		want *tCacheEntry
	}{
		{
			name: "01 - clone",
			ce: &tCacheEntry{
				ips:        tc1,
				bestBefore: t1.Add(defTTL),
			},
			want: &tCacheEntry{
				ips:        tc1,
				bestBefore: t1.Add(defTTL),
			},
		},
		{
			name: "02 - clone nil",
			ce:   nil,
			want: nil,
		},
		{
			name: "03 - clone empty",
			ce: &tCacheEntry{
				ips:        nil,
				bestBefore: t1.Add(defTTL),
			},
			want: &tCacheEntry{
				ips:        nil,
				bestBefore: t1.Add(defTTL),
			},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.ce.clone()
			if nil == got {
				if nil != tc.want {
					t.Error("tCacheEntry.clone() = nil, want non-nil")
				}
				return
			}

			if !tc.want.Equal(got) {
				t.Errorf("tCacheEntry.clone() =\n%v\nwant\n%v",
					got, tc.want)
			}
		})
	}
} // Test_tCacheEntry_clone()

func Test_tCacheEntry_Equal(t *testing.T) {
	tc1 := tIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}
	tc2 := tIpList{
		net.ParseIP("192.168.1.3"),
		net.ParseIP("192.168.1.4"),
	}
	tc3 := tIpList{
		net.ParseIP("192.168.1.3"),
		net.ParseIP("192.168.1.4"),
		net.ParseIP("192.168.1.5"),
	}
	tc4 := &tCacheEntry{
		ips: tc1,
	}

	tests := []struct {
		name string
		ce   *tCacheEntry
		oc   *tCacheEntry
		want bool
	}{
		{
			name: "equal",
			ce: &tCacheEntry{
				ips: tc1,
			},
			oc: &tCacheEntry{
				ips: tc1,
			},
			want: true,
		},
		{
			name: "not equal entries",
			ce: &tCacheEntry{
				ips: tc1,
			},
			oc: &tCacheEntry{
				ips: tc2,
			},
			want: false,
		},
		{
			name: "not equal lists",
			ce: &tCacheEntry{
				ips: tc2,
			},
			oc: &tCacheEntry{
				ips: tc3,
			},
			want: false,
		},
		{
			name: "same object",
			ce:   tc4,
			oc:   tc4,
			want: true,
		},
		{
			name: "nil ce",
			ce:   nil,
			oc:   &tCacheEntry{},
			want: false,
		},
		{
			name: "nil other",
			ce:   &tCacheEntry{},
			oc:   nil,
			want: false,
		},
		{
			name: "nil ce and oc",
			ce:   nil,
			oc:   nil,
			want: true,
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.ce.Equal(tc.oc)
			if got != tc.want {
				t.Errorf("tCacheEntry.Equal() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_tCacheEntry_Equal()

func Test_tCacheEntry_isExpired(t *testing.T) {
	tests := []struct {
		name        string
		ce          *tCacheEntry
		wantExpired bool
	}{
		{
			name: "01 - expired",
			ce: &tCacheEntry{
				bestBefore: time.Now().Add(-time.Hour),
			},
			wantExpired: true,
		},
		{
			name: "02 - not expired",
			ce: &tCacheEntry{
				bestBefore: time.Now().Add(time.Hour),
			},
			wantExpired: false,
		},
		{
			name: "03 - expired at creation",
			ce: &tCacheEntry{
				bestBefore: time.Now().Add(-time.Minute),
			},
			wantExpired: true,
		},
		{
			name: "04 - expired just now",
			ce: &tCacheEntry{
				bestBefore: time.Now().Add(-time.Minute),
			},
			wantExpired: true,
		},
		{
			name: "05 - not expired yet",
			ce: &tCacheEntry{
				bestBefore: time.Now().Add(-time.Minute).Add(time.Hour),
			},
			wantExpired: false,
		},
		{
			name: "06 - expired in the future",
			ce: &tCacheEntry{
				bestBefore: time.Now().Add(time.Hour).Add(time.Minute),
			},
			wantExpired: false,
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.ce.isExpired(); got != tc.wantExpired {
				t.Errorf("tCacheEntry.isExpired() = '%v', want '%v'",
					got, tc.wantExpired)
			}
		})
	}
} // Test_tCacheEntry_isExpired()

func Test_tCacheEntry_String(t *testing.T) {
	t1 := time.Now()
	tc1 := tIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
		net.ParseIP("192.168.1.3"),
	}
	te1 := &tCacheEntry{
		ips:        tc1,
		bestBefore: t1.Add(time.Hour),
	}
	tests := []struct {
		name string
		ce   *tCacheEntry
		want string
	}{
		{
			name: "01 - string representation",
			ce:   te1,
			want: fmt.Sprintf("%s\n%s",
				tc1.String(),
				te1.bestBefore.Format(defTimeFormat)),
		},
		{
			name: "02 - nil ce",
			ce:   nil,
			want: "",
		},
		{
			name: "03 - empty ce",
			ce: &tCacheEntry{
				ips:        nil,
				bestBefore: t1.Add(time.Hour),
			},
			want: te1.bestBefore.Format(defTimeFormat),
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.ce.String(); got != tc.want {
				t.Errorf("tCacheEntry.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tCacheEntry_String()

func Test_tCacheEntry_update(t *testing.T) {
	tests := []struct {
		name   string
		ce     *tCacheEntry
		newIPs tIpList
		wantCE *tCacheEntry
	}{
		{
			name: "01 - update with different IPs",
			ce: &tCacheEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			newIPs: tIpList{
				net.ParseIP("192.168.1.3"),
				net.ParseIP("192.168.1.4"),
			},
			wantCE: &tCacheEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.3"),
					net.ParseIP("192.168.1.4"),
				},
			},
		},
		{
			name: "02 - update with same IPs",
			ce: &tCacheEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			newIPs: tIpList{
				net.ParseIP("192.168.1.1"),
				net.ParseIP("192.168.1.2"),
			},
			wantCE: &tCacheEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
		},
		{
			name: "03 - update with nil IPs",
			ce: &tCacheEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			newIPs: nil,
			wantCE: &tCacheEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
		},
		{
			name: "04 - update with empty IPs",
			ce: &tCacheEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			newIPs: tIpList{},
			wantCE: &tCacheEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.ce.update(tc.newIPs, time.Minute)

			ok := slices.EqualFunc(tc.ce.ips, got.ips, func(ip1, ip2 net.IP) bool {
				return ip1.Equal(ip2)
			})
			if !ok {
				t.Errorf("tCacheEntry.update() = '%v', want '%v'",
					got, tc.wantCE)
			}
		})
	}
} // Test_tCacheEntry_update()

func Test_newCacheList(t *testing.T) {
	tests := []struct {
		name string
		size uint
		want *tCacheList
	}{
		{
			name: "zero size",
			size: 0,
			want: &tCacheList{},
		},
		{
			name: "positive size",
			size: 1,
			want: &tCacheList{},
		},
		{
			name: "default size",
			size: defCacheSize,
			want: &tCacheList{},
		},
		{
			name: "large size",
			size: 1024,
			want: &tCacheList{},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := newCacheList(tc.size)
			if nil == got {
				t.Error("newCacheList() = nil, want non-nil")
				return
			}

			if nil == tc.want {
				t.Errorf("newCacheList() = %v, want 'nil'",
					got)
				return
			}

			if !tc.want.Equal(got) {
				t.Errorf("newCacheList() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_newCacheList()

func Test_tCacheList_clone(t *testing.T) {
	tests := []struct {
		name string
		cl   *tCacheList
		want *tCacheList
	}{
		{
			name: "clone",
			cl: &tCacheList{
				"example.com": &tCacheEntry{
					ips: tIpList{
						net.ParseIP("192.168.1.1"),
						net.ParseIP("192.168.1.2"),
					},
				},
			},
			want: &tCacheList{
				"example.com": &tCacheEntry{
					ips: tIpList{
						net.ParseIP("192.168.1.1"),
						net.ParseIP("192.168.1.2"),
					},
				},
			},
		},
		{
			name: "clone nil",
			cl:   nil,
			want: nil,
		},
		{
			name: "clone empty",
			cl:   &tCacheList{},
			want: &tCacheList{},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cl.clone()
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
} // Test_tCacheList_clone()

func Test_tCacheList_delete(t *testing.T) {
	h1 := "example.com"
	h2 := "example.org"
	tc1 := tIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}

	tests := []struct {
		name string
		cl   *tCacheList
		host string
		want *tCacheList
	}{
		{
			name: "delete",
			cl: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
			host: h1,
			want: &tCacheList{},
		},
		{
			name: "delete nil",
			cl:   nil,
			host: h1,
			want: nil,
		},
		{
			name: "delete empty",
			cl:   &tCacheList{},
			host: h1,
			want: &tCacheList{},
		},
		{
			name: "delete non-existent",
			cl: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
			host: h2,
			want: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cl.delete(tc.host)
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
} // Test_tCacheList_delete()

func Test_tCacheList_Equal(t *testing.T) {
	h1 := "example.com"
	h2 := "example.org"
	tc1 := tIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}
	tc2 := tIpList{
		net.ParseIP("192.168.1.3"),
		net.ParseIP("192.168.1.4"),
	}

	tests := []struct {
		name string
		cl   *tCacheList
		ol   *tCacheList
		want bool
	}{
		{
			name: "equal",
			cl: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
			ol: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
			want: true,
		},
		{
			name: "not equal",
			cl: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
			ol: &tCacheList{
				h1: &tCacheEntry{
					ips: tc2,
				},
			},
			want: false,
		},
		{
			name: "nil cl",
			cl:   nil,
			ol: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
			want: false,
		},
		{
			name: "nil ol",
			cl: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
			ol:   nil,
			want: false,
		},
		{
			name: "nil cl and ol",
			cl:   nil,
			ol:   nil,
			want: true,
		},
		{
			name: "different length",
			cl: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
				h2: &tCacheEntry{
					ips: tc2,
				},
			},
			ol: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
			want: false,
		},
		{
			name: "different hostnames",
			cl: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
			ol: &tCacheList{
				h2: &tCacheEntry{
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

	tests := []struct {
		name string
		cl   *tCacheList
		want *tCacheList
	}{
		{
			name: "expire entries",
			cl: &tCacheList{
				h1: &tCacheEntry{
					bestBefore: time.Now().Add(-time.Hour).Add(time.Minute),
				},
				h2: &tCacheEntry{
					bestBefore: time.Now().Add(time.Hour),
				},
				h3: &tCacheEntry{
					bestBefore: time.Now().Add(-time.Hour).Add(time.Minute),
				},
			},
			want: &tCacheList{
				h2: &tCacheEntry{
					bestBefore: time.Now().Add(time.Hour),
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

func Test_tCacheList_getEntry(t *testing.T) {
	tests := []struct {
		name   string
		cl     *tCacheList
		host   string
		want   *tCacheEntry
		wantOK bool
	}{
		{
			name: "01 - found",
			cl: &tCacheList{
				"example.com": &tCacheEntry{
					ips: tIpList{
						net.ParseIP("192.168.1.1"),
						net.ParseIP("192.168.1.2"),
					},
				},
			},
			host: "example.com",
			want: &tCacheEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			wantOK: true,
		},
		{
			name: "02 - not found",
			cl: &tCacheList{
				"example.com": &tCacheEntry{
					ips: tIpList{
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
			got, gotOK := tc.cl.getEntry(tc.host)

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
} // Test_tCacheList_getEntry()

func Test_tCacheList_ips(t *testing.T) {
	h1 := "example.com"
	h2 := "example.org"
	tc1 := tIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}

	tests := []struct {
		name    string
		cl      *tCacheList
		host    string
		wantIPs tIpList
		wantOK  bool
	}{
		{
			name: "found",
			cl: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
			host:    h1,
			wantIPs: tc1,
			wantOK:  true,
		},
		{
			name: "not found",
			cl: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
			host:    h2,
			wantIPs: nil,
			wantOK:  false,
		},
		{
			name:    "nil cl",
			cl:      nil,
			host:    h2,
			wantIPs: nil,
			wantOK:  false,
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, gotOK := tc.cl.ips(tc.host)

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
} // Test_tCacheList_ips()

func Test_tCacheList_setEntry(t *testing.T) {
	type tArgs struct {
		aHostname string
		aIPs      tIpList
	}

	t1 := time.Now()
	h1 := "example.com"
	tc1 := tIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}
	tc2 := tIpList{
		net.ParseIP("192.168.1.3"),
		net.ParseIP("192.168.1.4"),
	}

	tests := []struct {
		name string
		cl   *tCacheList
		args tArgs
		want *tCacheList
	}{
		/* */
		{
			name: "01 - set new entry",
			cl:   &tCacheList{},
			args: tArgs{
				aHostname: h1,
				aIPs:      tc1,
			},
			want: &tCacheList{
				h1: &tCacheEntry{
					ips:        tc1,
					bestBefore: t1.Add(defTTL),
				},
			},
		},
		{
			name: "02 - update existing entry",
			cl: &tCacheList{
				h1: &tCacheEntry{
					ips:        tc1,
					bestBefore: t1.Add(defTTL),
				},
			},
			args: tArgs{
				aHostname: h1,
				aIPs:      tc2,
			},
			want: &tCacheList{
				h1: &tCacheEntry{
					ips:        tc2,
					bestBefore: t1.Add(defTTL),
				},
			},
		},
		{
			name: "03 - set nil IPs",
			cl: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
			args: tArgs{
				aHostname: h1,
				aIPs:      nil,
			},
			want: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
		},
		{
			name: "04 - set empty IPs",
			cl: &tCacheList{
				h1: &tCacheEntry{
					ips: tc1,
				},
			},
			args: tArgs{
				aHostname: h1,
				aIPs:      tIpList{},
			},
			want: &tCacheList{
				h1: &tCacheEntry{
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
			got := tc.cl.setEntry(tc.args.aHostname, tc.args.aIPs, defTTL)

			if !tc.want.Equal(got) {
				t.Errorf("tCacheList.setEntry() =\n'%v'\nwant\n'%v'",
					got, tc.want)
			}
		})
	}
} // Test_tCacheList_setEntry()

func Test_tCacheList_String(t *testing.T) {
	h1 := "example.com"
	tc1 := tIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
		net.ParseIP("192.168.1.3"),
	}
	t2 := time.Time{}

	tests := []struct {
		name string
		cl   *tCacheList
		want string
	}{
		{
			name: "00 - nil cache list",
			cl:   nil,
			want: "",
		},
		{
			name: "01 - empty cache list",
			cl:   &tCacheList{},
			want: "",
		},
		{
			name: "02 - one entry",
			cl: &tCacheList{
				h1: &tCacheEntry{
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
