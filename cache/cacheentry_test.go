/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"fmt"
	"net"
	"slices"
	"testing"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_tCacheEntry_clone(t *testing.T) {
	tc1 := TIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}
	t1 := time.Now()

	tests := []struct {
		name string
		ce   *TCacheEntry
		want *TCacheEntry
	}{
		{
			name: "01 - clone",
			ce: &TCacheEntry{
				ips:        tc1,
				bestBefore: t1.Add(DefaultTTL),
			},
			want: &TCacheEntry{
				ips:        tc1,
				bestBefore: t1.Add(DefaultTTL),
			},
		},
		{
			name: "02 - clone nil",
			ce:   nil,
			want: nil,
		},
		{
			name: "03 - clone empty",
			ce: &TCacheEntry{
				ips:        nil,
				bestBefore: t1.Add(DefaultTTL),
			},
			want: &TCacheEntry{
				ips:        nil,
				bestBefore: t1.Add(DefaultTTL),
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
	tc1 := TIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}
	tc2 := TIpList{
		net.ParseIP("192.168.1.3"),
		net.ParseIP("192.168.1.4"),
	}
	tc3 := TIpList{
		net.ParseIP("192.168.1.3"),
		net.ParseIP("192.168.1.4"),
		net.ParseIP("192.168.1.5"),
	}
	tc4 := &TCacheEntry{
		ips: tc1,
	}

	tests := []struct {
		name string
		ce   *TCacheEntry
		oc   *TCacheEntry
		want bool
	}{
		{
			name: "equal",
			ce: &TCacheEntry{
				ips: tc1,
			},
			oc: &TCacheEntry{
				ips: tc1,
			},
			want: true,
		},
		{
			name: "not equal entries",
			ce: &TCacheEntry{
				ips: tc1,
			},
			oc: &TCacheEntry{
				ips: tc2,
			},
			want: false,
		},
		{
			name: "not equal lists",
			ce: &TCacheEntry{
				ips: tc2,
			},
			oc: &TCacheEntry{
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
			oc:   &TCacheEntry{},
			want: false,
		},
		{
			name: "nil other",
			ce:   &TCacheEntry{},
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
		ce          *TCacheEntry
		wantExpired bool
	}{
		{
			name: "01 - expired",
			ce: &TCacheEntry{
				bestBefore: time.Now().Add(-time.Hour),
			},
			wantExpired: true,
		},
		{
			name: "02 - not expired",
			ce: &TCacheEntry{
				bestBefore: time.Now().Add(time.Hour),
			},
			wantExpired: false,
		},
		{
			name: "03 - expired at creation",
			ce: &TCacheEntry{
				bestBefore: time.Now().Add(-time.Minute),
			},
			wantExpired: true,
		},
		{
			name: "04 - expired just now",
			ce: &TCacheEntry{
				bestBefore: time.Now().Add(-time.Minute),
			},
			wantExpired: true,
		},
		{
			name: "05 - not expired yet",
			ce: &TCacheEntry{
				bestBefore: time.Now().Add(-time.Minute).Add(time.Hour),
			},
			wantExpired: false,
		},
		{
			name: "06 - expired in the future",
			ce: &TCacheEntry{
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
	tc1 := TIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
		net.ParseIP("192.168.1.3"),
	}
	te1 := &TCacheEntry{
		ips:        tc1,
		bestBefore: t1.Add(time.Hour),
	}
	tests := []struct {
		name string
		ce   *TCacheEntry
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
			ce: &TCacheEntry{
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
		ce     *TCacheEntry
		newIPs TIpList
		wantCE *TCacheEntry
	}{
		{
			name: "01 - update with different IPs",
			ce: &TCacheEntry{
				ips: TIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			newIPs: TIpList{
				net.ParseIP("192.168.1.3"),
				net.ParseIP("192.168.1.4"),
			},
			wantCE: &TCacheEntry{
				ips: TIpList{
					net.ParseIP("192.168.1.3"),
					net.ParseIP("192.168.1.4"),
				},
			},
		},
		{
			name: "02 - update with same IPs",
			ce: &TCacheEntry{
				ips: TIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			newIPs: TIpList{
				net.ParseIP("192.168.1.1"),
				net.ParseIP("192.168.1.2"),
			},
			wantCE: &TCacheEntry{
				ips: TIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
		},
		{
			name: "03 - update with nil IPs",
			ce: &TCacheEntry{
				ips: TIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			newIPs: nil,
			wantCE: &TCacheEntry{
				ips: TIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
		},
		{
			name: "04 - update with empty IPs",
			ce: &TCacheEntry{
				ips: TIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			newIPs: TIpList{},
			wantCE: &TCacheEntry{
				ips: TIpList{
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

/* _EoF_ */
