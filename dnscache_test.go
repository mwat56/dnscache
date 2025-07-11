/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package dnscache

import (
	"bytes"
	"context"
	"net"
	"slices"
	"testing"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_New(t *testing.T) {
	// Test with zero refresh interval
	r1 := New(0)
	if nil == r1 {
		t.Error("Expected non-nil resolver with zero refresh rate")
	} else if 3 != r1.retries {
		t.Errorf("Expected default retries to be 3, got %d",
			r1.retries)

	}

	// Test with positive refresh interval
	r2 := New(5)
	if nil == r2 {
		t.Error("Expected non-nil resolver with positive refresh rate")
	} else if 3 != r2.retries {
		t.Errorf("Expected default retries to be 3, got %d",
			r2.retries)
	}
} // Test_New()

func Test_NewWithOptions(t *testing.T) {
	customResolver := &net.Resolver{
		PreferGo: true,
	}
	tests := []struct {
		name    string
		options TResolverOptions
		check   func(*testing.T, *TResolver)
	}{
		/* */
		{
			name:    "01 - default options",
			options: TResolverOptions{},
			check: func(t *testing.T, r *TResolver) {
				if nil == r {
					t.Error("Expected non-nil resolver with default options")
					return
				}
				if nil == r.dnsServers {
					t.Error("Expected non-nil server list with default options")
					return
				}
				if 3 != r.retries {
					t.Errorf("Expected default retries to be 3, got %d",
						r.retries)
				}
			},
		},
		{
			name:    "02 - custom DNS servers",
			options: TResolverOptions{DNSservers: []string{"8.8.8.8", "8.8.4.4"}},
			check: func(t *testing.T, r *TResolver) {
				if nil == r {
					t.Error("Expected non-nil resolver with custom DNS servers")
					return
				}
				if nil == r.dnsServers {
					t.Error("Expected non-nil server list with custom DNS servers")
					return
				}
				if !slices.Equal(r.dnsServers, []string{"8.8.8.8", "8.8.4.4"}) {
					t.Errorf("Expected DNS servers to be ['8.8.8.8', '8.8.4.4'], got\n%v",
						r.dnsServers)
				}
			},
		},
		{
			name:    "03 - invalid DNS servers",
			options: TResolverOptions{DNSservers: []string{"234.567.890.1"}},
			check: func(t *testing.T, r *TResolver) {
				if nil == r {
					t.Error("Expected non-nil resolver with invalid DNS servers")
					return
				}
				if nil == r.dnsServers {
					t.Error("Expected non-nil server list with invalid DNS servers")
					return
				}
			},
		},
		{
			name:    "04 - custom refresh interval",
			options: TResolverOptions{RefreshInterval: 1},
			check: func(t *testing.T, r *TResolver) {
				if nil == r {
					t.Error("Expected non-nil resolver with custom refresh interval")
					return
				}
				if nil == r.dnsServers {
					t.Error("Expected non-nil server list with custom refresh interval")
					return
				}
			},
		},
		{
			name:    "05 - custom cache size",
			options: TResolverOptions{CacheSize: 128},
			check: func(t *testing.T, r *TResolver) {
				if nil == r {
					t.Error("Expected non-nil resolver with custom cache size")
					return
				}
				if nil == r.dnsServers {
					t.Error("Expected non-nil server list with custom cache size")
					return
				}

				testHost := "test.example.com"
				testIP := net.ParseIP("192.168.1.1")
				r.ICacheList.Create(context.TODO(), testHost, []net.IP{testIP}, 0)

				// Verify the element was added successfully
				if exists := r.ICacheList.Exists(context.TODO(), testHost); !exists {
					t.Error("Failed to add and retrieve element from cache with custom size")
				}

				// Clear the test data
				r.ICacheList.Delete(context.TODO(), testHost)
			},
		},
		{
			name:    "06 - custom resolver",
			options: TResolverOptions{Resolver: customResolver},
			check: func(t *testing.T, r *TResolver) {
				if nil == r {
					t.Error("Expected non-nil resolver with custom resolver")
					return
				}
				if customResolver != r.resolver {
					t.Error("Expected resolver to use custom resolver")
				}
			},
		},
		{
			name:    "07 - custom max retries",
			options: TResolverOptions{MaxRetries: 5},
			check: func(t *testing.T, r *TResolver) {
				if nil == r {
					t.Error("Expected non-nil resolver with custom max retries")
					return
				}
				if 5 != r.retries {
					t.Errorf("Expected retries to be 5, got %d",
						r.retries)
				}
			},
		},
		{
			name:    "08 - custom TTL",
			options: TResolverOptions{TTL: 10},
			check: func(t *testing.T, r *TResolver) {
				if nil == r {
					t.Error("Expected non-nil resolver with custom TTL")
					return
				}
				if time.Minute*10 != r.ttl {
					t.Errorf("Expected TTL to be 10 minutes, got %v",
						r.ttl)
				}
			},
		},
		/* */
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resolver := NewWithOptions(tc.options)
			tc.check(t, resolver)
		})
	}
} // Test_NewWithOptions()

