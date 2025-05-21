/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_TMetrics_Equal(t *testing.T) {
	tests := []struct {
		name    string
		m       *TMetrics
		metrics *TMetrics
		want    bool
	}{
		{
			name:    "01 - nil",
			m:       nil,
			metrics: nil,
			want:    true,
		},
		{
			name:    "02 - nil and non-nil",
			m:       nil,
			metrics: &TMetrics{},
			want:    false,
		},
		{
			name:    "03 - non-nil and nil",
			m:       &TMetrics{},
			metrics: nil,
			want:    false,
		},
		{
			name:    "04 - equal",
			m:       &TMetrics{},
			metrics: &TMetrics{},
			want:    true,
		},
		{
			name: "05 - not equal",
			m: &TMetrics{
				PoolCreations: 1,
			},
			metrics: &TMetrics{},
			want:    false,
		},
		{
			name: "06 - not equal (2)",
			m: &TMetrics{
				PoolCreations: 1,
				PoolReturns:   1,
			},
			metrics: &TMetrics{},
			want:    false,
		},
		{
			name: "07 - not equal (3)",
			m: &TMetrics{
				PoolSize: 1,
			},
			metrics: &TMetrics{
				PoolCreations: 1,
				PoolReturns:   1,
			},
			want: false,
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.m.Equal(tc.metrics); got != tc.want {
				t.Errorf("TMetrics.Equal() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_TMetrics_Equal()

func Test_TMetrics_String(t *testing.T) {
	tests := []struct {
		name string
		m    *TMetrics
		want string
	}{
		{
			name: "01 - nil",
			m:    nil,
			want: "",
		},
		{
			name: "02 - empty",
			m:    &TMetrics{},
			want: "Pool.Creations: 0\nPool.Returns: 0\nPool.Size: 0\nTrie.Nodes: 0\nTrie.Patterns: 0\nTrie.Hits: 0\nTrie.Misses: 0\nTrie.Reloads: 0\nTrie.Retries: 0\n",
		},
		{
			name: "03 - non-empty",
			m: &TMetrics{
				PoolCreations: 1,
				PoolReturns:   2,
				PoolSize:      3,
				Nodes:         4,
				Patterns:      5,
				Hits:          6,
				Misses:        7,
				Reloads:       8,
				Retries:       9,
			},
			want: "Pool.Creations: 1\nPool.Returns: 2\nPool.Size: 3\nTrie.Nodes: 4\nTrie.Patterns: 5\nTrie.Hits: 6\nTrie.Misses: 7\nTrie.Reloads: 8\nTrie.Retries: 9\n",
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.m.String(); got != tc.want {
				t.Errorf("TMetrics.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_TMetrics_String()

func Test_TMetrics_clone(t *testing.T) {
	tests := []struct {
		name string
		m    *TMetrics
		want *TMetrics
	}{
		{
			name: "01 - nil",
			m:    nil,
			want: nil,
		},
		{
			name: "02 - empty",
			m:    &TMetrics{},
			want: &TMetrics{},
		},
		{
			name: "03 - non-empty",
			m: &TMetrics{
				PoolCreations: 1,
				PoolReturns:   2,
				PoolSize:      3,
				Nodes:         4,
				Patterns:      5,
				Hits:          6,
				Misses:        7,
				Reloads:       8,
				Retries:       9,
			},
			want: &TMetrics{
				PoolCreations: 1,
				PoolReturns:   2,
				PoolSize:      3,
				Nodes:         4,
				Patterns:      5,
				Hits:          6,
				Misses:        7,
				Reloads:       8,
				Retries:       9,
			},
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.m.clone()

			if nil == got {
				if nil != tc.want {
					t.Error("TMetrics.clone() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("TMetrics.clone() = %v, want nil",
					got)
				return
			}
			if !tc.want.Equal(got) {
				t.Errorf("TMetrics.clone() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_TMetrics_clone()

/* _EoF_ */
