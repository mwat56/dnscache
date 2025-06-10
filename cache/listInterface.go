/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"context"
	"net"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `iCacheList` is the basic interface for a cache list.
	// It provides a CRUD interface for caching hostname's IP addresses:
	//
	//   - `C`: Create a new hostname's data cache [Create],
	//   - `R`: Retrieve a hostname's cached data [IPs],
	//   - `U`: Update a hostname's cached data [Update],
	//   - `D`: Delete a hostname's cached data [Delete].
	iCacheList interface {

		// `Create()` adds a new cache entry for the given hostname.
		//
		// Parameters:
		//   - `context.Context`: Timeout context to use for the operation.
		//   - `string`: The hostname to add a cache entry for.
		//   - `[]net.IP`: List of IP addresses to add to the cache entry.
		//   - `time.Duration`: Time to live for the hostname's cache entry.
		//
		// Returns:
		//   - `iCacheList`: The updated cache list.
		Create(context.Context, string, []net.IP, time.Duration) iCacheList

		// `Delete()` removes a hostname pattern from the node's trie.
		//
		// The method returns `true` if at least one part of the
		// hostname's path was deleted, `false` otherwise.
		//
		// Parameters:
		//   - `context.Context`: Timeout context to use for the operation.
		//   - `string`: The hostname to remove from the cache.
		//
		// Returns:
		//   - `bool`: `true` if a node was deleted, `false` otherwise.
		Delete(context.Context, string) bool

		// `IPs()` returns the IP addresses for the given hostname.
		//
		// Parameters:
		//   - `context.Context`: Timeout context to use for the operation.
		//   - `string`: The hostname to lookup in the cache.
		//
		// Returns:
		//   - `[]net.IP`: List of IP addresses for the given hostname.
		//   - `bool`: `true` if the hostname was found in the cache, `false` otherwise.
		IPs(context.Context, string) ([]net.IP, bool)

		// `Update()` updates the cache entry with the given IP addresses.
		//
		// Parameters:
		//   - `context.Context`: Timeout context to use for the operation.
		//   - `string`: The hostname to lookup in the cache.
		//   - `[]net.IP`: List of IP addresses to update the cache entry with.
		//   - `time.Duration`: Time to live for the cache entry.
		//
		// Returns:
		//   - `iCacheNode`: The updated cache node.
		Update(context.Context, string, []net.IP, time.Duration) iCacheList
	}
)

/* _EoF_ */