func Test_TResolver_Fetch(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		setup    func(*TResolver)
		wantIPs  []string
		wantErr  bool
	}{
		{
			name:     "01 - fetch from cache",
			hostname: "cached.example.com",
			setup: func(r *TResolver) {
				r.ICacheList.Create(context.TODO(), "cached.example.com", []net.IP{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				}, 0)
			},
			wantIPs: []string{"192.168.1.1", "192.168.1.2"},
			wantErr: false,
		},
		{
			name:     "02 - fetch uncached (lookup)",
			hostname: "dnscache.ggl.io",
			setup:    func(r *TResolver) {},
			wantIPs:  []string{"3.33.165.172", "15.197.228.149"},
			wantErr:  false,
		},
		{
			name:     "03 - fetch invalid hostname",
			hostname: "invalid.end.of.universe",
			setup:    func(r *TResolver) {},
			wantIPs:  nil,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resolver := New(0)
			tc.setup(resolver)

			ips, err := resolver.Fetch(tc.hostname)

			// Check error
			if (nil != err) != tc.wantErr {
				t.Errorf("Fetch() error = '%v', wantErr '%v'",
					err, tc.wantErr)
				return
			}

			// Check IPs
			if nil == tc.wantIPs {
				if nil != ips {
					t.Errorf("Expected nil IPs, got: %v",
						ips)
				}
				return
			}
			assertIps(t, ips, tc.wantIPs)
		})
	}
} // Test_TResolver_Fetch()

func Test_TResolver_FetchFirstString(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		setup    func(*TResolver)
		want     string
		wantErr  bool
	}{
		{
			name:     "01 - fetch from cache",
			hostname: "cached.example.com",
			setup: func(r *TResolver) {
				r.ICacheList.Create(context.TODO(), "cached.example.com", []net.IP{
					net.ParseIP("192.168.1.1"),
				}, 0)
			},
			want:    "192.168.1.1",
			wantErr: false,
		},
		{
			name:     "02 - fetch uncached (lookup)",
			hostname: "dnscache.ggl.io",
			setup:    func(r *TResolver) {},
			want:     "3.33.165.172", // Assuming this is the first IP returned
			wantErr:  false,
		},
		{
			name:     "03 - fetch invalid hostname",
			hostname: "invalid.end.of.universe",
			setup:    func(r *TResolver) {},
			want:     "",
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resolver := New(0)
			tc.setup(resolver)

			got, err := resolver.FetchFirstString(tc.hostname)

			// Check error
			if (nil != err) != tc.wantErr {
				t.Errorf("FetchFirstString() error = '%v', wantErr '%v'",
					err, tc.wantErr)
				return
			}

			// For error cases, we don't need to check the result
			if tc.wantErr {
				return
			}

			// For the lookup case, we can't predict the exact IP
			// so we just check that we got a non-empty string
			if "fetch uncached (lookup)" == tc.name {
				if "" == got {
					t.Errorf("FetchFirstString() got empty string, want non-empty")
				}
				return
			}

			// For cached case, check exact match
			if got != tc.want {
				t.Errorf("FetchFirstString() got = %q, want %q",
					got, tc.want)
			}
		})
	}
} // Test_TResolver_FetchFirstString()

