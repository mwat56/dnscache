/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"slices"
	"strings"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `tPartsList` is a reversed list of a hostname's parts.
	tPartsList []string
)

// ---------------------------------------------------------------------------
// Helper function:

// `pattern2parts()` converts a hostname pattern to a reversed list of parts.
//
// The pattern is expected to be a valid FQDN or wildcard pattern, and it's
// not checked for validity. It is trimmed and converted to lower case for
// case-insensitive matching.
//
// Parameters:
//   - `aPattern`: The pattern to check and convert.
//
// Returns:
//   - `tPartsList`: The list of parts.
func pattern2parts(aPattern string) tPartsList {
	if aPattern = strings.TrimSpace(aPattern); 0 == len(aPattern) {
		return nil
	}

	parts := strings.Split(strings.ToLower(aPattern), ".")
	slices.Reverse(parts)

	return parts
} // pattern2parts()

// ---------------------------------------------------------------------------
// `tPartsList` methods:

// `Equal()` checks whether the parts list is equal to the given one.
//
// Parameters:
//   - `aList`: The parts list to compare with.
//
// Returns:
//   - `bool`: `true` if the lists are equal, `false` otherwise.
func (pl tPartsList) Equal(aList tPartsList) bool {
	if nil == pl {
		return (nil == aList)
	}
	if nil == aList {
		return false
	}

	return slices.Equal(pl, aList)
} // Equal()

// `Len()` returns the number of pattern parts in the list.
//
// Returns:
//   - `int`: Number of pattern parts in the list.
func (pl tPartsList) Len() int {
	return len(pl)
} // Len()

// `String()` implements the `fmt.Stringer` interface for a string
// representation of the parts list.
//
// The list is returned in the reversed order, i.e. the TLD is the first
// element.
//
// Returns:
//   - `string`: String representation of the parts list.
func (pl tPartsList) String() string {
	switch len(pl) {
	case 0:
		return ""
	case 1:
		return pl[0]
	default:
		return strings.Join(pl, ".")
	}
} // String()

/* _EoF_ */
