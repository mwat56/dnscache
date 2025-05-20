/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"io"
	"os"
	"strings"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `tTrie` is a thread-safe trie for FQDN wildcards. It
	// basically provides a CRUD interface for FQDN patterns.
	//
	//   - `C`: Create a new pattern [Add],
	//   - `R`: Retrieve a pattern [Match],
	//   - `U`: Update a pattern [Update],
	//   - `D`: Delete a pattern [Delete].
	tTrie struct {
		root *tNode
	}
)

// ---------------------------------------------------------------------------
// `tTrie` constructor:

// `newTrie()` creates a new `tTrie` instance.
//
// Returns:
//   - `*tTrie`: A new `tTrie` instance.
func newTrie() *tTrie {
	return &tTrie{root: newNode()}
} // newTrie()

// ---------------------------------------------------------------------------
// `tTrie` methods:

// `Add()` inserts an FQDN pattern (with optional wildcard) into the list.
//
// If `aPattern` is an empty string, the method returns `false`.
//
// Parameters:
//   - `aPattern`: The FQDN pattern to insert.
//
// Returns:
//   - `rOK`: `true` if the pattern was added, `false` otherwise.
func (t *tTrie) Add(aPattern string) (rOK bool) {
	if nil == t || nil == t.root {
		return
	}

	parts := pattern2parts(aPattern)
	if 0 == len(parts) {
		return
	}

	t.root.Lock()
	rOK = t.root.add(parts)
	t.root.Unlock()

	return
} // Add()

// `AllPatterns()` returns all patterns in the trie.
//
// Returns:
//   - `rList`: A list of all patterns in the trie.
func (t *tTrie) AllPatterns() (rList tPartsList) {
	if nil == t || nil == t.root || (0 == len(t.root.tChildren)) {
		return
	}

	t.root.RLock()
	rList = t.root.allPatterns()
	t.root.RUnlock()

	return
} // AllPatterns()

// `clone()` returns a deep copy of the trie.
//
// Returns:
//   - `*tTrie`: A deep copy of the trie.
func (t *tTrie) clone() *tTrie {
	clone := newTrie()
	clone.root = t.root.clone()

	return clone
} // clone()

// `Count()` returns the number of nodes and patterns in the trie.
//
// Returns:
//   - `rNodes`: The number of nodes in the trie.
//   - `rPatterns`: The number of patterns in the trie.
func (t *tTrie) Count() (rNodes, rPatterns int) {
	if nil == t || nil == t.root {
		return
	}

	t.root.RLock()
	rNodes, rPatterns = t.root.count()
	t.root.RUnlock()

	return
} // Count()

// `Delete()` removes a pattern (FQDN or wildcard) from the list.
//
// The method returns a boolean value indicating whether the pattern
// was found and deleted.
//
// If `aPattern` is an empty string, the method returns `false`.
//
// Parameters:
//   - `aPattern`: The pattern to remove.
//
// Returns:
//   - `bool`: `true` if the pattern was found and deleted, `false` otherwise.
func (t *tTrie) Delete(aPattern string) (rOK bool) {
	if nil == t || nil == t.root {
		return
	}

	parts := pattern2parts(aPattern) // reversed list of parts
	if 0 == len(parts) {
		return
	}

	// To delete an FQDN or wildcard entry from your trie-based list:
	//
	// - Traverse the trie using the reversed labels of the entry.
	// - Unmark the terminal node’s isEnd or isWild flags.
	// - Recursively prune nodes that are no longer needed (i.e.,
	// nodes that are not terminal and have no children).

	t.root.Lock()
	rOK = t.root.delete(parts)
	t.root.Unlock()

	return
} // Delete()

// `Equal()` checks whether the trie is equal to another one.
//
// Parameters:
//   - `aTrie`: The trie to compare with.
//
// Returns:
//   - `bool`: `true` if the trie is equal to the other one, `false` otherwise.
func (t *tTrie) Equal(aTrie *tTrie) bool {
	if nil == t {
		return (nil == aTrie)
	}
	if nil == aTrie {
		return false
	}
	if t == aTrie {
		return true
	}
	if nil == t.root {
		return (nil == aTrie.root)
	}
	if nil == aTrie.root {
		return false
	}

	t.root.RLock()
	result := t.root.Equal(aTrie.root)
	t.root.RUnlock()

	return result
} // Equal()