func Test_TResolver_FetchRandomString(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		setup    func(*TResolver)
		wantErr  bool
	}{
		{
			name:     "01 - fetch from cache",
			hostname: "cached.example.com",
			setup: func(r *TResolver) {
				r.ICacheList.Create(context.TODO(), "cached.example.com", []net.IP{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
				}, 0)
			},
			wantErr: false,
		},
		{
			name:     "02 - fetch uncached (lookup)",
			hostname: "dnscache.ggl.io",
			setup:    func(r *TResolver) {},
			wantErr:  false,
		},
		{
			name:     "03 - fetch invalid hostname",
			hostname: "invalid.end.of.universe",
			setup:    func(r *TResolver) {},
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resolver := New(0)
			tc.setup(resolver)

			got, err := resolver.FetchRandomString(tc.hostname)

			// Check error
			if (nil != err) != tc.wantErr {
				t.Errorf("FetchRandomString() error = '%v', wantErr '%v'",
					err, tc.wantErr)
				return
			}

			// For error cases, we don't need to check the result
			if tc.wantErr {
				return
			}

			// For non-error cases, check that we got a non-empty string
			if "" == got {
				t.Errorf("FetchRandomString() got empty string, want non-empty")
			}

			// For cached case, verify the result is one of the expected IPs
			if "fetch from cache" == tc.name {
				expected := []string{"192.168.1.1", "192.168.1.2"}
				if !slices.Contains(expected, got) {
					t.Errorf("FetchRandomString() got = %q, want one of '%v'",
						got, expected)
				}
			}
		})
	}
} // Test_TResolver_FetchRandomString()

func Test_TResolver_lookup(t *testing.T) {
	tests := []struct {
		name     string
		resolver *TResolver
		hostname string
		setup    func(*TResolver)
		wantIPs  []net.IP
		wantErr  bool
	}{
		/* */
		{
			name:     "01 - lookup cached",
			resolver: New(0),
			hostname: "cached.example.com",
			setup: func(r *TResolver) {
				// The cache isn't actually used in this execution path.
				r.ICacheList.Create(context.TODO(), "cached.example.com", []net.IP{
					net.ParseIP("192.168.1.1"),
				}, 0)
			},
			wantIPs: nil,
			wantErr: true,
		},
		{
			name:     "02 - lookup uncached (lookup)",
			resolver: New(0),
			hostname: "dnscache.ggl.io",
			setup:    func(r *TResolver) {},
			wantIPs: []net.IP{
				net.ParseIP("3.33.165.172"),
				net.ParseIP("15.197.228.149"),
			},
			wantErr: false,
		},
		{
			name:     "03 - lookup invalid hostname",
			resolver: New(0),
			hostname: "invalid.end.of.universe",
			setup:    func(r *TResolver) {},
			wantIPs:  nil,
			wantErr:  true,
		},
		/* */
		{
			name:     "04 - lookup with DNS servers",
			resolver: New(0),
			hostname: "dnscache.ggl.io",
			setup: func(r *TResolver) {
				r.dnsServers = []string{"8.8.8.8", "8.8.4.4"}
			},
			wantIPs: []net.IP{
				net.ParseIP("3.33.165.172"),
				net.ParseIP("15.197.228.149"),
			},
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(tc.resolver)
			gotIPs, err := tc.resolver.lookup(context.TODO(), tc.hostname)

			if (err != nil) != tc.wantErr {
				t.Errorf("TResolver.lookup() error = '%v', wantErr '%v'",
					err, tc.wantErr)
				return
			}
			if nil == gotIPs {
				if nil != tc.wantIPs {
					t.Error("TResolver.lookup() = nil, want non-nil")
				}
				return
			}
			if nil == tc.wantIPs {
				t.Errorf("TResolver.lookup() = %v, want 'nil'",
					gotIPs)
				return
			}

			// Sort for consistency
			slices.SortFunc(gotIPs, func(ip1, ip2 net.IP) int {
				return bytes.Compare(ip1, ip2)
			})

			if !slices.EqualFunc(gotIPs, tc.wantIPs,
				func(ip1, ip2 net.IP) bool {
					return ip1.Equal(ip2)
				}) {
				t.Errorf("TResolver.lookup() =\n%v\nwant\n%v",
					gotIPs, tc.wantIPs)
			}
		})
	}
} // Test_TResolver_lookup()

