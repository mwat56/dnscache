/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package dnscache

import (
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_TMetrics_check(t *testing.T) {
	tests := []struct {
		name    string
		metrics *TMetrics
		correct bool
		want    bool
	}{
		{
			name: "consistent",
			metrics: &TMetrics{
				Lookups: 10,
				Hits:    7,
				Misses:  3,
				Retries: 2,
				Errors:  1,
				Peak:    8,
			},
			correct: false,
			want:    true,
		},
		{
			name:    "nil",
			metrics: nil,
			correct: true,
			want:    true,
		},
		{
			name:    "empty",
			metrics: &TMetrics{},
			correct: true,
			want:    true,
		},
		{
			name: "lookups < hits + misses",
			metrics: &TMetrics{
				Lookups: 10,
				Hits:    7,
				Misses:  4,
				Retries: 2,
				Errors:  1,
				Peak:    8,
			},
			correct: false,
			want:    false,
		},
		{
			name: "lookups < hits + misses (correct)",
			metrics: &TMetrics{
				Lookups: 10,
				Hits:    7,
				Misses:  4,
				Retries: 2,
				Errors:  1,
				Peak:    8,
			},
			correct: true,
			want:    false,
		},
		{
			name: "lookups > hits + misses (correct)",
			metrics: &TMetrics{
				Lookups: 10,
				Hits:    7,
				Misses:  2,
				Retries: 2,
				Errors:  1,
				Peak:    8,
			},
			correct: true,
			want:    false,
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.metrics.check(tc.correct); got != tc.want {
				t.Errorf("TMetrics.check() = %v, want %v", got, tc.want)
			}
		})
	}
} // Test_TMetrics_check()

func Test_TMetrics_clone(t *testing.T) {
	tests := []struct {
		name         string
		setup        func()
		wantRMetrics *TMetrics
	}{
		{
			name: "all zero",
			setup: func() {
				// Reset metrics
				gMetrics = new(TMetrics)
			},
			wantRMetrics: &TMetrics{
				Lookups: 0,
				Hits:    0,
				Misses:  0,
				Retries: 0,
				Errors:  0,
				Peak:    0,
			},
		},
		{
			name: "all non-zero",
			setup: func() {
				// Set non-zero metrics
				gMetrics = &TMetrics{
					Lookups: 10,
					Hits:    6,
					Misses:  4,
					Retries: 2,
					Errors:  1,
					Peak:    8,
				}
				// Increment metrics
				incMetricsFields(&gMetrics.Lookups, &gMetrics.Hits,
					&gMetrics.Lookups, &gMetrics.Misses,
					&gMetrics.Retries, &gMetrics.Errors)
				setMetricsFieldMax(&gMetrics.Peak, 9)
			},
			wantRMetrics: &TMetrics{
				Lookups: 12,
				Hits:    7,
				Misses:  5,
				Retries: 3,
				Errors:  2,
				Peak:    9,
			},
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			gotRMetrics := gMetrics.clone()
			if nil == gotRMetrics {
				t.Error("GetMetrics() = nil, want non-nil")
				return
			}

			if !tc.wantRMetrics.Equal(gotRMetrics) {
				t.Errorf("GetMetrics() = %v, want %v",
					gotRMetrics, tc.wantRMetrics)
			}
		})
	}
} // Test_TMetrics_clone()

func Test_TMetrics_String(t *testing.T) {
	tests := []struct {
		name    string
		metrics TMetrics
		want    string
	}{
		{
			name: "all zero",
			metrics: TMetrics{
				Lookups: 0,
				Hits:    0,
				Misses:  0,
				Retries: 0,
				Errors:  0,
				Peak:    0,
			},
			want: "Lookups: 0\nHits: 0\nMisses: 0\nRetries: 0\nErrors: 0\nPeak: 0\n",
		},
		{
			name: "all non-zero",
			metrics: TMetrics{
				Lookups: 10,
				Hits:    7,
				Misses:  3,
				Retries: 2,
				Errors:  1,
				Peak:    8,
			},
			want: "Lookups: 10\nHits: 7\nMisses: 3\nRetries: 2\nErrors: 1\nPeak: 8\n",
		},

		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.metrics.String(); got != tc.want {
				t.Errorf("TMetrics.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_TMetrics_String()

/* _EoF_ */