// `ForEach()` calls the given function for each node in the trie.
//
// Since all fields of the nodes in this trie are private, this method
// doesn't provide access to the node's data. Its only use from outside
// this package would be to gather statistics for example by calling
// a node's public methods like `Hits()` or `String()`.
//
// The given `aFunc()` is called in a locked R/O context for each node in
// the trie. That means that `aFunc()` can safely access the node's public
// methods like [Hits] or [String] while all of the node's internal fields
// remain private (i.e. inaccessible).
//
// Parameters:
//   - `aFunc`: The function to call for each node.
func (t *tTrie) ForEach(aFunc func(aNode *tNode)) {
	if nil == t || nil == t.root || nil == aFunc {
		return
	}

	t.root.RLock()
	t.root.forEach(aFunc)
	t.root.RUnlock()
} // ForEach()

// `Load()` reads hostname patterns (FQDN or wildcards) from the reader
// and inserts them into the list.
//
// The given reader is expected to return one pattern per line. The
// method ignores empty lines and comment lines (starting with `#` or
// `;`). No attempt is made to validate the patterns regardless of FQDN
// or wildcard syntax, neither are the patterns checked for invalid
// characters or invalid endings.
//
// The method returns an error, if any. If it returns an error, the
// loading process has encountered a problem while reading the patterns
// and the trie may have not loaded all patterns.
//
// Parameters:
//   - `aReader`: The reader to read the patterns from.
//
// Returns:
//   - `error`: `nil` if the patterns were read successfully, the error otherwise.
//     see [Store]
func (t *tTrie) Load(aReader io.Reader) error {
	if (nil == t) || (nil == t.root) || (nil == aReader) {
		return ErrNodeNil
	}

	t.root.Lock()
	err := t.root.load(aReader)
	t.root.Unlock()

	return err
} // Load()

// `Match()` checks if the given hostname matches any pattern in the list.
//
// If aHostname is an empty string, the method returns `false`.
//
// The given hostname is matched against the patterns in the list
// in a case-insensitive manner.
//
// Parameters:
//   - `aHostname`: The hostname to check.
//
// Returns:
//   - `rOK`: `true` if the hostname matches any pattern, `false` otherwise.
func (t *tTrie) Match(aHostname string) (rOK bool) {
	if nil == t || nil == t.root {
		return
	}
	parts := pattern2parts(aHostname)
	if 0 == len(parts) {
		return
	}

	t.root.RLock()
	rOK = t.root.match(parts, true)
	t.root.RUnlock()

	return
} // Match()

// `Store()` writes all patterns currently in the list to the writer,
// one per line.
//
// Parameters:
//   - `aWriter`: The writer to write the patterns to.
//
// Returns:
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
//     see [Load]
func (t *tTrie) Store(aWriter io.Writer) error {
	if (nil == t) || (nil == t.root) || (nil == aWriter) {
		return ErrNodeNil
	}

	t.root.RLock()
	err := t.root.store(aWriter)
	t.root.RUnlock()

	return err
} // Store()

func (t *tTrie) store2file(aFilename string) error {
	if (nil == t) || (nil == t.root) {
		return ErrListNil
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

	t.root.RLock()
	err = t.root.store(file)
	t.root.RUnlock()

	if nil != err {
		_ = os.Remove(tmpName)
	} else {
		// Replace `aFilename` if it exists
		_ = os.Rename(tmpName, aFilename)
	}

	return err
} // store2file()

// `String()` implements the `fmt.Stringer` interface for the trie.
//
// Returns:
//   - `string`: The string representation of the trie.
func (t *tTrie) String() (rStr string) {
	if (nil == t) || (nil == t.root) {
		return ErrNodeNil.Error()
	}
	// var builder strings.Builder

	t.root.RLock()
	// t.root.prepString(&builder, "Trie", 0)
	rStr = t.root.string("Trie")
	t.root.RUnlock()

	return
} // String()

// `Update()` replaces an old pattern with a new one.
//
// Parameters:
//   - `aOldPattern`: The old pattern to replace.
//   - `aNewPattern`: The new pattern to replace the old one with.
//
// Returns:
//   - `rOK`: `true` if the pattern was updated, `false` otherwise.
func (t *tTrie) Update(aOldPattern, aNewPattern string) (rOK bool) {
	if nil == t || nil == t.root {
		return
	}
	oldParts := pattern2parts(aOldPattern) // reversed list of parts
	if 0 == len(oldParts) {
		return
	}
	newParts := pattern2parts(aNewPattern)
	if 0 == len(newParts) {
		return
	}
	if oldParts.Equal(newParts) {
		return
	}

	t.root.Lock()
	rOK = t.root.update(oldParts, newParts)
	t.root.Unlock()

	return
} // Update()

/* _EoF_ */