func Test_TResolver_Refresh(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*TResolver)
		validate func(*testing.T, *TResolver)
	}{
		{
			name: "01 - multiple entries with valid hosts",
			setup: func(r *TResolver) {
				// Use real domains that should resolve successfully
				r.ICacheList.Create(context.TODO(), "example.com", []net.IP{
					net.ParseIP("93.184.216.34"), // example.com's IP
				}, 0)
				r.ICacheList.Create(context.TODO(), "google.com", []net.IP{
					net.ParseIP("142.250.185.78"), // one of Google's IPs
				}, 0)
			},
			validate: func(t *testing.T, r *TResolver) {
				// After refresh, these entries should still exist
				// but might have different IPs due to DNS changes
				ips1, err := r.Fetch("example.com")
				if nil != err || 0 == len(ips1) {
					t.Errorf("Valid entry should be preserved: err=%v, ips=%v",
						err, ips1)
				}

				ips2, err := r.Fetch("google.com")
				if nil != err || 0 == len(ips2) {
					t.Errorf("Valid entry should be preserved: err=%v, ips=%v",
						err, ips2)
				}
			},
		},
		{
			name: "02 - entries with invalid hosts",
			setup: func(r *TResolver) {
				// Use a non-existent domain that should fail DNS lookup
				r.ICacheList.Create(context.TODO(), "invalid.example.nonexistent", []net.IP{
					net.ParseIP("192.168.1.1"),
				}, 0)
			},
			validate: func(t *testing.T, r *TResolver) {
				// After refresh, this entry should be removed
				r.RLock()
				exists := r.ICacheList.Exists(context.TODO(), "invalid.example.nonexistent")
				r.RUnlock()

				if exists {
					t.Errorf("Invalid entry should be removed from cache")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resolver := New(0)
			tc.setup(resolver)

			// Call Refresh
			resolver.Refresh()

			// Validate results
			tc.validate(t, resolver)
		})
	}
} // Test_TResolver_Refresh()

func assertIps(t *testing.T, actuals []net.IP, expected []string) {
	if len(actuals) != len(expected) {
		t.Errorf("Expecting %d ips, got %d\n%v\ninstead of:\n%v",
			len(expected), len(actuals), actuals, expected)
		return
	}

	for _, ip := range actuals {
		if !slices.Contains(expected, ip.String()) {
			t.Errorf("Unexpected IP: '%v', missing in '%v",
				ip, expected)
		}
	}
} // assertIps()

func Test_validateDNSServers(t *testing.T) {
	tests := []struct {
		name     string
		servers  []string
		expected []string
	}{
		{
			name:     "01 - empty list",
			servers:  []string{},
			expected: nil,
		},
		{
			name:     "02 - valid IPs",
			servers:  []string{"8.8.8.8", "1.1.1.1"},
			expected: []string{"8.8.8.8", "1.1.1.1"},
		},
		{
			name:     "03 - mixed valid and invalid",
			servers:  []string{"8.8.8.8", "invalid", "1.1.1.1"},
			expected: []string{"8.8.8.8", "1.1.1.1"},
		},
		{
			name:     "04 - all invalid",
			servers:  []string{"invalid1", "invalid2"},
			expected: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := validateDNSServers(tc.servers)
			if !slices.Equal(got, tc.expected) {
				t.Errorf("validateDNSServers() = %v, want %v",
					got, tc.expected)
			}
		})
	}
} // Test_validateDNSServers()

/* _EoF_ */
