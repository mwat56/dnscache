/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"errors"
	"os"
	"strings"
	"sync"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `TADlist` is a list of allow and deny patterns for FQDN hosts
	// and wildcards.
	TADlist struct {
		allow *tTrie
		deny  *tTrie
	}

	// `TADresult` is the result type of a test by [TADlist.Match].
	TADresult int8

	// `TADlistError` is a special error type for `TADlist` errors.
	ADlistError struct {
		error
	}
)

const (
	// `ADallow` is the result of a test by [TADlist.Match].
	ADallow = TADresult(1)

	// `ADdeny` is the result of a test by [TADlist.Match].
	ADdeny = TADresult(-1)

	// `ADneutral` is the result of a test by [TADlist.Match].
	ADneutral = TADresult(0)
)

var (
	ErrListNil = ADlistError{errors.New("ADlist is nil")}
)

// ---------------------------------------------------------------------------
// `TADlist` constructor:

// `NewADlist()` returns a new `TADlist` instance.
//
// Returns:
//   - `*TADlist`: A new `TADlist` instance.
func NewADlist() *TADlist {
	return &TADlist{
		allow: newTrie(),
		deny:  newTrie(),
	}
} // NewADlist()

// ---------------------------------------------------------------------------
// `TADlist` methods:

// `AddAllow()` inserts an FQDN pattern (with optional wildcard)
// into the allow list.
//
// Parameters:
//   - `aPattern`: The FQDN pattern to insert.
//
// Returns:
//   - `*TADlist`: The updated list.
//   - `bool`: `true` if the pattern was added, `false` otherwise.
func (adl *TADlist) AddAllow(aPattern string) (*TADlist, bool) {
	if (nil == adl) || (0 == len(aPattern)) {
		return nil, false //TODO: special error type
	}
	if _, ok := adl.allow.Add(aPattern); !ok {
		return nil, false //TODO: special error type
	}

	return adl, true
} // AddAllow()

// `AddDeny()` inserts an FQDN pattern (with optional wildcard) into
// the deny list.
//
// Parameters:
//   - `aPattern`: The FQDN pattern to insert.
//
// Returns:
//   - `*TADlist`: The updated list.
//   - `bool`: `true` if the pattern was added, `false` otherwise.
func (adl *TADlist) AddDeny(aPattern string) (*TADlist, bool) {
	if (nil == adl) || (0 == len(aPattern)) {
		return nil, false //TODO: special error type
	}
	if _, ok := adl.deny.Add(aPattern); !ok {
		return nil, false //TODO: special error type
	}

	return adl, true
} // AddDeny()

// `deletePattern()` removes an FQDN pattern (with optional wildcard)
// from the given list.
//
// Parameters:
//   - `aPattern`: The FQDN pattern to remove.
//   - `aList`: The list to remove the pattern from.
//
// Returns:
//   - `bool`: `true` if the pattern was found and deleted, `false` otherwise.
func deletePattern(aPattern string, aList *tTrie) bool {
	if nil == aList || nil == aList.root {
		return false //TODO: special error type
	}

	if aPattern = strings.TrimSpace(aPattern); 0 == len(aPattern) {
		// An empty pattern can not be removed from the list.
		return false
	}
	_, ok := aList.Delete(aPattern)

	return ok
} // deletePattern()

// `DeleteAllow()` removes an FQDN pattern (with optional wildcard)
// from the allow list.
//
// Parameters:
//   - `aPattern`: The FQDN pattern to remove.
//
// Returns:
//   - `*TADlist`: The updated list.
//   - `bool`: `true` if the pattern was found and deleted, `false` otherwise.
func (adl *TADlist) DeleteAllow(aPattern string) (*TADlist, bool) {
	if nil == adl {
		return nil, false //TODO: special error type
	}
	if ok := deletePattern(aPattern, adl.allow); !ok {
		return nil, false //TODO: special error type
	}

	return adl, true
} // DeleteAllow()

