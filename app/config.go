/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

var (
	// `gMe` is the path/name of this program.
	gMe = func() (rStr string) {
		rStr, _ = filepath.Abs(os.Args[0])
		return
	}()

	// `gConfigFile` is the global configuration path/filename
	gConfigFile = func() string {
		name := "dnscache.json"
		// look in the current directory
		fName, _ := filepath.Abs("./" + name)
		if _, err := os.Stat(fName); nil == err {
			return fName
		}
		// look in the user's home directory
		fName, _ = filepath.Abs("~/." + name)
		if _, err := os.Stat(fName); nil == err {
			return fName
		}
		// look in the user's configuration directory
		fName, _ = filepath.Abs("~/.config/" + name)
		if _, err := os.Stat(fName); nil == err {
			return fName
		}
		// look in the system's configuration directory
		fName, _ = filepath.Abs("/etc/" + name)
		if _, err := os.Stat(fName); nil == err {
			return fName
		}

		// Fallback: system's tmp. directory
		return filepath.Join(os.TempDir(), name)
	}()
)

type (
	// `tCmdLineArgs` represents the possible command line arguments.
	tCmdLineArgs struct {
		ConfigPathName string // Path to configuration file
		Address        string // IP address to bind to for DNS requests
		Port           int    // Port to listen on for DNS requests
		ConsoleMode    bool   // Run in console UI mode
		DaemonMode     bool   // Run as a daemon (Linux only)
	}

	// `tConfiguration` represents the DNS cache configuration
	tConfiguration struct {
		DNSServers      []string `json:"dnsServers,omitempty"`
		Address         string   `json:"address,omitempty"`
		DataDir         string   `json:"dataDir,omitempty"`
		Forwarder       string   `json:"forwarder,omitempty"`
		CacheSize       int      `json:"cacheSize,omitempty"`
		Port            int      `json:"port,omitempty"`
		RefreshInterval uint8    `json:"refreshInterval,omitempty"`
		TTL             uint8    `json:"ttl,omitempty"`
	}
)

// ---------------------------------------------------------------------------
// Helper functions:

// `parseCmdLineArgs()` parses the command line arguments.
//
// Parameters:
//   - `aArgList`: The command line arguments to parse.
//
// Returns:
//   - `tCmdLineArgs`: The parsed command line argument values.
func parseCmdLineArgs(aArgList []string) (rArgs tCmdLineArgs) {

	//TODO: Use `getopts` package

	fs := flag.NewFlagSet(gMe, flag.ContinueOnError)
	fs.StringVar(&rArgs.ConfigPathName, "config", gConfigFile,
		"Path to configuration file")
	fs.BoolVar(&rArgs.ConsoleMode, "console", false,
		"Run in console UI mode")
	fs.BoolVar(&rArgs.DaemonMode, "daemon", false,
		"Run as a daemon (Linux only)")
	fs.StringVar(&rArgs.Address, "address", "",
		"IP address to bind to (empty for all interfaces)")
	fs.IntVar(&rArgs.Port, "port", 53,
		"Port to listen on for DNS requests")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n\tUsage: %s [OPTIONS]\n\n", os.Args[0])
		fs.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\n\tMost options can be set in an JSON config file to keep the command-line short ;-)\n\t")
		//os.Exit(0)
	}
	_ = fs.Parse(aArgList) // an error will result in default values being used

	// Some sanity checks:
	if 0 >= rArgs.Port {
		rArgs.Port = 53
	}
	if !rArgs.ConsoleMode {
		rArgs.DaemonMode = true
	}
	if rArgs.DaemonMode {
		rArgs.ConsoleMode = false
		// Check for root privileges
		if (1024 > rArgs.Port) && (0 < os.Getuid()) {
			fmt.Fprintf(os.Stderr, "\nWarning: Port %d requires root privileges!\n", rArgs.Port)
		}
	}

	// Check whether the config file exists:
	fName, _ := filepath.Abs(rArgs.ConfigPathName)
	if _, err := os.Stat(fName); nil != err {
		rArgs.ConfigPathName = gConfigFile
	} else {
		rArgs.ConfigPathName = fName
	}

	return
} // parseCmdLineArgs()

