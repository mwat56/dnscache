/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package main

import (
	"os"
	"path/filepath"
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_tCmdLineArgs_Equal(t *testing.T) {
	tests := []struct {
		name    string
		cmdLine *tCmdLineArgs
		other   *tCmdLineArgs
		wantOK  bool
	}{
		/* */
		{
			name:    "01 - nil cmdLine and other",
			cmdLine: nil,
			other:   nil,
			wantOK:  true,
		},
		{
			name:    "02 - nil cmdLine",
			cmdLine: nil,
			other:   &tCmdLineArgs{},
			wantOK:  false,
		},
		{
			name:    "03 - nil other",
			cmdLine: &tCmdLineArgs{},
			other:   nil,
			wantOK:  false,
		},
		{
			name:    "04 - equal",
			cmdLine: &tCmdLineArgs{},
			other:   &tCmdLineArgs{},
			wantOK:  true,
		},
		{
			name:    "05 - not equal (1)",
			cmdLine: &tCmdLineArgs{ConfigPathName: "config.json"},
			other:   &tCmdLineArgs{},
			wantOK:  false,
		},
		{
			name:    "06 - not equal (2)",
			cmdLine: &tCmdLineArgs{Port: 53},
			other:   &tCmdLineArgs{},
			wantOK:  false,
		},
		{
			name:    "07 - not equal (3)",
			cmdLine: &tCmdLineArgs{ConsoleMode: true},
			other:   &tCmdLineArgs{},
			wantOK:  false,
		},
		{
			name:    "08 - not equal (4)",
			cmdLine: &tCmdLineArgs{DaemonMode: true},
			other:   &tCmdLineArgs{},
			wantOK:  false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if gotOK := tc.cmdLine.Equal(tc.other); gotOK != tc.wantOK {
				t.Errorf("tCmdLineArgs.Equal() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_tCmdLineArgs_Equal()

func Test_parseCmdLineArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want tCmdLineArgs
	}{
		/* */
		{
			name: "01 - empty args",
			args: []string{},
			want: tCmdLineArgs{
				ConfigPathName: gConfigFile,
				Port:           53,
				ConsoleMode:    false,
				DaemonMode:     true, // set by sanity check
			},
		},
		{
			name: "02 - invalid config path",
			args: []string{"-config", "/path/to/config.json"},
			want: tCmdLineArgs{
				ConfigPathName: "/mnt/tmp/dnscache.json", // set by sanity check
				Port:           53,
				ConsoleMode:    false,
				DaemonMode:     true, // set by sanity check
			},
		},
		{
			name: "03 - port",
			args: []string{"-port", "8053"},
			want: tCmdLineArgs{
				ConfigPathName: gConfigFile,
				Port:           8053,
				ConsoleMode:    false,
				DaemonMode:     true, // set by sanity check
			},
		},
		{
			name: "04 - console mode",
			args: []string{"-console"},
			want: tCmdLineArgs{
				ConfigPathName: gConfigFile,
				Port:           53,
				ConsoleMode:    true,
				DaemonMode:     false,
			},
		},
		{
			name: "05 - daemon mode",
			args: []string{"-daemon"},
			want: tCmdLineArgs{
				ConfigPathName: gConfigFile,
				Port:           53,
				ConsoleMode:    false,
				DaemonMode:     true,
			},
		},
		{
			name: "06 - multiple flags",
			args: []string{"-config", "custom.json", "-port", "5353", "-console", "-daemon"},
			want: tCmdLineArgs{
				ConfigPathName: "/mnt/tmp/dnscache.json", // set by sanity check
				Port:           5353,
				ConsoleMode:    false, // disabled by sanity check
				DaemonMode:     true,
			},
		},
		{
			name: "07 - invalid port",
			args: []string{"-port", "invalid"},
			want: tCmdLineArgs{
				ConfigPathName: gConfigFile,
				Port:           53,
				ConsoleMode:    false,
				DaemonMode:     true, // set by sanity check
			},
		},
		{
			name: "08 - negative port",
			args: []string{"-port", "-1"},
			want: tCmdLineArgs{
				ConfigPathName: gConfigFile,
				Port:           53, // corrected by sanity check
				ConsoleMode:    false,
				DaemonMode:     true, // set by sanity check
			},
		},
		/* */
		{
			name: "09 - help request",
			args: []string{"--help"},
			want: tCmdLineArgs{
				ConfigPathName: "/mnt/tmp/dnscache.json", // set by sanity check
				Port:           53,
				ConsoleMode:    false,
				DaemonMode:     true, // set by sanity check
			},
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseCmdLineArgs(tc.args)
			if !got.Equal(&tc.want) {
				t.Errorf("parseCmdLineArgs() =\n%+v\nwant\n%+v",
					got, tc.want)
			}
		})
	}
} // Test_parseCmdLineArgs()

func Test_loadConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     tConfiguration
		wantErr  bool
	}{
		/* */
		{
			name:     "01 - load config",
			filename: filepath.Join(t.TempDir(), "config.json"),
			want: tConfiguration{
				DataDir:         "",
				CacheSize:       0,
				RefreshInterval: 0,
				TTL:             0,
			},
			wantErr: true,
		},
		{
			name: "02 - load config",
			filename: func() string {
				fName := filepath.Join(t.TempDir(), "config.json")
				f, _ := os.Create(fName)
				_, _ = f.WriteString(`{"DataDir": "` + os.TempDir() + `","CacheSize": 1024,"RefreshInterval": 10,"TTL": 20}`)
				_ = f.Close()
				return fName
			}(),
			want: tConfiguration{
				DataDir:         os.TempDir(),
				CacheSize:       1024,
				RefreshInterval: 10,
				TTL:             20,
			},
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := loadConfiguration(tc.filename)
			if (err != nil) != tc.wantErr {
				t.Errorf("loadConfiguration() error = '%v', wantErr '%v'",
					err, tc.wantErr)
				return
			}
			if !got.Equal(&tc.want) {
				t.Errorf("loadConfiguration() =\n%v\nwant\n%v",
					got, tc.want)
			}
		})
	}
} // Test_loadConfiguration()

