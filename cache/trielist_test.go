/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"context"
	"net"
	"testing"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_newTrie(t *testing.T) {
	tests := []struct {
		name string
		want *TTrieList
	}{
		{
			name: "01 - new trie",
			want: &TTrieList{
				tRoot: tRoot{
					node: newNode(),
				},
			},
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := newTrie()
			if nil == got {
				if nil != tc.want {
					t.Error("newTrie() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("newTrie() =\n%v\nwant 'nil'",
					got)
				return
			}
			if !tc.want.Equal(got) {
				t.Errorf("newTrie() =\n%v\nwant\n%v",
					got, tc.want)
			}
		})
	}
} // Test_newTrie()

func Test_TTrieList_AutoExpire(t *testing.T) {
	tests := []struct {
		name string
		tl   *TTrieList
	}{
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.tl.AutoExpire(0, nil)
		})
	}
} // Test_TTrieList_AutoExpire()

func Test_TTrieList_Create(t *testing.T) {
	tests := []struct {
		name string
		host string
		ips  tIpList
		tl   *TTrieList
		want iCacheList
	}{
		/* */
		{
			name: "01 - nil list",
			tl:   nil,
			want: nil,
		},
		{
			name: "02 - empty list",
			tl:   newTrie(),
			want: newTrie(),
		},
		{
			name: "03 - set tld",
			host: "tld",
			ips:  tIpList{net.ParseIP("192.168.1.1")},
			tl:   newTrie(),
			want: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				return tl
			}(),
		},
		{
			name: "04 - set domain.tld",
			host: "domain.tld",
			ips:  tIpList{net.ParseIP("192.168.1.2")},
			tl:   newTrie(),
			want: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "domain.tld", tIpList{net.ParseIP("192.168.1.2")}, 0)
				return tl
			}(),
		},
		{
			name: "05 - set sub.domain.tld",
			host: "sub.domain.tld",
			ips:  tIpList{net.ParseIP("192.168.1.3")},
			tl:   newTrie(),
			want: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "sub.domain.tld", tIpList{net.ParseIP("192.168.1.3")}, 0)
				return tl
			}(),
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.tl.Create(context.TODO(), tc.host, tc.ips, 0)

			if nil == got {
				if nil != tc.want {
					t.Error("TTrieList.Create() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("TTrieList.Create() =\n%v\nwant 'nil'",
					got)
				return
			}
			gotTrie := got.(*TTrieList)
			wantTrie := tc.want.(*TTrieList)
			if !wantTrie.Equal(gotTrie) {
				t.Errorf("TTrieList.Create() =\n%v\nwant\n%v",
					got, tc.want)
			}
		})
	}
} // Test_TTrieList_Create()

func Test_TTrieList_Delete(t *testing.T) {
	tests := []struct {
		name string
		host string
		tl   *TTrieList
		want bool
	}{
		/* */
		{
			name: "01 - nil list",
			tl:   nil,
			host: "tld",
			want: false,
		},
		{
			name: "02 - empty list",
			tl:   newTrie(),
			host: "tld",
			want: false,
		},
		{
			name: "03 - delete tld",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				return tl
			}(),
			host: "tld",
			want: true,
		},
		{
			name: "04 - delete domain.tld",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				tl.Create(context.TODO(), "domain.tld", tIpList{net.ParseIP("192.168.1.2")}, 0)
				return tl
			}(),
			host: "domain.tld",
			want: true,
		},
		{
			name: "05 - delete sub.domain.tld",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				tl.Create(context.TODO(), "domain.tld", tIpList{net.ParseIP("192.168.1.2")}, 0)
				tl.Create(context.TODO(), "sub.domain.tld", tIpList{net.ParseIP("192.168.1.3")}, 0)
				return tl
			}(),
			host: "sub.domain.tld",
			want: true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.tl.Delete(context.TODO(), tc.host)
			if !got {
				if tc.want {
					t.Error("TTrieList.Delete() = false, want true")
				}
				return
			}
			if !tc.want {
				t.Error("TTrieList.Delete() = true, want false")
				return
			}
			if nil == tc.tl {
				t.Error("TTrieList.Delete() = nil, want non-nil")
				return
			}
		})
	}
} // Test_TTrieList_Delete()

