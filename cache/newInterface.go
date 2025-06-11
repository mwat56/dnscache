/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import "time"

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `TCacheType` is the type of cache list to create.
	TCacheType int8
)

const (
	// `CacheTypeMap` is a map-based cache list.
	CacheTypeMap = TCacheType(-1)

	// `CacheTypeTrie` is a trie-based cache list.
	CacheTypeTrie = TCacheType(1)

	// `DefaultCacheSize` is the initial size of the cache list.
	DefaultCacheSize = 1 << 10 // 1024

	// `DefaultTTL` is the default time to live for a DNS cache entry.
	DefaultTTL = time.Duration(time.Minute << 9) // ~8 hours
)

// ---------------------------------------------------------------------------
// `ICacheList` constructor:

// `New()` returns a new IP address cache list.
//
// There are currently two types of cache lists available:
//   - `CacheTypeMap`: A map-based cache list,
//   - `CacheTypeTrie`: A trie-based cache list.
//
// The cache type is determined by the `aType` parameter.
// The trie-based cache is the default.
//
// The `aSize` argument is relevant only for the map-based cache list.
// If the value is zero, the default size (`1024`) is used.
//
// Parameters:
//   - `aType`: Type of cache to create.
//   - `aSize`: Initial size of the map-based cache.
//
// Returns:
//   - `ICacheList`: A new IP address cache list.
func New(aType TCacheType, aSize uint) ICacheList {
	if CacheTypeMap == aType {
		return newMap(aSize)
	}

	return newTrie()
} // New()

/* _EoF_ */
