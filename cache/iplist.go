/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package cache

import (
	"fmt"
	"net"
	"slices"
	"strings"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `tIpList` is a list of IP addresses.
	tIpList []net.IP
)

// ---------------------------------------------------------------------------
// `tIpList` methods:

// `Equal()` checks whether the IP list is equal to the given one.
//
// Parameters:
//   - `aList`: List to compare with.
//
// Returns:
//   - `bool`: `true` if the lists are equal, `false` otherwise.
func (il tIpList) Equal(aList tIpList) bool {
	if nil == il {
		return nil == aList
	}
	if nil == aList {
		return false
	}
	if len(il) != len(aList) {
		return false
	}

	return slices.EqualFunc(il, aList, func(ip1, ip2 net.IP) bool {
		return ip1.Equal(ip2)
	})
} // Equal()

// `First()` returns the first IP address in the list.
//
// Returns:
//   - `net.IP`: First IP address in the list.
func (il tIpList) First() net.IP {
	if 0 == len(il) {
		return nil
	}

	return il[0]
} // First()

// `Len()` returns the number of IP addresses in the list.
//
// Returns:
//   - `int`: Number of IP addresses in the list.
func (il tIpList) Len() int {
	return len(il)
} // Len()

// `String()` implements the `fmt.Stringer` interface for a string
// representation of the IP list.
//
// Returns:
//   - `string`: String representation of the IP list.
func (il tIpList) String() string {
	lLen := len(il)
	switch lLen {
	case 0:
		return ""
	case 1:
		return il[0].String()
	}

	var builder strings.Builder
	for i := range lLen {
		if nil != il[i] {
			fmt.Fprint(&builder, il[i].String())
			if i < lLen-1 {
				fmt.Fprintf(&builder, " - ")
			}
		}
	}

	return builder.String()
} // String()

/* _EoF_ */
