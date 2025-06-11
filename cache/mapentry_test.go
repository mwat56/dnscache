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
		ce   *tMapEntry
		want *tMapEntry
	}{
		{
			name: "01 - clone",
			ce: &tMapEntry{
				ips:        tc1,
				bestBefore: t1.Add(DefaultTTL),
			},
			want: &tMapEntry{
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
			ce: &tMapEntry{
				ips:        nil,
				bestBefore: t1.Add(DefaultTTL),
			},
			want: &tMapEntry{
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
					t.Error("tMapEntry.clone() = nil, want non-nil")
				}
				return
			}
			if !tc.want.Equal(got) {
				t.Errorf("tMapEntry.clone() =\n%q\nwant\n%q",
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
		ce   *tMapEntry
		args tArgs
		want bool
	}{
		{
			name: "01 - create",
			ce:   &tMapEntry{},
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
				t.Errorf("tMapEntry.Create() = '%v', want '%v'",
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
		entry *tMapEntry
		args  tArgs
		want  bool
	}{
		{
			name:  "01 - delete",
			entry: &tMapEntry{},
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
	ce4 := &tMapEntry{
		ips: ipl1,
	}

	tests := []struct {
		name  string
		ce    *tMapEntry
		other *tMapEntry
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
			other: &tMapEntry{},
			want:  false,
		},
		{
			name:  "03 - nil other",
			ce:    &tMapEntry{},
			other: nil,
			want:  false,
		},
		{
			name: "04 - equal",
			ce: &tMapEntry{
				ips: ipl1,
			},
			other: &tMapEntry{
				ips: ipl1,
			},
			want: true,
		},
		{
			name: "05 - not equal entries",
			ce: &tMapEntry{
				ips: ipl1,
			},
			other: &tMapEntry{
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
			ce: &tMapEntry{
				ips: nil,
			},
			other: &tMapEntry{
				ips: tIpList{},
			},
			want: false,
		},
		{
			name: "08 - not equal lists",
			ce: &tMapEntry{
				ips: ipl2,
			},
			other: &tMapEntry{
				ips: nil,
			},
			want: false,
		},
		{
			name: "08 - not equal lists",
			ce: &tMapEntry{
				ips: ipl2,
			},
			other: &tMapEntry{
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
				t.Errorf("tMapEntry.Equal() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_tCacheEntry_Equal()

func Test_tCacheEntry_isExpired(t *testing.T) {
	tests := []struct {
		name        string
		ce          *tMapEntry
		wantExpired bool
	}{
		{
			name: "01 - expired",
			ce: &tMapEntry{
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
			ce: &tMapEntry{
				bestBefore: time.Now().Add(time.Hour),
			},
			wantExpired: true,
		},
		{
			name: "03 - expired at creation",
			ce: &tMapEntry{
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
			ce: &tMapEntry{
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
			ce: &tMapEntry{
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
			ce: &tMapEntry{
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
				t.Errorf("tMapEntry.isExpired() = '%v', want '%v'",
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
		ce   *tMapEntry
		args tArgs
		want tIpList
	}{
		/* */
		{
			name: "01 - retrieve",
			ce: &tMapEntry{
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
			ce: &tMapEntry{
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
					t.Error("tMapEntry.Retrieve() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("tMapEntry.Retrieve() =\n%q\nwant 'nil'",
					got)
				return
			}
			if !tc.want.Equal(got) {
				t.Errorf("tMapEntry.Retrieve() =\n%q\nwant\n%q",
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
	te1 := &tMapEntry{
		ips:        tc1,
		bestBefore: t1.Add(time.Hour),
	}
	tests := []struct {
		name string
		ce   *tMapEntry
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
			ce: &tMapEntry{
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
				t.Errorf("tMapEntry.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tCacheEntry_String()

func Test_tCacheEntry_update(t *testing.T) {
	tests := []struct {
		name   string
		ce     *tMapEntry
		newIPs tIpList
		wantCE *tMapEntry
	}{
		{
			name: "01 - update with different IPs",
			ce: &tMapEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			newIPs: tIpList{
				net.ParseIP("192.168.1.3"),
				net.ParseIP("192.168.1.4"),
			},
			wantCE: &tMapEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.3"),
					net.ParseIP("192.168.1.4"),
				},
			},
		},
		{
			name: "02 - update with same IPs",
			ce: &tMapEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			newIPs: tIpList{
				net.ParseIP("192.168.1.1"),
				net.ParseIP("192.168.1.2"),
			},
			wantCE: &tMapEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
		},
		{
			name: "03 - update with nil IPs",
			ce: &tMapEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			newIPs: nil,
			wantCE: &tMapEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
		},
		{
			name: "04 - update with empty IPs",
			ce: &tMapEntry{
				ips: tIpList{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				},
			},
			newIPs: tIpList{},
			wantCE: &tMapEntry{
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
			gotI := tc.ce.Update(context.TODO(), tc.newIPs, time.Minute)
			got := gotI.(*tMapEntry)

			ok := slices.EqualFunc(tc.ce.ips, got.ips, func(ip1, ip2 net.IP) bool {
				return ip1.Equal(ip2)
			})
			if !ok {
				t.Errorf("tMapEntry.Update() = '%v', want '%v'",
					got, tc.wantCE)
			}
		})
	}
} // Test_tCacheEntry_update()

/* _EoF_ */
