/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `TADlist` is a list of allow and deny patterns for FQDN hosts
	// and wildcards.
	TADlist struct {
		datadir string // directory for local storage
		allow   *tTrie
		deny    *tTrie
	}

	// `TADresult` is the result type of a test by [TADlist.Match].
	TADresult int8

	// `ADlistError` is a special error type for `TADlist` errors.
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
// Parameters:
//   - `aDataDir`: The directory to use for local storage.
//
// Returns:
//   - `*TADlist`: A new `TADlist` instance.
func NewADlist(aDataDir string) *TADlist {
	if aDataDir = strings.TrimSpace(aDataDir); 0 == len(aDataDir) {
		aDataDir = os.TempDir()
	}

	//
	//TODO: Check whether `aDataDir` exists and create it if necessary.
	//

	return &TADlist{
		datadir: aDataDir,
		allow:   newTrie(),
		deny:    newTrie(),
	}
} // NewADlist()

// ---------------------------------------------------------------------------
// `TADlist` methods:

// `add()` inserts a FQDN name/pattern (with optional wildcard) into the given list.
//
// This function is not exported, as it is only used internally by the
// `AddAllow()` and `AddDeny()` methods.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aPattern`: The FQDN name/pattern to insert.
//   - `aList`: The list to insert the pattern into.
//
// Returns:
//   - `bool`: `true` if the pattern was added, `false` otherwise.
func add(aCtx context.Context, aPattern string, aList *tTrie) bool {
	if nil == aList || (nil == aList.root.node) {
		return false
	}
	if aPattern = strings.TrimSpace(aPattern); 0 == len(aPattern) {
		return false
	}
	if nil != aCtx.Err() {
		return false
	}

	ctx, cancel := context.WithTimeout(aCtx, time.Second<<2)
	defer cancel() // Ensure cancel is called

	return aList.Add(ctx, aPattern)
} // add()

// `AddAllow()` inserts a FQDN name/pattern (with optional wildcard)
// into the allow list.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aPattern`: The FQDN name/pattern to insert.
//
// Returns:
//   - `bool`: `true` if the pattern was added, `false` otherwise.
func (adl *TADlist) AddAllow(aCtx context.Context, aPattern string) bool {
	if nil == adl {
		return false
	}

	ctx, cancel := context.WithTimeout(aCtx, time.Second<<2)
	defer cancel() // Ensure cancel is called

	if ok := add(aCtx, aPattern, adl.allow); !ok {
		return false
	}

	// Save the modified allow list
	filename, err := filepath.Abs(filepath.Join(adl.datadir, "allow.txt"))
	if nil != err {
		return false
	}

	go func(aCtx context.Context) {
		// Check for timeout or cancellation
		if err := aCtx.Err(); nil != err {
			return
		}

		_ = adl.allow.storeFile(aCtx, filename)
	}(ctx)
	time.Sleep(time.Millisecond << 9) // yield to the new goroutine

	return true
} // AddAllow()

// `AddDeny()` inserts a FQDN name/pattern (with optional wildcard) into
// the deny list.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aPattern`: The FQDN name/pattern to insert.
//
// Returns:
//   - `bool`: `true` if the pattern was added, `false` otherwise.
func (adl *TADlist) AddDeny(aCtx context.Context, aPattern string) bool {
	if nil == adl {
		return false
	}

	return add(aCtx, aPattern, adl.deny)
} // AddDeny()

// `deletePattern()` removes a FQDN name/pattern (with optional wildcard)
// from the given list.
//
// This function is not exported, as it is only used internally by the
// `DeleteAllow()` and `DeleteDeny()` methods.
//
// Parameters:
//   - `aPattern`: The FQDN name/pattern to remove.
//   - `aList`: The list to remove the pattern from.
//
// Returns:
//   - `bool`: `true` if the pattern was found and deleted, `false` otherwise.
func deletePattern(aPattern string, aList *tTrie) bool {
	if (nil == aList) || (nil == aList.root.node) {
		return false
	}

	if aPattern = strings.TrimSpace(aPattern); 0 == len(aPattern) {
		// An empty pattern can not be removed from the list.
		return false
	}

	return aList.Delete(context.TODO(), aPattern)
} // deletePattern()

