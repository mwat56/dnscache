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

func Test_newMap(t *testing.T) {
	tests := []struct {
		name string
		size uint
		want *tMapList
	}{
		{
			name: "zero size",
			size: 0,
			want: &tMapList{},
		},
		{
			name: "positive size",
			size: 1,
			want: &tMapList{},
		},
		{
			name: "default size",
			size: DefaultCacheSize,
			want: &tMapList{},
		},
		{
			name: "large size",
			size: 1024,
			want: &tMapList{},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := newMap(tc.size)

			if !tc.want.Equal(got) {
				t.Errorf("newMap() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_newMap()

func Test_tCacheList_Clone(t *testing.T) {
	tests := []struct {
		name string
		cl   *tMapList
		want ICacheList
	}{
		{
			name: "01 - clone nil",
			cl:   nil,
			want: nil,
		},
		{
			name: "02 - clone empty",
			cl:   &tMapList{},
			want: &tMapList{},
		},
		{
			name: "03 - clone",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					"example.com": {
						ips: tIpList{
							net.ParseIP("192.168.1.1"),
							net.ParseIP("192.168.1.2"),
						},
					},
				},
			},
			want: &tMapList{
				Cache: map[string]*tMapEntry{
					"example.com": {
						ips: tIpList{
							net.ParseIP("192.168.1.1"),
							net.ParseIP("192.168.1.2"),
						},
					},
				},
			},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cl.Clone()

			if nil == got {
				if nil != tc.want {
					t.Error("tMapList.clone() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("tMapList.clone() = %v, want 'nil'",
					got)
				return
			}

			gotT, gOK := got.(*tMapList)
			wantT, wOK := tc.want.(*tMapList)
			if !gOK || !wOK {
				t.Errorf("tMapList.clone() = %v, want %v",
					got, tc.want)
				return
			}
			if !wantT.Equal(gotT) {
				t.Errorf("tMapList.clone() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_tCacheList_Clone()

func Test_tCacheList_Create(t *testing.T) {
	type tArgs struct {
		aHostname string
		aIPs      tIpList
	}

	h1 := "example.com"
	tc1 := tIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}
	t2 := time.Now()
	tc2 := tIpList{
		net.ParseIP("192.168.1.3"),
		net.ParseIP("192.168.1.4"),
	}

	tests := []struct {
		name string
		cl   *tMapList
		args tArgs
		want ICacheList
	}{
		/* */
		{
			name: "01 - set new entry",
			cl:   &tMapList{},
			args: tArgs{
				aHostname: h1,
				aIPs:      tc1,
			},
			want: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips:        tc1,
						bestBefore: t2.Add(DefaultTTL),
					},
				},
			},
		},
		{
			name: "02 - update existing entry",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips:        tc1,
						bestBefore: t2.Add(DefaultTTL),
					},
				},
			},
			args: tArgs{
				aHostname: h1,
				aIPs:      tc2,
			},
			want: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips:        tc2,
						bestBefore: t2.Add(DefaultTTL),
					},
				},
			},
		},
		{
			name: "03 - set nil IPs",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
				},
			},
			args: tArgs{
				aHostname: h1,
				aIPs:      nil,
			},
			want: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {},
				},
			},
		},
		{
			name: "04 - set empty IPs",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
				},
			},
			args: tArgs{
				aHostname: h1,
				aIPs:      tIpList{},
			},
			want: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tIpList{},
					},
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
			got := tc.cl.Create(context.TODO(), tc.args.aHostname, tc.args.aIPs, DefaultTTL)
			if nil == got {
				if nil != tc.want {
					t.Error("tMapList.Create() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("tMapList.Create() =\n%q\nwant 'nil'",
					got)
				return
			}
			gotT := got.(*tMapList)
			wantT := tc.want.(*tMapList)
			if !wantT.Equal(gotT) {
				t.Errorf("tMapList.Create() =\n%q\nwant\n%q",
					gotT, wantT)
			}
		})
	}
} // Test_tCacheList_Create()

