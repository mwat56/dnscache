/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package nodepool

import (
	"testing"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_TPool_Clear(t *testing.T) {
	tests := []struct {
		name string
		pool *TPool
	}{
		/* */
		{
			name: "01 - clear nil pool",
			pool: nil,
		},
		{
			name: "02 - clear empty pool",
			pool: &TPool{},
		},
		{
			name: "03 - clear non-empty pool",
			pool: func() *TPool {
				p := &TPool{nodes: make(chan any, 1)}
				p.Put("node 03")
				return p
			}(),
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.pool.Clear()

			if nil == tc.pool {
				return
			}
			if 0 != len(tc.pool.nodes) {
				t.Errorf("Clear() = %d, want 0", len(tc.pool.nodes))
			}
		})
	}
} // Test_TPool_Clear()

func Test_TPool_Get(t *testing.T) {
	np := Init(func() any { return "nil" }, 1)
	tests := []struct {
		name    string
		pool    *TPool
		prepare func()
		want    string
		wantErr bool
	}{
		/* */
		{
			name:    "01 - get from nil pool",
			pool:    nil,
			prepare: nil,
			want:    "",
			wantErr: true,
		},
		{
			name:    "02 - get from non-initialised pool",
			pool:    &TPool{},
			prepare: nil,
			want:    "",
			wantErr: false,
		},
		{
			name:    "03 - empty initialised pool",
			pool:    np,
			prepare: np.Clear,
			want:    "nil",
			wantErr: false,
		},
		{
			name: "04 - get from empty initialised pool",
			pool: np,
			prepare: func() {
				np.Clear()
				np.Put("node 04")
			},
			want:    "node 04",
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if nil != tc.prepare {
				tc.prepare()
			}
			got, gotErr := tc.pool.Get()

			if (nil != gotErr) != tc.wantErr {
				t.Errorf("Get() error = '%v', wantErr '%v'",
					gotErr, tc.wantErr)
				return
			}
			if nil == got {
				if "" != tc.want {
					t.Errorf("Get() = nil, want %q", tc.want)
				}
				return
			}

			if gotStr, ok := got.(string); ok {
				if gotStr != tc.want {
					t.Errorf("Get() = %q, want %q", gotStr, tc.want)
				}
			} else if "" != tc.want {
				t.Errorf("Get() returned %T, want string %q", got, tc.want)
			}
		})
	}
} // Test_TPool_Get()

func Test_TPool_Metrics(t *testing.T) {
	np := Init(func() any { return "nil" }, 0)
	clear := func(aPool *TPool) {
		aPool.Clear()
	}
	np2 := &TPool{}
	np3 := &TPool{New: func() any { return "np3" }}

	tests := []struct {
		name    string
		pool    *TPool
		prepare func()
		want    *TPoolMetrics
		wantErr bool
	}{
		/* */
		{
			name: "01 - empty pool",
			pool: np,
			prepare: func() {
				clear(np)
			},
			want: &TPoolMetrics{
				Size:     0,
				Created:  0,
				Returned: 0,
			},
			wantErr: false,
		},
		{
			name: "02 - after creating one node",
			pool: np,
			prepare: func() {
				clear(np)
				_, _ = np.Get()
			},
			want: &TPoolMetrics{
				Size:     0,
				Created:  1,
				Returned: 0,
			},
			wantErr: false,
		},
		{
			name: "03 - after creating and returning one node",
			pool: np,
			prepare: func() {
				clear(np)
				node, _ := np.Get()
				np.Put(node)
			},
			want: &TPoolMetrics{
				Size:     1,
				Created:  1,
				Returned: 1,
			},
			wantErr: false,
		},
		{
			name: "04 - after creating multiple nodes",
			pool: np,
			prepare: func() {
				clear(np)
				for range 5 {
					_, _ = np.Get()
				}
			},
			want: &TPoolMetrics{
				Size:     0,
				Created:  5,
				Returned: 0,
			},
			wantErr: false,
		},
		{
			name: "05 - after creating and returning multiple nodes",
			pool: np,
			prepare: func() {
				clear(np)
				nodes := make([]any, 5)
				for i := range nodes {
					n, _ := np.Get()
					nodes[i] = n
				}
				for _, node := range nodes {
					np.Put(node)
				}
			},
			want: &TPoolMetrics{
				Size:     5,
				Created:  5,
				Returned: 5,
			},
			wantErr: false,
		},
		{
			name:    "06 - queuing metrics from nil pool",
			pool:    nil,
			want:    nil,
			wantErr: true,
		},
		{
			name: "07 - queuing metrics from non-initialised pool",
			pool: np2,
			want: &TPoolMetrics{
				Size:     0,
				Created:  0,
				Returned: 0,
			},
			wantErr: false,
		},
		{
			name: "08 - creating and returning multiple nodes in empty pool",
			pool: np3,
			want: &TPoolMetrics{
				Size:     448, // (512 / 8) * 7
				Created:  0,
				Returned: 512, // returned during `reset()`
			},
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if nil != tc.prepare {
				tc.prepare()
			}
			got, gotErr := tc.pool.Metrics()

			if (nil != gotErr) != tc.wantErr {
				t.Errorf("Metrics() error = '%v', wantErr '%v'",
					gotErr, tc.wantErr)
				return
			}
			if nil == got {
				if nil != tc.want {
					t.Error("Metrics() = nil, want non-nil")
				}
				return
			}
			if nil == tc.want {
				t.Errorf("Metrics() =\n%v\nwant 'nil'",
					got)
				return
			}
			if got.Created != tc.want.Created {
				t.Errorf("Metrics().Created = %v, want %v",
					got.Created, tc.want.Created)
			}

			if got.Returned != tc.want.Returned {
				t.Errorf("Metrics().Returned = %v, want %v",
					got.Returned, tc.want.Returned)
			}

			if got.Size != tc.want.Size {
				t.Errorf("Metrics().Size = %v, want %v",
					got.Size, tc.want.Size)
			}
		})
	}
} // Test_TPool_Metrics()