// `DeleteAllow()` removes a FQDN name/pattern (with optional wildcard)
// from the allow list.
//
// Parameters:
//   - `aPattern`: The FQDN name/pattern to remove.
//
// Returns:
//   - `bool`: `true` if the pattern was found and deleted, `false` otherwise.
func (adl *TADlist) DeleteAllow(aPattern string) bool {
	if nil == adl {
		return false
	}

	return deletePattern(aPattern, adl.allow)
} // DeleteAllow()

// `DeleteDeny()` removes a FQDN name/pattern (with optional wildcard)
// from the deny list.
//
// Parameters:
//   - `aPattern`: The FQDN name/pattern to remove.
//
// Returns:
//   - `bool`: `true` if the pattern was found and deleted, `false` otherwise.
func (adl *TADlist) DeleteDeny(aPattern string) bool {
	if nil == adl {
		return false
	}

	return deletePattern(aPattern, adl.deny)
} // DeleteDeny()

// `Equal()` checks whether the two lists are equal.
//
// NOTE: This method is of nor practical use apart from unit-testing.
//
// Parameters:
//   - `aList`: The list to compare with.
//
// Returns:
//   - `rOK`: `true` if the lists are equal, `false` otherwise.
func (adl *TADlist) Equal(aList *TADlist) (rOK bool) {
	if nil == adl {
		return (nil == aList)
	}
	if nil == aList {
		return
	}
	if adl == aList {
		return true
	}
	if !adl.allow.Equal(aList.allow) {
		return
	}
	rOK = adl.deny.Equal(aList.deny)

	return
} // Equal()

// `LoadAllow()` reads hostname patterns (FQDN or wildcards) from
// `aFilename` and inserts them into the allow list.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aFilename`: The path/file name to read the 'allow' patterns from.
//
// Returns:
//   - `error`: An error in case of problems, or `nil` otherwise.
//
// see [LoadDeny], [StoreAllow]
func (adl *TADlist) LoadAllow(aCtx context.Context, aFilename string) (rErr error) {
	if nil == adl {
		return ErrListNil
	}
	if aFilename = strings.TrimSpace(aFilename); 0 == len(aFilename) {
		aFilename = "allow.txt"
	}
	aFilename = filepath.Join(adl.datadir, aFilename)
	if aFilename, rErr = filepath.Abs(aFilename); nil != rErr {
		return
	}

	if rErr = aCtx.Err(); nil != rErr {
		return
	}

	// Consider slow connections and low bandwidth
	ctx, cancel := context.WithTimeout(aCtx, time.Second<<4)
	defer cancel() // Ensure cancel is called

	rErr = adl.allow.loadLocal(ctx, aFilename)

	//TODO: See, whether the `datadir` property is `os.TempDir()` and
	// copy the allow file to a permanent location.

	return
} // LoadAllow()

var (
	// `filenameRE` is a regular expression for invalid characters
	// in a filename.
	filenameRE = regexp.MustCompile(`[^a-zA-Z0-9-_\.]`)
)

// `urlPath2Filename()` converts an URL path to a filename.
//
// The function replaces all invalid characters with underscores where
// "invalid" is everything that is not a letter, a digit, a dash or an
// underscore.
//
// Parameters:
//   - `aURL`: The URL to convert.
//
// Returns:
//   - `string`: The generated filename.
//   - `error`: An error in case of problems, or `nil` otherwise.
func urlPath2Filename(aURL string) (string, error) {
	if aURL = strings.TrimSpace(aURL); 0 == len(aURL) {
		return "", ErrInvalidUrl
	}

	url, err := url.Parse(aURL)
	if nil != err {
		return "", err
	}

	path := url.Path
	// Remove leading slash
	if 0 != len(path) && '/' == path[0] {
		path = path[1:]
	}
	if 0 == len(path) {
		return "", ErrInvalidUrl
	}

	// Replace invalid characters with underscores
	return string(filenameRE.ReplaceAll([]byte(path), []byte("_"))), nil
} // urlPath2Filename()

