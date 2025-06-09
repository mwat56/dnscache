/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"context"
	"fmt"
	"net"
	"slices"
	"testing"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

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
				bestBefore: t1.Add(DefaultTTL),
			},
			want: &tCacheEntry{
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
			ce: &tCacheEntry{
				ips:        nil,
				bestBefore: t1.Add(DefaultTTL),
			},
			want: &tCacheEntry{
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
				t.Errorf("tCacheEntry.clone() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tCacheEntry_clone()

func Test_tCacheEntry_Create(t *testing.T) {
	type tArgs struct {
		aIPs tIpList
		aTTL time.Duration
	}
	tests := []struct {
		name string
		ce   *tCacheEntry
		args tArgs
		want bool
	}{
		{
			name: "01 - create",
			ce:   &tCacheEntry{},
			args: tArgs{
				aIPs: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
				aTTL: time.Hour,
			},
			want: true,
		},
		{
			name: "02 - create nil",
			ce:   nil,
			args: tArgs{
				aIPs: tIpList{
					net.ParseIP("192.168.2.1"),
					net.ParseIP("192.168.2.2"),
				},
				aTTL: time.Hour,
			},
			want: true,
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.ce.Create(context.Background(), nil, tc.args.aIPs, tc.args.aTTL)
			if got != tc.want {
				t.Errorf("tCacheEntry.Create() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_tCacheEntry_Create()

func TestTCacheEntry_Delete(t *testing.T) {
	type tArgs struct {
		in0 context.Context
		in1 tPartsList
	}
	tests := []struct {
		name  string
		entry *tCacheEntry
		args  tArgs
		want  bool
	}{
		{
			name:  "01 - delete",
			entry: &tCacheEntry{},
			args:  tArgs{},
			want:  true,
		},
		{
			name:  "02 - delete nil",
			entry: nil,
			args:  tArgs{},
			want:  true,
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.entry.Delete(tc.args.in0, tc.args.in1)

			if nil != tc.entry {
				if !tc.entry.ips.Equal(tIpList{}) {
					t.Error("TCacheEntry.Delete() did not clear IPs")
				}
				if !tc.entry.bestBefore.IsZero() {
					t.Error("TCacheEntry.Delete() did not clear bestBefore")
				}
			}
			if got != tc.want {
				t.Errorf("TCacheEntry.Delete() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // TestTCacheEntry_Delete()

func Test_tCacheEntry_Equal(t *testing.T) {
	ipl1 := tIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}
	ipl2 := tIpList{
		net.ParseIP("192.168.2.3"),
		net.ParseIP("192.168.2.4"),
	}
	ipl3 := tIpList{
		net.ParseIP("192.168.3.3"),
		net.ParseIP("192.168.3.4"),
		net.ParseIP("192.168.3.5"),
	}
	ce4 := &tCacheEntry{
		ips: ipl1,
	}

	tests := []struct {
		name  string
		ce    *tCacheEntry
		other *tCacheEntry
		want  bool
	}{
		{
			name:  "01 - nil ce and oc",
			ce:    nil,
			other: nil,
			want:  true,
		},
		{
			name:  "02 - nil ce",
			ce:    nil,
			other: &tCacheEntry{},
			want:  false,
		},
		{
			name:  "03 - nil other",
			ce:    &tCacheEntry{},
			other: nil,
			want:  false,
		},
		{
			name: "04 - equal",
			ce: &tCacheEntry{
				ips: ipl1,
			},
			other: &tCacheEntry{
				ips: ipl1,
			},
			want: true,
		},
		{
			name: "05 - not equal entries",
			ce: &tCacheEntry{
				ips: ipl1,
			},
			other: &tCacheEntry{
				ips: ipl2,
			},
			want: false,
		},
		{
			name:  "06 - same object",
			ce:    ce4,
			other: ce4,
			want:  true,
		},
		{
			name: "07 - not equal lists",
			ce: &tCacheEntry{
				ips: nil,
			},
			other: &tCacheEntry{
				ips: tIpList{},
			},
			want: false,
		},
		{
			name: "08 - not equal lists",
			ce: &tCacheEntry{
				ips: ipl2,
			},
			other: &tCacheEntry{
				ips: nil,
			},
			want: false,
		},
		{
			name: "08 - not equal lists",
			ce: &tCacheEntry{
				ips: ipl2,
			},
			other: &tCacheEntry{
				ips: ipl3,
			},
			want: false,
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.ce.Equal(tc.other)
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
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
				bestBefore: time.Now().Add(-time.Hour),
			},
			wantExpired: true,
		},
		{
			name: "02 - expired (no IPs)",
			ce: &tCacheEntry{
				bestBefore: time.Now().Add(time.Hour),
			},
			wantExpired: true,
		},
		{
			name: "03 - expired at creation",
			ce: &tCacheEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
				bestBefore: time.Now().Add(-time.Minute),
			},
			wantExpired: true,
		},
		{
			name: "04 - expired just now",
			ce: &tCacheEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
				bestBefore: time.Now().Add(-time.Minute),
			},
			wantExpired: true,
		},
		{
			name: "05 - not expired yet",
			ce: &tCacheEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
				bestBefore: time.Now().Add(-time.Minute).Add(time.Hour),
			},
			wantExpired: false,
		},
		{
			name: "06 - expired in the future",
			ce: &tCacheEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
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

func Test_tCacheEntry_Retrieve(t *testing.T) {
	type tArgs struct {
		in0 context.Context
		in1 tPartsList
	}
	tests := []struct {
		name string
		ce   *tCacheEntry
		args tArgs
		want tIpList
	}{
		/* */
		{
			name: "01 - retrieve",
			ce: &tCacheEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			args: tArgs{},
			want: tIpList{
				net.ParseIP("192.168.1.1"),
				net.ParseIP("192.168.1.2"),
			},
		},
		{
			name: "02 - retrieve nil",
			ce:   nil,
			args: tArgs{},
			want: tIpList{},
		},
		{
			name: "03 - retrieve empty",
			ce: &tCacheEntry{
				ips: nil,
			},
			args: tArgs{},
			want: nil,
		},
		/* */

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.ce.Retrieve(tc.args.in0, tc.args.in1)

			if nil == got {
				if nil != tc.want {
					t.Error("tCacheEntry.Retrieve() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("tCacheEntry.Retrieve() =\n%q\nwant 'nil'",
					got)
				return
			}
			if !tc.want.Equal(got) {
				t.Errorf("tCacheEntry.Retrieve() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tCacheEntry_Retrieve()

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
			gotI := tc.ce.Update(tc.newIPs, time.Minute)
			got := gotI.(*tCacheEntry)

			ok := slices.EqualFunc(tc.ce.ips, got.ips, func(ip1, ip2 net.IP) bool {
				return ip1.Equal(ip2)
			})
			if !ok {
				t.Errorf("tCacheEntry.Update() = '%v', want '%v'",
					got, tc.wantCE)
			}
		})
	}
} // Test_tCacheEntry_update()

/* _EoF_ */
