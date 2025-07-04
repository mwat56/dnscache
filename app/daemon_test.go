/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_isInstanceRunning(t *testing.T) {
	// Save original PID file path and restore it after tests
	origPidFile := gPidFile
	defer func() {
		gPidFile = origPidFile
	}()

	tests := []struct {
		name        string
		setupFunc   func() error
		cleanupFunc func()
		want        bool
	}{
		/* */
		{
			name: "01 - no PID file exists",
			setupFunc: func() error {
				// Use a non-existent file
				gPidFile = filepath.Join(os.TempDir(), "nonexistent-pid-file.pid")
				return nil
			},
			cleanupFunc: func() {},
			want:        false,
		},
		{
			name: "02 - PID file with invalid content",
			setupFunc: func() error {
				gPidFile = filepath.Join(os.TempDir(), "invalid-pid-file.pid")
				return os.WriteFile(gPidFile, []byte("not-a-number"), 0600)
			},
			cleanupFunc: func() {
				os.Remove(gPidFile)
			},
			want: false,
		},
		/* */
		{
			name: "03 - PID file with non-existent process",
			setupFunc: func() error {
				gPidFile = filepath.Join(os.TempDir(), "nonexistent-process.pid")
				// Use a very high PID that's unlikely to exist
				return os.WriteFile(gPidFile, []byte("999999"), 0600)
			},
			cleanupFunc: func() {
				os.Remove(gPidFile)
			},
			want: false,
		},
		/* */
		{
			name: "04 - PID file with current process",
			setupFunc: func() error {
				gPidFile = filepath.Join(os.TempDir(), "current-process.pid")
				return os.WriteFile(gPidFile, fmt.Appendf(nil, "%d", os.Getpid()), 0600)
			},
			cleanupFunc: func() {
				os.Remove(gPidFile)
			},
			want: true,
		},
		{
			name: "05 - PID file with another process",
			setupFunc: func() error {
				gPidFile = filepath.Join(os.TempDir(), "another-process.pid")
				// Use a very high PID that's unlikely to exist
				return os.WriteFile(gPidFile, []byte("999998"), 0600)
			},
			cleanupFunc: func() {
				os.Remove(gPidFile)
			},
			want: false,
		},
		/* */
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.setupFunc(); nil != err {
				t.Fatalf("Setup failed: %v", err)
			}
			defer tc.cleanupFunc()

			got := isInstanceRunning()
			if got != tc.want {
				t.Errorf("isInstanceRunning() = %v, want %v", got, tc.want)
			}
		})
	}
} // Test_isInstanceRunning()

/* _EoF_ */