func Test_TTrieList_Equal(t *testing.T) {
	tests := []struct {
		name  string
		tl    *TTrieList
		other *TTrieList
		want  bool
	}{
		{
			name:  "01 - nil list",
			tl:    nil,
			other: nil,
			want:  true,
		},
		{
			name:  "02 - nil list and non-nil list",
			tl:    nil,
			other: newTrie(),
			want:  false,
		},
		{
			name:  "03 - non-nil list and nil list",
			tl:    newTrie(),
			other: nil,
			want:  false,
		},
		{
			name:  "04 - equal lists",
			tl:    newTrie(),
			other: newTrie(),
			want:  true,
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.tl.Equal(tc.other); got != tc.want {
				t.Errorf("TTrieList.Equal() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_TTrieList_Equal()

func Test_TTrieList_expireEntries(t *testing.T) {
	tests := []struct {
		name string
		tl   *TTrieList
		want *TTrieList
	}{
		/* */
		{
			name: "01 - nil list",
			tl:   nil,
			want: nil,
		},
		{
			name: "02 - empty list",
			tl:   newTrie(),
			want: newTrie(),
		},
		{
			name: "03 - expire tld",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, -time.Hour)
				return tl
			}(),
			want: newTrie(),
		},
		{
			name: "04 - expire domain.tld",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				tl.Create(context.TODO(), "domain.tld", tIpList{net.ParseIP("192.168.1.2")}, -time.Hour)
				return tl
			}(),
			want: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				return tl
			}(),
		},
		{
			name: "05 - expire domain.tld, keeping child",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				tl.Create(context.TODO(), "domain.tld", tIpList{net.ParseIP("192.168.1.2")}, -time.Hour)
				tl.Create(context.TODO(), "sub.domain.tld", tIpList{net.ParseIP("192.168.1.3")}, 0)
				return tl
			}(),
			want: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				tl.Create(context.TODO(), "sub.domain.tld", tIpList{net.ParseIP("192.168.1.3")}, 0)
				return tl
			}(),
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.tl.expireEntries()
			if nil == tc.tl {
				if nil != tc.want {
					t.Error("TTrieList.expireEntries() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("TTrieList.expireEntries() =\n%v\nwant 'nil'",
					tc.tl)
				return
			}
			if !tc.want.Equal(tc.tl) {
				t.Errorf("TTrieList.expireEntries() =\n%v\nwant\n%v",
					tc.tl, tc.want)
			}
		})
	}
} // Test_TTrieList_expireEntries()

func Test_TTrieList_IPs(t *testing.T) {
	tests := []struct {
		name string
		host string
		tl   *TTrieList
		want tIpList
	}{
		/* */
		{
			name: "01 - nil list",
			tl:   nil,
			host: "tld",
			want: nil,
		},
		{
			name: "02 - empty list",
			tl:   newTrie(),
			host: "tld",
			want: nil,
		},
		{
			name: "03 - get tld",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				return tl
			}(),
			host: "tld",
			want: tIpList{net.ParseIP("192.168.1.1")},
		},
		{
			name: "04 - get domain.tld",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				tl.Create(context.TODO(), "domain.tld", tIpList{net.ParseIP("192.168.1.2")}, 0)
				return tl
			}(),
			host: "domain.tld",
			want: tIpList{net.ParseIP("192.168.1.2")},
		},
		{
			name: "05 - get sub.domain.tld",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				tl.Create(context.TODO(), "domain.tld", tIpList{net.ParseIP("192.168.1.2")}, 0)
				tl.Create(context.TODO(), "sub.domain.tld", tIpList{net.ParseIP("192.168.1.3")}, 0)
				return tl
			}(),
			host: "sub.domain.tld",
			want: tIpList{net.ParseIP("192.168.1.3")},
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := tc.tl.IPs(context.TODO(), tc.host)

			if nil == got {
				if nil != tc.want {
					t.Error("TTrieList.IPs() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("TTrieList.IPs() = %v, want 'nil'",
					got)
				return
			}
			if !tc.want.Equal(got) {
				t.Errorf("TTrieList.IPs() =\n%v\nwant\n%v",
					got, tc.want)
			}
		})
	}
} // Test_TTrieList_IPs()

func Test_TTrieList_Len(t *testing.T) {
	tests := []struct {
		name string
		tl   *TTrieList
		want int
	}{
		/* */
		{
			name: "01 - nil list",
			tl:   nil,
			want: 0,
		},
		{
			name: "02 - empty list",
			tl:   newTrie(),
			want: 0,
		},
		{
			name: "03 - one entry",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				return tl
			}(),
			want: 1,
		},
		{
			name: "04 - two entries",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				tl.Create(context.TODO(), "domain.tld", tIpList{net.ParseIP("192.168.1.2")}, 0)
				return tl
			}(),
			want: 2,
		},
		{
			name: "05 - three entries",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				tl.Create(context.TODO(), "domain.tld", tIpList{net.ParseIP("192.168.1.2")}, 0)
				tl.Create(context.TODO(), "sub.domain.tld", tIpList{net.ParseIP("192.168.1.3")}, 0)
				return tl
			}(),
			want: 3,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.tl.Len(); got != tc.want {
				t.Errorf("TTrieList.Len() = %d, want %d",
					got, tc.want)
			}
		})
	}
} // Test_TTrieList_Len()

