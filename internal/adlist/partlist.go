/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"fmt"
	"slices"
	"strings"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `tPartsList` is a reversed list of a hostname's parts.
	tPartsList []string
)

// ---------------------------------------------------------------------------
// `tPartsList` methods:

// `Equal()` checks whether the current parts list is equal to the given one.
//
// Parameters:
//   - `aPartsList`: The parts list to compare with.
//
// Returns:
//   - `bool`: `true` if the parts list is equal to the given one, `false` otherwise.
func (pl tPartsList) Equal(aPartsList tPartsList) bool {
	if nil == pl {
		return (nil == aPartsList)
	}
	if nil == aPartsList {
		return false
	}
	if len(pl) != len(aPartsList) {
		return false
	}

	return slices.Equal(pl, aPartsList)
} // Equal()

// `String()` implements the `fmt.Stringer` interface for the parts list.
//
// Returns:
//   - `string`: The string representation of the parts list.
func (pl tPartsList) String() string {
	if 0 == len(pl) {
		return ""
	}
	var builder strings.Builder

	fmt.Fprint(&builder, "[ ")
	for _, part := range pl {
		if 0 < len(part) {
			fmt.Fprintf(&builder, "'%s' ", part)
		}
	}
	fmt.Fprint(&builder, "]")

	return builder.String()
} // String()

/* _EoF_ */