// `DeleteDeny()` removes an FQDN pattern (with optional wildcard)
// from the deny list.
//
// Parameters:
//   - `aPattern`: The FQDN pattern to remove.
//
// Returns:
//   - `*TADlist`: The updated list.
//   - `bool`: `true` if the pattern was found and deleted, `false` otherwise.
func (adl *TADlist) DeleteDeny(aPattern string) (*TADlist, bool) {
	if nil == adl {
		return nil, false //TODO: special error type
	}
	if ok := deletePattern(aPattern, adl.deny); !ok {
		return nil, false //TODO: special error type
	}

	return adl, true
} // DeleteDeny()

// `loadFile()` reads hostname patterns (FQDN or wildcards) from the
// file and inserts them into the list.
//
// If `aFilename` does not exist, the method returns `nil` without error.
//
// The method ignores empty lines and comment lines (starting with `#` or
// `;`). No attempt is made to validate the patterns regardless of FQDN or
// wildcard syntax, neither are the patterns checked for invalid characters
// or invalid endings.
//
// Parameters:
//   - `aFilename`: The file to read the patterns from.
//   - `aList`: The list to insert the patterns into.
//
// Returns:
//   - `error`: `nil` if the patterns were read successfully, the error otherwise.
func loadFile(aFilename string, aList *tTrie) error {
	if (nil == aList) || (nil == aList.root) {
		return nil //TODO: special error type
	}
	if aFilename = strings.TrimSpace(aFilename); 0 == len(aFilename) {
		return nil
	}
	if _, err := os.Stat(aFilename); nil != err {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	file, err := os.Open(aFilename) //#nosec G304
	if nil != err {
		return err
	}

	_, err = aList.Load(file)
	_ = file.Close()

	return err
} // loadFile()

// `LoadAllow()` reads hostname patterns (FQDN or wildcards) from the
// file and inserts them into the allow list.
//
// If `aFilename` does not exist, the method returns `nil` without error.
//
// Parameters:
//   - `aFilename`: The file to read the 'allow' patterns from.
//
// Returns:
//   - `*TADlist`: The updated list.
//   - `error`: `nil` if the patterns were read successfully, the error otherwise.
//     see [LoadDeny], [StoreAllow]
func (adl *TADlist) LoadAllow(aFilename string) (*TADlist, error) {
	if nil == adl {
		return nil, nil //TODO: special error type
	}
	if err := loadFile(aFilename, adl.allow); nil != err {
		return nil, err
	}

	return adl, nil
} // LoadAllow()

// `LoadDeny()` reads hostname patterns (FQDN or wildcards) from the
// file and inserts them into the deny list.
//
// If `aFilename` does not exist, the method returns `nil` without error.
//
// Parameters:
//   - `aFilename`: The file to read the 'deny' patterns from.
//
// Returns:
//   - `*TADlist`: The updated list.
//   - `error`: `nil` if the patterns were read successfully, the error otherwise.
//     see [LoadAllow], [StoreDeny]
func (adl *TADlist) LoadDeny(aFilename string) (*TADlist, error) {
	if nil == adl {
		return nil, nil //TODO: special error type
	}
	if err := loadFile(aFilename, adl.deny); nil != err {
		return nil, err
	}

	return adl, nil
} // LoadDeny()

// `Match()` checks whether the given hostname should be allowed or blocked.
//
// The method returns `ADallow` if the hostname is in the allow list,
// `ADdeny` if it is in the deny list, and `ADneutral` otherwise.
//
// Parameters:
//   - `aHostname`: The hostname to check.
//
// Returns:
//   - `TADresult`: The result of the lookup.
func (adl *TADlist) Match(aHostname string) TADresult {
	if nil == adl {
		return ADneutral
	}
	if aHostname = strings.TrimSpace(aHostname); 0 == len(aHostname) {
		return ADneutral
	}

	var (
		// `allowOK` and `denyOK` are used to store the results
		// of the concurrent lookups in the allow and deny lists.
		allowOK, denyOK bool
		wg              sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		_, denyOK = adl.deny.Match(aHostname)
		wg.Done()
	}()

	go func() {
		_, allowOK = adl.allow.Match(aHostname)
		wg.Done()
	}()

	wg.Wait()

	// The allow list is usually shorter (and more specific) than the
	// block list. Hence we give it preference.
	if allowOK {
		return ADallow
	}
	if denyOK {
		return ADdeny
	}

	return ADneutral
} // Match()

// `store2file()` writes all patterns currently in the list to the file.
//
// If `aFilename` is an empty string or `aList` is `nil`, the function
// returns `nil` without error.
//
// The function uses a temporary file to write the patterns to, and then
// renames it to the target filename. This way, the target file is always
// either empty or contains a valid list of patterns. If `aFilename`
// already exists, it is replaced.
//
// Parameters:
//   - `aFilename`: The file to write the patterns to.
//   - `aList`: The list to write.
//
// Returns:
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
func store2file(aFilename string, aList *tTrie) error {
	if (nil == aList) || (nil == aList.root) {
		return nil
	}
	if aFilename = strings.TrimSpace(aFilename); 0 == len(aFilename) {
		return nil
	}

	tmpName := aFilename + "~"
	if _, err := os.Stat(tmpName); nil == err {
		_ = os.Remove(tmpName)
	}

	file, err := os.Create(tmpName) //#nosec G304
	if nil != err {
		return err
	}
	defer file.Close()

	err = aList.Store(file)
	if nil != err {
		_ = os.Remove(tmpName)
	} else {
		// Replace `aFilename` if it exists
		_ = os.Rename(tmpName, aFilename)
	}

	return err
} // store2file()

// `StoreAllow()` writes all patterns currently in the allow list to the file.
//
// If `aFilename` is an empty string, the method returns `nil` without error.
//
// Parameters:
//   - `aFilename`: The file to write the patterns to.
//
// Returns:
//   - `*TADlist`: The updated list.
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
//     see [StoreDeny], [LoadAllow]
func (adl *TADlist) StoreAllow(aFilename string) (*TADlist, error) {
	if nil == adl {
		return nil, nil //TODO: special error type
	}
	if err := store2file(aFilename, adl.allow); nil != err {
		return nil, err
	}

	return adl, nil
} // StoreAllow()

// `StoreDeny()` writes all patterns currently in the deny list to the file.
//
// If `aFilename` is an empty string, the method returns `nil` without error.
//
// Parameters:
//   - `aFilename`: The file to write the patterns to.
//
// Returns:
//   - `*TADlist`: The updated list.
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
//     see [StoreAllow], [LoadDeny]
func (adl *TADlist) StoreDeny(aFilename string) (*TADlist, error) {
	if nil == adl {
		return nil, nil //TODO: special error type
	}
	if err := store2file(aFilename, adl.deny); nil != err {
		return nil, err
	}

	return adl, nil
} // StoreDeny()

// `String()` returns a string representation of the list.
//
// Returns:
//   - `string`: String representation of the list.
func (adl *TADlist) String() string {
	if nil == adl {
		return ""
	}
	var builder strings.Builder

	builder.WriteString("Allow:\n")
	builder.WriteString(adl.allow.String())
	builder.WriteString("\nDeny:\n")
	builder.WriteString(adl.deny.String())

	return builder.String()
} // String()

// `UpdateAllow()` replaces an old pattern with a new one in the allow list.
//
// Parameters:
//   - `aOldPattern`: The old pattern to replace.
//   - `aNewPattern`: The new pattern to replace the old one with.
//
// Returns:
//   - `*TADlist`: The updated list.
//   - `bool`: `true` if the pattern was updated, `false` otherwise.
func (adl *TADlist) UpdateAllow(aOldPattern, aNewPattern string) (*TADlist, bool) {
	if nil == adl {
		return nil, false //TODO: special error type
	}
	if _, ok := adl.allow.Update(aOldPattern, aNewPattern); !ok {
		return nil, false //TODO: special error type
	}

	return adl, true
} // UpdateAllow()

// `UpdateDeny()` replaces an old pattern with a new one in the deny list.
//
// Parameters:
//   - `aOldPattern`: The old pattern to replace.
//   - `aNewPattern`: The new pattern to replace the old one with.
//
// Returns:
//   - `*TADlist`: The updated list.
//   - `bool`: `true` if the pattern was updated, `false` otherwise.
func (adl *TADlist) UpdateDeny(aOldPattern, aNewPattern string) (*TADlist, error) {
	if nil == adl {
		return nil, ErrListNil
	}
	if _, ok := adl.deny.Update(aOldPattern, aNewPattern); !ok {
		return nil, ADlistError{errors.New("pattern not found")}
	}

	return adl, nil
} // UpdateDeny()

/* _EoF_ */