func Test_TPool_MetricsChannel(t *testing.T) {
	tests := []struct {
		name    string
		pool    *TPool
		wantErr bool
	}{
		/* */
		{
			name:    "01 - get metrics channel from nil pool",
			pool:    nil,
			wantErr: true,
		},
		{
			name:    "02 - get metrics channel from uninitialised pool",
			pool:    &TPool{},
			wantErr: false,
		},
		{
			name:    "03 - get metrics channel from empty pool",
			pool:    &TPool{New: func() any { return "nil" }},
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			metricsCh, gotErr := tc.pool.MetricsChannel()

			if nil == metricsCh {
				if !tc.wantErr {
					t.Error("MetricsChannel() = `nil`, want non-nil")
				}
				return
			}
			if nil == tc.pool {
				if nil != gotErr {
					t.Errorf("MetricsChannel() error = '%v', want nil",
						gotErr)
					return
				}
			}
			if (nil != gotErr) != tc.wantErr {
				t.Errorf("MetricsChannel() error = '%v', wantErr '%v'",
					gotErr, tc.wantErr)
				return
			}

			// Test that the channel works by adding a node to the pool
			if tc.wantErr {
				// Create a node and put it back to trigger metrics update
				node, _ := tc.pool.Get()
				tc.pool.Put(node)

				// Try to read from channel with timeout
				select {
				case metrics := <-metricsCh:
					if nil == metrics {
						t.Errorf("MetricsChannel() channel returned metric '%v', want non-nil",
							metrics)
						return
					}
					if 0 >= metrics.Size {
						t.Errorf("MetricsChannel() channel returned size %d, want > 0",
							metrics.Size)
					}

				case <-time.After(time.Millisecond << 8):
					t.Error(
						"MetricsChannel() channel didn't receive metrics update within timeout",
					)
				}
			}
		})
	}
} // Test_TPool_MetricsChannel()

func Test_TPool_Put(t *testing.T) {
	clear := func(aPool *TPool) {
		aPool.Clear()
	}
	np3 := &TPool{New: func() any { return "returned-node" }}
	np4 := Init(func() any { return "nil" }, 1)

	tests := []struct {
		name        string
		pool        *TPool
		prepare     func()
		wantMetrics *TPoolMetrics
		wantErr     bool
	}{
		/* */
		{
			name:    "01 - put node nil pool",
			pool:    nil,
			wantErr: true,
		},
		{
			name: "02 - put node uninitialised pool",
			pool: &TPool{},
			wantMetrics: &TPoolMetrics{
				Size:     1,
				Created:  0,
				Returned: 1,
			},
			wantErr: false,
		},
		{
			name: "03 - put node in empty pool",
			pool: np3,
			prepare: func() {
				clear(np3)
			},
			wantMetrics: &TPoolMetrics{
				Size:     449, // (512 / 8) * 7) + 1
				Created:  0,
				Returned: 513, // 512 + 1
			},
			wantErr: false,
		},
		{
			name: "04 - put node to pool",
			pool: np4,
			prepare: func() {
				clear(np4)
			},
			wantMetrics: &TPoolMetrics{
				Size:     1,
				Created:  0,
				Returned: 1,
			},
			wantErr: false,
		},
		{
			name: "05 - put multiple nodes to pool",
			pool: np4,
			prepare: func() {
				clear(np4)
				for range 3 {
					np4.Put("returned-node")
				}
			},
			wantMetrics: &TPoolMetrics{
				Size:     4, // 3 + 1
				Created:  0,
				Returned: 4, // 3 + 1
			},
			wantErr: false,
		},
		/* */
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if nil != tc.prepare {
				tc.prepare()
			}
			gotErr := tc.pool.Put("returned-node")

			if (nil != gotErr) != tc.wantErr {
				t.Errorf("Put() error = '%v', wantErr '%v'",
					gotErr, tc.wantErr)
				return
			}

			// Verify gotMetrics
			gotMetrics, err := tc.pool.Metrics()
			if (nil != gotErr) != tc.wantErr {
				t.Errorf("Metrics() error = '%v', wantErr '%v'",
					err, tc.wantErr)
				return
			}
			if nil == gotMetrics {
				if nil != tc.wantMetrics {
					t.Error("Metrics() = nil, want non-nil")
				}
				return
			}
			if nil == tc.wantMetrics {
				t.Errorf("Metrics() =\n%v\nwant 'nil'",
					gotMetrics)
				return
			}
			if gotMetrics.Created != tc.wantMetrics.Created {
				t.Errorf("After Put(), Metrics().Created = %v, want %d",
					gotMetrics.Created, tc.wantMetrics.Created)
			}
			if gotMetrics.Returned != tc.wantMetrics.Returned {
				t.Errorf("After Put(), Metrics().Returned = %v, want %d", gotMetrics.Returned, tc.wantMetrics.Returned)
			}
			if gotMetrics.Size != tc.wantMetrics.Size {
				t.Errorf("After Put(), Metrics().Size = %d, want %d",
					gotMetrics.Size, tc.wantMetrics.Size)
			}

			// Verify we can get the node back
			got2, _ := tc.pool.Get()
			if gotStr, ok := got2.(string); ok {
				if "returned-node" != gotStr {
					t.Errorf("After Put(), Get() = %q, want %q",
						gotStr, "returned-node")
				}
			} else {
				t.Errorf("After Put(), Get() returned %T, want string",
					got2)
			}
		})
	}
} // Test_TPool_Put()

/* _EoF_ */