// `loadRemote()` downloads a file from the given URL and saves it in
// the specified directory with the given filename.
//
// Afterwards it reads hostname patterns (FQDN or wildcards) from the file
// and inserts them into the given list.
//
// This function is not exported, as it is only used internally by the
// `LoadDeny()` method.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aURL`: The URL to download the file from.
//   - `aDir`: The directory name to save the file in.
//   - `aList`: The list to add the patterns to.
//
// Returns:
//   - `error`: An error in case of problems, or `nil` otherwise.
func loadRemote(aCtx context.Context, aURL, aDir string, aList *tTrie) (rErr error) {
	// No need to check arguments as that is done by the calling method.
	if destUrl, err := url.Parse(aURL); nil == err {
		// Turn URL string into net.URL and check for validity
		if ("" == destUrl.Host) || ("" == destUrl.Scheme) {
			rErr = ErrInvalidUrl
			return
		}
		aURL = destUrl.String()
	} else {
		rErr = ErrInvalidUrl
		return
	}
	if aDir = strings.TrimSpace(aDir); 0 == len(aDir) {
		aDir = os.TempDir()
	}

	var filename string
	if filename, rErr = urlPath2Filename(aURL); nil != rErr {
		return
	}
	filename = filepath.Join(aDir, filename)
	if filename, rErr = filepath.Abs(filename); nil != rErr {
		return
	}

	if rErr = aCtx.Err(); nil != rErr {
		return
	}

	// Consider slow connections and low bandwidth
	ctx, cancel := context.WithTimeout(aCtx, time.Second<<2)
	defer cancel() // Ensure cancel is called

	rErr = aList.loadRemote(ctx, aURL, filename)

	return
} // loadRemote()

// `LoadDeny()` downloads a file from the given URL and saves it in
// the specified directory with the given filename.
//
// Afterwards it reads hostname patterns (FQDN or wildcards) from the
// file and inserts them into the deny list.
//
// If `aURL` is an empty string or the list itself is empty, the method
// returns an error.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aURL`: The URL to download the file from.
//
// Returns:
//   - `error`: An error in case of problems, or `nil` otherwise.
//
// see [LoadAllow], [StoreDeny]
func (adl *TADlist) LoadDeny(aCtx context.Context, aURL string) error {
	if nil == adl {
		return ErrListNil
	}

	//
	//TODO: Loop through all arguments and load them all into a single list.
	//

	if aURL = strings.TrimSpace(aURL); 0 == len(aURL) {
		return ErrInvalidUrl
	}

	return loadRemote(aCtx, aURL, adl.datadir, adl.deny)
} // LoadDeny()

// `Match()` checks whether the given hostname should be allowed or blocked.
//
// The method returns `ADallow` if the hostname is in the allow list,
// `ADdeny` if it is in the deny list, and `ADneutral` otherwise.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aHostPattern`: The hostname to check.
//
// Returns:
//   - `TADresult`: The result of the lookup.
func (adl *TADlist) Match(aCtx context.Context, aHostPattern string) TADresult {
	if nil == adl {
		return ADneutral
	}

	if aHostPattern = strings.TrimSpace(aHostPattern); 0 == len(aHostPattern) {
		return ADneutral
	}

	if nil != aCtx.Err() {
		return ADneutral
	}

	ctx, cancel := context.WithTimeout(aCtx, time.Second<<2)
	defer cancel() // Ensure cancel is called

	var (
		// `allowOK` and `denyOK` are used to store the results
		// of the concurrent lookups in the allow and deny lists.
		allowOK, denyOK atomic.Bool
		wg              sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		denyOK.Store(adl.deny.Match(ctx, aHostPattern))
		wg.Done()
	}()

	go func() {
		allowOK.Store(adl.allow.Match(ctx, aHostPattern))
		wg.Done()
	}()

	wg.Wait()

	// The allow list is usually shorter (and more specific) than the
	// block list. Hence we give it preference.
	if allowOK.Load() {
		return ADallow
	}
	if denyOK.Load() {
		return ADdeny
	}

	return ADneutral
} // Match()