func Test_tCacheList_Delete(t *testing.T) {
	h1 := "example.com"
	h2 := "example.org"
	tc1 := tIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}

	tests := []struct {
		name string
		cl   *tMapList
		host string
		want bool
	}{
		{
			name: "01 - delete",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
				},
			},
			host: h1,
			want: true,
		},
		{
			name: "02 - delete nil",
			cl:   nil,
			host: h1,
			want: false,
		},
		{
			name: "03 - delete empty",
			cl:   &tMapList{},
			host: h1,
			want: false,
		},
		{
			name: "04 - delete non-existent",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
				},
			},
			host: h2,
			want: false,
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cl.Delete(context.TODO(), tc.host)
			if got != tc.want {
				t.Errorf("tMapList.delete() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_tCacheList_Delete()

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
		cl   *tMapList
		ol   *tMapList
		want bool
	}{
		{
			name: "01 - equal",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
				},
			},
			ol: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
				},
			},
			want: true,
		},
		{
			name: "02 - not equal",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
				},
			},
			ol: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc2,
					},
				},
			},
			want: false,
		},
		{
			name: "03 - nil cl",
			cl:   nil,
			ol: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
				},
			},
			want: false,
		},
		{
			name: "04 - nil ol",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
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
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
					h2: {
						ips: tc2,
					},
				},
			},
			ol: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
				},
			},
			want: false,
		},
		{
			name: "07 - different hostnames",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
				},
			},
			ol: &tMapList{
				Cache: map[string]*tMapEntry{
					h2: {
						ips: tc1,
					},
				},
			},
			want: false,
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cl.Equal(tc.ol); got != tc.want {
				t.Errorf("tMapList.Equal() = '%v', want '%v'",
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
		cl   *tMapList
		want *tMapList
	}{
		{
			name: "01 - expire entries",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tIpList{
							net.ParseIP("192.168.1.1"),
							net.ParseIP("192.168.1.2"),
						},
						bestBefore: t1.Add(-time.Hour).Add(time.Minute),
					},
					h2: {
						ips: tIpList{
							net.ParseIP("192.168.1.3"),
							net.ParseIP("192.168.1.4"),
						},
						bestBefore: t1.Add(time.Hour),
					},
					h3: {
						ips: tIpList{
							net.ParseIP("192.168.1.5"),
							net.ParseIP("192.168.1.6"),
						},
						bestBefore: t1.Add(-time.Hour).Add(time.Minute),
					},
				},
			},
			want: &tMapList{
				Cache: map[string]*tMapEntry{
					h2: {
						ips: tIpList{
							net.ParseIP("192.168.1.3"),
							net.ParseIP("192.168.1.4"),
						},
						bestBefore: t1.Add(time.Hour),
					},
				},
			},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.cl.expireEntries()
			if !tc.cl.Equal(tc.want) {
				t.Errorf("tMapList.expireEntries() =\n'%v'\nwant\n'%v'",
					tc.cl, tc.want)
			}
		})
	}
} // Test_tCacheList_expireEntries()

func Test_tCacheList_Exists(t *testing.T) {
	tests := []struct {
		name   string
		cl     *tMapList
		host   string
		wantOK bool
	}{
		/* */
		{
			name: "01 - found",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					"example.com": {
						ips: tIpList{
							net.ParseIP("192.168.1.1"),
							net.ParseIP("192.168.1.2"),
						},
					},
				},
			},
			host:   "example.com",
			wantOK: true,
		},
		{
			name: "02 - not found",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					"example.com": {
						ips: tIpList{
							net.ParseIP("192.168.2.1"),
							net.ParseIP("192.168.2.2"),
						},
					},
				},
			},
			host:   "example.org",
			wantOK: false,
		},
		{
			name:   "03 - nil cl",
			cl:     nil,
			host:   "example.com",
			wantOK: false,
		},
		{
			name: "04 - empty hostname",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					"example.com": {
						ips: tIpList{
							net.ParseIP("192.168.4.4"),
						},
					},
				},
			},
			host:   " ",
			wantOK: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := tc.cl.Exists(context.TODO(), tc.host)

			if gotOK != tc.wantOK {
				t.Errorf("tMapList.Exists() got = %v, want %v",
					gotOK, tc.wantOK)
				return
			}
		})
	}
} // Test_tCacheList_Exists()

