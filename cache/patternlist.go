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
	// `tPatternList` is a list of hostname patterns.
	tPatternList []string
)

// ---------------------------------------------------------------------------
// `tPatternList` methods:

// `Equal()` checks whether the parts list is equal to the given one.
//
// Parameters:
//   - `aList`: The parts list to compare with.
//
// Returns:
//   - `bool`: `true` if the lists are equal, `false` otherwise.
func (pl tPatternList) Equal(aList tPatternList) bool {
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
func (pl tPatternList) Len() int {
	return len(pl)
} // Len()

// `String()` implements the `fmt.Stringer` interface for a string
// representation of the parts list.
//
// Returns:
//   - `string`: String representation of the parts list.
func (pl tPatternList) String() string {
	switch len(pl) {
	case 0:
		return ""
	case 1:
		return pl[0]
	default:
		return strings.Join(pl, "\n")
	}
} // String()

/* _EoF_ */