// `loadConfiguration()` reads the configuration from a file.
//
// Parameters:
//   - `aFilename`: The file to load from.
//
// Returns:
//   - `tConfiguration`: The loaded configuration.
//   - `error`: Error if any occurred.
func loadConfiguration(aFilename string) (tConfiguration, error) {
	var config tConfiguration

	data, err := os.ReadFile(aFilename) //#nosec G304
	if nil == err {
		err = json.Unmarshal(data, &config)
	}

	return config, err
} // loadConfiguration()

// `saveConfiguration()` writes the configuration to a file.
//
// Parameters:
//   - `aConfig`: The configuration data to save.
//   - `aFilename`: The file to save to.
//
// Returns:
//   - `error`: Error if any occurred.
func saveConfiguration(aConfig tConfiguration, aFilename string) error {
	data, err := json.MarshalIndent(aConfig, "", "\t")
	if nil != err {
		return err
	}

	tmpName := aFilename + "~"
	if _, err = os.Stat(tmpName); nil == err {
		_ = os.Remove(tmpName)
	}

	// Write to the temporary file
	err = os.WriteFile(tmpName, data, 0640) //#nosec G306
	if nil != err {
		// Remove the temporary file if writing failed
		_ = os.Remove(tmpName)

		//TODO: See whether we should retry a few times before giving up.

		return err
	}

	// Replace older configuration if it exists by the new one
	return os.Rename(tmpName, aFilename)
} // saveConfiguration()

// ---------------------------------------------------------------------------
// `tCmdLineArgs` methods:

func (c *tCmdLineArgs) Equal(aCmdLine *tCmdLineArgs) (rOK bool) {
	if nil == c {
		return (nil == aCmdLine)
	}
	if nil == aCmdLine {
		return
	}
	if c.ConfigPathName != aCmdLine.ConfigPathName {
		return
	}
	if c.Address != aCmdLine.Address {
		return
	}
	if c.Port != aCmdLine.Port {
		return
	}
	if c.ConsoleMode != aCmdLine.ConsoleMode {
		return
	}
	if c.DaemonMode != aCmdLine.DaemonMode {
		return
	}
	rOK = true

	return
} // Equal()

// ---------------------------------------------------------------------------
// `tConfiguration` methods:

// `Equal()` checks whether the configuration is equal to the given one.
//
// Parameters:
//   - `aConfig`: The configuration to compare with.
//
// Returns:
//   - `bool`: `true` if the configuration is equal to the given one, `false` otherwise.
func (c *tConfiguration) Equal(aConfig *tConfiguration) bool {
	if nil == c {
		return (nil == aConfig)
	}
	if nil == aConfig {
		return false
	}
	if !slices.Equal(c.DNSServers, aConfig.DNSServers) {
		return false
	}

	return (c.Address == aConfig.Address) &&
		(c.DataDir == aConfig.DataDir) &&
		(c.CacheSize == aConfig.CacheSize) &&
		(c.Forwarder == aConfig.Forwarder) &&
		(c.Port == aConfig.Port) &&
		(c.RefreshInterval == aConfig.RefreshInterval) &&
		(c.TTL == aConfig.TTL)
} // Equal()

// `String()` implements the `fmt.Stringer` interface for the
// configuration data.
//
// The method returns the configuration data as a JSON string.
//
// Returns:
//   - `string`: String representation of the configuration data.
func (c *tConfiguration) String() string {
	if nil == c {
		return ""
	}
	data, err := json.MarshalIndent(c, "", "\t")
	if nil == err {
		return string(data)
	}

	return err.Error()
} // String()

/* _EoF_ */