// `store()` writes all patterns currently in the list to the file.
//
// This function is not exported, as it is only used internally by the
// `StoreAllow()` and `StoreDeny()` methods.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aDir`: The directory to store the file in.
//   - `aFilename`: The filename to store the file as.
//   - `aList`: The list to store.
//
// Returns:
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
func store(aCtx context.Context, aDir, aFilename string, aList *tTrie) (rErr error) {
	aFilename = filepath.Join(aDir, aFilename)
	if aFilename, rErr = filepath.Abs(aFilename); nil != rErr {
		return
	}
	if rErr = aCtx.Err(); nil != rErr {
		return
	}
	if 0 == len(aList.root.node.tChildren) {
		rErr = ErrListNil
		return
	}
	ctx, cancel := context.WithTimeout(aCtx, time.Second<<2)
	defer cancel() // Ensure cancel is called

	rErr = aList.storeFile(ctx, aFilename)

	return
} // store()

// `StoreAllow()` writes all patterns currently in the allow list to the file.
//
// If the list is empty, the method returns the `ErrListNil` error.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//
// Returns:
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
//     see [StoreDeny], [LoadAllow]
func (adl *TADlist) StoreAllow(aCtx context.Context) error {
	if nil == adl {
		return ErrListNil
	}

	return store(aCtx, adl.datadir, "allow.txt", adl.allow)
} // StoreAllow()

// `StoreDeny()` writes all patterns currently in the deny list to the file.
//
// If the list is empty, the method returns the `ErrListNil` error.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//
// Returns:
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
//     see [StoreAllow], [LoadDeny]
func (adl *TADlist) StoreDeny(aCtx context.Context) error {
	if nil == adl {
		return ErrListNil
	}

	return store(aCtx, adl.datadir, "deny.txt", adl.allow)
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

// `update()` replaces an old pattern with a new one in the given list.
//
// This function is not exported, as it is only used internally by the
// `UpdateAllow()` and `UpdateDeny()` methods.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aOldPattern`: The old pattern to replace.
//   - `aNewPattern`: The new pattern to replace the old one with.
//   - `aList`: The list to update the pattern in.
//
// Returns:
//   - `bool`: `true` if the pattern was updated, `false` otherwise.
func update(aCtx context.Context, aOldPattern, aNewPattern string, aList *tTrie) bool {
	if nil == aList {
		return false
	}
	if aOldPattern = strings.TrimSpace(aOldPattern); 0 == len(aOldPattern) {
		return false
	}
	if aNewPattern = strings.TrimSpace(aNewPattern); 0 == len(aNewPattern) {
		return false
	}
	if nil != aCtx.Err() {
		return false
	}

	ctx, cancel := context.WithTimeout(aCtx, time.Second<<2)
	defer cancel() // Ensure cancel is called

	return aList.Update(ctx, aOldPattern, aNewPattern)
} // update()

// `UpdateAllow()` replaces an old pattern with a new one in the allow list.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aOldPattern`: The old pattern to replace.
//   - `aNewPattern`: The new pattern to replace the old one with.
//
// Returns:
//   - `bool`: `true` if the pattern was updated, `false` otherwise.
func (adl *TADlist) UpdateAllow(aCtx context.Context, aOldPattern, aNewPattern string) bool {
	if nil == adl {
		return false
	}

	return update(aCtx, aOldPattern, aNewPattern, adl.allow)
} // UpdateAllow()

// `UpdateDeny()` replaces an old pattern with a new one in the deny list.
//
// Parameters:
//   - `aCtx`: The context to use for the operation.
//   - `aOldPattern`: The old pattern to replace.
//   - `aNewPattern`: The new pattern to replace the old one with.
//
// Returns:
//   - `bool`: `true` if the pattern was updated, `false` otherwise.
func (adl *TADlist) UpdateDeny(aCtx context.Context, aOldPattern, aNewPattern string) bool {
	if nil == adl {
		return false
	}

	return update(aCtx, aOldPattern, aNewPattern, adl.deny)
} // UpdateDeny()

/* _EoF_ */
