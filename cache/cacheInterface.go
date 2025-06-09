/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"context"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `iCacheNode` is the basic interface for a node in a cache trie.
	// It provides a CRUD interface for caching nodes:
	//
	//   - `C`: Create a new hostname's data cache [Create],
	//   - `R`: Retrieve a hostname's cached data [Retrieve],
	//   - `U`: Update a hostname's cached data [Update],
	//   - `D`: Delete a hostname's cached data [Delete].
	iCacheNode interface {
		// `Create()` inserts a hostname pattern into cache.
		//
		// The method returns `true` if at least one part of the
		// hostname pattern was added in order to have the whole
		// `tPartsList` present in the trie or found an existing
		// node that already represented the pattern.
		//
		// Parameters:
		//   - `context.Context`: The timeout context to use for the operation.
		//   - `tPartsList`: The list of parts of the hostname to add.
		//   - `tIpList`: List of IP addresses to store with the cache entry.
		//   - `time.Duration`: Time to live for the cache entry.
		//
		// Returns:
		//   - `bool`: `true` if a pattern was added, `false` otherwise.
		Create(context.Context, tPartsList, tIpList, time.Duration) bool

		// `Delete()` removes a hostname pattern from the node's trie.
		//
		// The method returns `true` if at least one part of the
		// hostname's path was deleted, `false` otherwise.
		//
		// Parameters:
		//   - `context.Context`: The timeout context to use for the operation.
		//   - `tPartsList`: List of parts of the hostname pattern to delete.
		//
		// Returns:
		//   - `bool`: `true` if a node was deleted, `false` otherwise.
		Delete(context.Context, tPartsList) bool

		// `Retrieve()` returns the IP addresses for the given hostname pattern.
		//
		// Parameters:
		//   - `context.Context`: The timeout context to use for the operation.
		//   - `tPartsList`: The list of parts of the hostname to use.
		//
		// Returns:
		//   - `tIpList`: The list of IP addresses for the given pattern.
		Retrieve(context.Context, tPartsList) tIpList

		// `Update()` updates the cache entry with the given IP addresses.
		//
		// Parameters:
		//   - `tIpList`: List of IP addresses to update the cache entry with.
		//   - `time.Duration`: Time to live for the cache entry.
		//
		// Returns:
		//   - `iCacheNode`: The updated cache node.
		Update(tIpList, time.Duration) iCacheNode
	}
)

/* _EoF_ */