func Test_TTrieList_String(t *testing.T) {
	tests := []struct {
		name string
		tl   *TTrieList
		want string
	}{
		{
			name: "01 - nil list",
			tl:   nil,
			want: "",
		},
		{
			name: "02 - empty list",
			tl:   newTrie(),
			want: "",
		},
		{
			name: "03 - one entry",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				return tl
			}(),
			want: "192.168.1.1 tld\n",
		},
		{
			name: "04 - two entries",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				tl.Create(context.TODO(), "domain.tld", tIpList{net.ParseIP("192.168.1.2")}, 0)
				return tl
			}(),
			want: "192.168.1.1 tld\n192.168.1.2 domain.tld\n",
		},
		{
			name: "05 - three entries",
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld", tIpList{net.ParseIP("192.168.1.1")}, 0)
				tl.Create(context.TODO(), "domain.tld", tIpList{net.ParseIP("192.168.1.2")}, 0)
				tl.Create(context.TODO(), "sub.domain.tld", tIpList{net.ParseIP("192.168.1.3")}, 0)
				return tl
			}(),
			want: "192.168.1.1 tld\n192.168.1.2 domain.tld\n192.168.1.3 sub.domain.tld\n",
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.tl.String(); got != tc.want {
				t.Errorf("TTrieList.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_TTrieList_String()

func Test_TTrieList_Update(t *testing.T) {
	tests := []struct {
		name string
		host string
		ips  tIpList
		tl   *TTrieList
		want iCacheList
	}{
		/* */
		{
			name: "01 - nil list",
			tl:   nil,
			want: nil,
		},
		{
			name: "02 - empty list",
			tl:   newTrie(),
			want: newTrie(),
		},
		{
			name: "03 - update tld",
			host: "tld",
			ips:  tIpList{net.ParseIP("192.168.3.3")},
			tl:   newTrie(),
			want: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld",
					tIpList{net.ParseIP("192.168.3.3")}, 0)
				return tl
			}(),
		},
		{
			name: "04 - update domain.tld",
			host: "domain.tld",
			ips:  tIpList{net.ParseIP("192.168.4.4")},
			tl:   newTrie(),
			want: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "domain.tld",
					tIpList{net.ParseIP("192.168.4.4")}, 0)
				return tl
			}(),
		},
		{
			name: "05 - update sub.domain.tld",
			host: "sub.domain.tld",
			ips:  tIpList{net.ParseIP("192.168.5.5")},
			tl:   newTrie(),
			want: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "sub.domain.tld",
					tIpList{net.ParseIP("192.168.5.5")}, 0)
				return tl
			}(),
		},
		{
			name: "06 - update tld, existing",
			host: "tld",
			ips:  tIpList{net.ParseIP("192.168.6.6")},
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld",
					tIpList{net.ParseIP("192.168.1.6")}, 0)
				return tl
			}(),
			want: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "tld",
					tIpList{net.ParseIP("192.168.6.6")}, 0)
				return tl
			}(),
		},
		{
			name: "07 - update domain.tld, existing",
			host: "domain.tld",
			ips:  tIpList{net.ParseIP("192.168.7.7")},
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "domain.tld",
					tIpList{net.ParseIP("192.168.1.7")}, 0)
				return tl
			}(),
			want: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "domain.tld",
					tIpList{net.ParseIP("192.168.7.7")}, 0)
				return tl
			}(),
		},
		{
			name: "08 - update sub.domain.tld, existing",
			host: "sub.domain.tld",
			ips:  tIpList{net.ParseIP("192.168.8.8")},
			tl: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "sub.domain.tld",
					tIpList{net.ParseIP("192.168.1.8")}, 0)
				return tl
			}(),
			want: func() *TTrieList {
				tl := newTrie()
				tl.Create(context.TODO(), "sub.domain.tld",
					tIpList{net.ParseIP("192.168.8.8")}, 0)
				return tl
			}(),
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.tl.Update(context.TODO(), tc.host, tc.ips, 0)

			if nil == got {
				if nil != tc.want {
					t.Error("TTrieList.Update() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("TTrieList.Update() =\n%q\nwant 'nil'",
					got)
				return
			}

			gotTrie := got.(*TTrieList)
			wantTrie := tc.want.(*TTrieList)
			if !wantTrie.Equal(gotTrie) {
				t.Errorf("TTrieList.Update() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_TTrieList_Update()

/* _EoF_ */
