/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package main

import (
	"context"
	"net"
	"testing"

	"github.com/mwat56/dnscache"
	"github.com/rivo/tview"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

// compareEntryLists compares two lists of cache entries regardless of order
func compareEntryLists(a, b []tCacheEntry) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps for easier comparison
	aMap := make(map[string][]net.IP)
	bMap := make(map[string][]net.IP)

	for _, entry := range a {
		aMap[entry.Hostname] = entry.IPs
	}

	for _, entry := range b {
		bMap[entry.Hostname] = entry.IPs
	}

	// Compare maps
	for hostname, ips := range aMap {
		bIPs, ok := bMap[hostname]
		if !ok {
			return false
		}

		if !compareIPLists(ips, bIPs) {
			return false
		}
	}

	return true
} // compareEntryLists()

// compareIPLists compares two lists of IP addresses regardless of order
func compareIPLists(a, b []net.IP) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps for easier comparison
	aMap := make(map[string]bool)

	for _, ip := range a {
		aMap[ip.String()] = true
	}

	// Check if all IPs in b are in a
	for _, ip := range b {
		if !aMap[ip.String()] {
			return false
		}
	}

	return true
} // compareIPLists()

func Test_getCacheEntries(t *testing.T) {
	// Create a mock cache list for testing
	mockCache := dnscache.New(0)
	// Add some test entries
	mockCache.ICacheList.Create(context.TODO(), "example.com", []net.IP{net.ParseIP("192.168.1.1")}, 0)
	mockCache.ICacheList.Create(context.TODO(), "test.com", []net.IP{net.ParseIP("10.0.0.1"), net.ParseIP("10.0.0.2")}, 0)

	tests := []struct {
		name    string
		state   *tAppState
		want    []tCacheEntry
		wantErr bool
	}{
		/* */
		{
			name:    "01 - nil state",
			state:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name: "02 - nil resolver",
			state: &tAppState{
				app:       tview.NewApplication(),
				pages:     tview.NewPages(),
				resolver:  nil,
				statusBar: tview.NewTextView(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "03 - valid state with entries",
			state: &tAppState{
				app:       tview.NewApplication(),
				pages:     tview.NewPages(),
				resolver:  mockCache,
				statusBar: tview.NewTextView(),
			},
			want: []tCacheEntry{
				{
					Hostname: "example.com",
					IPs:      []net.IP{net.ParseIP("192.168.1.1")},
				},
				{
					Hostname: "test.com",
					IPs:      []net.IP{net.ParseIP("10.0.0.1"), net.ParseIP("10.0.0.2")},
				},
			},
			wantErr: false,
		},
		/* */
		// TODO: Add more test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := getCacheEntries(tc.state)

			if (nil != err) != tc.wantErr {
				t.Errorf("getCacheEntries() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if nil == got {
				if nil != tc.want {
					t.Errorf("getCacheEntries() = nil, want %v", tc.want)
				}
				return
			}

			if nil == tc.want {
				t.Errorf("getCacheEntries() = %v, want nil", got)
				return
			}

			// Sort the results for consistent comparison
			// This is needed because map iteration order is not guaranteed
			if !compareEntryLists(got, tc.want) {
				t.Errorf("getCacheEntries() = %v, want %v", got, tc.want)
			}
		})
	}
} // Test_getCacheEntries()

/* _EoF_ */