func Test_saveConfiguration(t *testing.T) {
	type tArgs struct {
		aConfig   tConfiguration
		aFilename string
	}
	tmpDir := t.TempDir()
	tests := []struct {
		name    string
		args    tArgs
		wantErr bool
	}{
		/* */
		{
			name: "01 - save config",
			args: tArgs{
				aConfig: tConfiguration{
					DataDir:         filepath.Join(tmpDir, "testdata"),
					CacheSize:       1024,
					RefreshInterval: 0,
					TTL:             0,
				},
				aFilename: filepath.Join(tmpDir, "config.json"),
			},
			wantErr: false,
		},
		{
			name: "02 - save config",
			args: tArgs{
				aConfig: tConfiguration{
					DataDir:         filepath.Join(tmpDir, "testdata"),
					CacheSize:       1024,
					RefreshInterval: 10,
					TTL:             20,
				},
				aFilename: filepath.Join(tmpDir, "config.json"),
			},
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := saveConfiguration(tc.args.aConfig, tc.args.aFilename)
			if (err != nil) != tc.wantErr {
				t.Errorf("saveConfiguration() error = '%v', wantErr '%v'",
					err, tc.wantErr)
			}
		})
	}
} // Test_saveConfiguration()

func Test_tConfiguration_Equal(t *testing.T) {
	tests := []struct {
		name   string
		config *tConfiguration
		other  *tConfiguration
		want   bool
	}{
		/* */
		{
			name:   "01 - nil config and other",
			config: nil,
			other:  nil,
			want:   true,
		},
		{
			name:   "02 - nil config",
			config: nil,
			other:  &tConfiguration{},
			want:   false,
		},
		{
			name:   "03 - nil other",
			config: &tConfiguration{},
			other:  nil,
			want:   false,
		},
		{
			name:   "04 - equal",
			config: &tConfiguration{},
			other:  &tConfiguration{},
			want:   true,
		},
		{
			name:   "05 - not equal",
			config: &tConfiguration{DataDir: "testdata"},
			other:  &tConfiguration{},
			want:   false,
		},
		{
			name:   "06 - not equal (2)",
			config: &tConfiguration{CacheSize: 1024},
			other:  &tConfiguration{},
			want:   false,
		},
		{
			name:   "07 - not equal (3)",
			config: &tConfiguration{RefreshInterval: 10},
			other:  &tConfiguration{},
			want:   false,
		},
		{
			name:   "08 - not equal (4)",
			config: &tConfiguration{TTL: 20},
			other:  &tConfiguration{},
			want:   false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.config.Equal(tc.other)
			if got != tc.want {
				t.Errorf("tConfiguration.Equal() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_tConfiguration_Equal()

func Test_tConfiguration_String(t *testing.T) {
	tests := []struct {
		name   string
		config *tConfiguration
		want   string
	}{
		/* */
		{
			name:   "01 - nil config",
			config: nil,
			want:   "",
		},
		{
			name:   "02 - empty config",
			config: &tConfiguration{},
			want:   "{}",
		},
		{
			name:   "03 - config with data",
			config: &tConfiguration{DataDir: "testdata", CacheSize: 1024, RefreshInterval: 10, TTL: 20},
			want:   "{\n\t\"dataDir\": \"testdata\",\n\t\"cacheSize\": 1024,\n\t\"refreshInterval\": 10,\n\t\"ttl\": 20\n}",
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.config.String()
			if got != tc.want {
				t.Errorf("tConfiguration.String() =\n%q\nwant\n%q",
					got, tc.want)
			}
		})
	}
} // Test_tConfiguration_String()

/* _EoF_ */