func Test_tCacheList_IPs(t *testing.T) {
	h1 := "example.com"
	h2 := "example.org"
	tc1 := tIpList{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}

	tests := []struct {
		name    string
		cl      *tMapList
		host    string
		wantIPs tIpList
		wantOK  bool
	}{
		{
			name: "01 - found",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
				},
			},
			host:    h1,
			wantIPs: tc1,
			wantOK:  true,
		},
		{
			name: "02 - not found",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
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
		{
			name: "04 - empty hostname",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
				},
			},
			host:    " ",
			wantIPs: nil,
			wantOK:  false,
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, gotOK := tc.cl.IPs(context.TODO(), tc.host)

			if !tc.wantIPs.Equal(got) {
				t.Errorf("tMapList.ips() got = %q, want %q",
					got, tc.wantIPs)
			}

			if gotOK != tc.wantOK {
				t.Errorf("tMapList.ips() got1 = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_tCacheList_IPs()

func Test_tCacheList_Range(t *testing.T) {
	tests := []struct {
		name string
		cl   *tMapList
		want []string // filled in test code
	}{
		/* */
		{
			name: "01 - nil list",
			cl:   nil,
			want: []string{},
		},
		{
			name: "02 - empty list",
			cl:   newMap(0),
			want: []string{},
		},
		{
			name: "03 - one entry",
			cl: func() *tMapList {
				l := newMap(0)
				l.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				return l
			}(),
			want: []string{"tld"},
		},
		{
			name: "04 - two entries",
			cl: func() *tMapList {
				l := newMap(0)
				l.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				l.Create(context.TODO(), "domain.tld", tIpList{net.ParseIP("192.168.1.2")}, 0)
				return l
			}(),
			want: []string{"tld", "domain.tld"},
		},
		{
			name: "05 - three entries",
			cl: func() *tMapList {
				l := newMap(0)
				l.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				l.Create(context.TODO(), "domain.tld", tIpList{net.ParseIP("192.168.1.2")}, 0)
				l.Create(context.TODO(), "sub.domain.tld", tIpList{net.ParseIP("192.168.1.3")}, 0)
				return l
			}(),
			want: []string{"tld", "domain.tld", "sub.domain.tld"},
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cl.Range(context.TODO())

			if nil == got {
				if nil != tc.want {
					t.Error("tTrieList.Range() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("tTrieList.Range() = %v, want 'nil'",
					got)
				return
			}
			var gotList []string
			for fqdn := range got {
				gotList = append(gotList, fqdn)
			}
			if !slices.Equal(gotList, tc.want) {
				t.Errorf("tTrieList.Range() =\n%v\nwant\n%v",
					gotList, tc.want)
			}
		})
	}
} // Test_TTrieList_Range()

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
		cl   *tMapList
		want string
	}{
		{
			name: "00 - nil cache list",
			cl:   nil,
			want: "",
		},
		{
			name: "01 - empty cache list",
			cl:   &tMapList{},
			want: "",
		},
		{
			name: "02 - one entry",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					h1: {
						ips: tc1,
					},
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
				t.Errorf("tMapList.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tCacheList_String()

func TestTCacheList_Update(t *testing.T) {
	type tArgs struct {
		aHostname string
		aIPs      []net.IP
	}
	tests := []struct {
		name string
		cl   *tMapList
		args tArgs
		want ICacheList
	}{
		/* */
		{
			name: "01 - nil cache list",
			cl:   nil,
			args: tArgs{
				aHostname: "example.com",
				aIPs:      []net.IP{net.ParseIP("192.168.1.1")},
			},
			want: nil,
		},
		/* */
		{
			name: "02 - empty cache list",
			cl:   &tMapList{},
			args: tArgs{
				aHostname: "example.com",
				aIPs:      []net.IP{net.ParseIP("192.168.2.2")},
			},
			want: &tMapList{
				Cache: map[string]*tMapEntry{
					"example.com": {
						ips: tIpList{
							net.ParseIP("192.168.2.2"),
						},
					},
				},
			},
		},
		{
			name: "03 - empty arguments",
			cl:   &tMapList{},
			args: tArgs{},
			want: &tMapList{},
		},
		{
			name: "04 - update existing entry",
			cl: &tMapList{
				Cache: map[string]*tMapEntry{
					"example.com": {
						ips: tIpList{
							net.ParseIP("192.168.1.3"),
						},
					},
				},
			},
			args: tArgs{
				aHostname: "example.com",
				aIPs:      []net.IP{net.ParseIP("192.168.3.3")},
			},
			want: &tMapList{
				Cache: map[string]*tMapEntry{
					"example.com": {
						ips: tIpList{
							net.ParseIP("192.168.3.3"),
						},
					},
				},
			},
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cl.Update(context.TODO(), tc.args.aHostname, tc.args.aIPs, 0)

			if nil == got {
				if nil != tc.want {
					t.Error("TCacheList.Update() = nil, want non-nil")
				}
				return
			}
			tGot := got.(*tMapList)
			if nil == tc.want {
				t.Errorf("TCacheList.Update() =\n%q\nwant 'nil'",
					tGot)
				return
			}
			tWant := tc.want.(*tMapList)
			if !tWant.Equal(tGot) {
				t.Errorf("TCacheList.Update() =\n%v\nwant\n%v",
					tGot, tWant)
			}
		})
	}
} // TestTCacheList_Update()

/* _EoF_ */
