/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package dnscache

import (
	"context"
	"reflect"
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_getDNSServers(t *testing.T) {
	tests := []struct {
		name    string
		want    []string
		wantErr bool
	}{
		{
			name: "multiple DNS servers found",
			want: []string{
				"9.9.9.9",         // Quad9
				"81.169.163.106",  // service-rns30.rz-ip.net.
				"85.214.7.22",     // service-rns20.rz-ip.net.
				"212.227.123.16",  // rec1.svc.1u1.it.
				"212.227.123.17"}, // rec2.svc.1u1.it.
			wantErr: false,
		},

		// TODO: Add test cases.
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := getDNSServers()
			if (err != nil) != tc.wantErr {
				t.Errorf("getDNSServers() error = '%v', wantErr '%v'",
					err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("getDNSServers() = %v, want %v", got, tc.want)
			}
		})
	}
} // Test_getDNSServers()

func Test_lookupDNS(t *testing.T) {
	type testCase struct {
		name     string
		server   string
		hostname string
		wantIPs  []string
		wantErr  bool
	}

	tests := []testCase{
		{
			name:   "multiple entries with valid hosts",
			server: "8.8.8.8",
			// Use real domains that should resolve successfully
			hostname: "example.com",
			wantIPs: []string{
				"2600:1406:bc00:53::b81e:94ce",
				"2600:1406:bc00:53::b81e:94c8",
				"2600:1406:3a00:21::173e:2e66",
				"2600:1408:ec00:36::1736:7f31",
				"2600:1408:ec00:36::1736:7f24",
				"2600:1406:3a00:21::173e:2e65",
				"23.192.228.80",
				"23.215.0.136",
				"96.7.128.198",
				"96.7.128.175",
				"23.192.228.84",
				"23.215.0.138"},
			wantErr: false,
		},
		{
			name:     "entries with invalid hosts",
			server:   "192.168.0.1",
			hostname: "invalid.example.nonexistent",
			wantIPs:  nil,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := lookupDNS(context.TODO(), tc.server, tc.hostname)

			// Check error
			if (nil != err) != tc.wantErr {
				t.Errorf("lookupDNS() error = '%v', wantErr '%v'",
					err, tc.wantErr)
				return
			}

			// Check IPs
			if nil == tc.wantIPs {
				if nil != got {
					t.Errorf("Expected nil IPs, got: %v",
						got)
				}
				return
			}
			assertIps(t, got, tc.wantIPs) // defined in `dnscache_test.go`
		})
	}
} // Test_TResolver_Refresh()

/* _EoF_ */
