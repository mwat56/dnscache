/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package main

import (
	"encoding/json"
	"os"
	"slices"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// `configFile` is the configuration file path
	configFile = "dnscache-config.json"
)

type (
	// `tConfiguration` represents the DNS cache configuration
	tConfiguration struct {
		DNSServers      []string `json:"dnsServers,omitempty"`
		DataDir         string   `json:"dataDir,omitempty"`
		CacheSize       int      `json:"cacheSize,omitempty"`
		RefreshInterval uint8    `json:"refreshInterval,omitempty"`
		TTL             uint8    `json:"ttl,omitempty"`
	}
)

// ---------------------------------------------------------------------------
// Helper functions:

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
	if nil != err {
		return config, err
	}

	err = json.Unmarshal(data, &config)
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

	return (c.DataDir == aConfig.DataDir) &&
		(c.CacheSize == aConfig.CacheSize) &&
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

/* * /
// `validate()` checks and corrects the configuration data.
//
// Returns:
//   - `*tConfiguration`: The validated configuration.
func (c *tConfiguration) validate() *tConfiguration {
	if nil == c {
		return &tConfiguration{}
	}
	if dLen := len(c.DNSServers); 0 < dLen {
		clone := make([]string, dLen)
		copy(clone, c.DNSServers)
		for idx, server := range clone {
			if nil == net.ParseIP(server) {
				c.DNSServers = append(c.DNSServers[:idx], c.DNSServers[idx+1:]...)
			}
		}
		clone = nil
		c.DNSServers = slices.Clip(c.DNSServers)
	}
	if 0 == len(c.DataDir) {
		c.DataDir = os.TempDir()
	}
	if 0 >= c.CacheSize {
		c.CacheSize = 1 << 10
	}

	return c
} // validate()
/* */

/* _EoF_ */
