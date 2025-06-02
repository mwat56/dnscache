/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions
//lint:file-ignore ST1005 - I like capitalisation

type (
	// `ILoader` is the interface for a loader of allow/deny lists.
	ILoader interface {
		// `Load()` reads hostname patterns from the file and adds
		// them to the trie node.
		//
		// Parameters:
		//   - `aCtx`: The timeout context to use for the operation.
		//   - `aFilename`: The path/name to read the patterns from.
		//   - `aNode`: The trie node to add the patterns to.
		//
		// Returns:
		//   - `error`: `nil` if the patterns were read successfully, the error otherwise.
		Load(aCtx context.Context, aFilename string, aNode *tNode) error
	}

	// `ISaver` is the interface for a saver of allow/deny lists.
	ISaver interface {
		// `Save()` writes all patterns currently in the node to the writer,
		// one hostname pattern per line.
		//
		// Parameters:
		//   - `aCtx`: The timeout context to use for the operation.
		//   - `aWriter`: The writer to write the patterns to.
		//   - `aNode`: The trie node to write the patterns from.
		//
		// Returns:
		//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
		Save(aCtx context.Context, aWriter io.Writer, aNode *tNode) error
	}

	// `ILoaderSaver` is the interface for a loader and saver of
	// allow/deny lists.
	ILoaderSaver interface {
		ILoader
		ISaver
	}

	// `tABPLoader` is a loader of ABP filter lists.
	tABPLoader struct{}

	// `tHostsLoader` is a loader of text files in `hosts(5)` format.
	tHostsLoader struct{}

	/*
		// `tHostsSaver` is a saver for text files in `hosts(5)` format.
		tHostsSaver struct{}
	*/

	// `tSimpleLoader` is a loader of simple text files with one
	// hostname per line.
	tSimpleLoader struct{}

	// `tSimpleSaver` is a saver for simple text files with one
	// hostname per line.
	tSimpleSaver struct{}

	/*
		// `tHostsLoaderSaver` is a loader and saver for text files in `hosts(5)` format.
		tHostsLoaderSaver struct {
			tHostsLoader
			tHostsSaver
		}

		// `tSimpleLoaderSaver` is a loader and saver for simple text
		// files with one hostname per line.
		tSimpleLoaderSaver struct {
			tSimpleLoader
			tSimpleSaver
		}
	*/
)

var (
	// `ErrLoaderNil` is returned if a loader or a method's required
	// arguments is `nil`.
	ErrLoaderNil = ADlistError{errors.New("Loader, Reader, or Node is nil")}

	toplevelDomains []string
	toplevelOnce    sync.Once
)

func init() {
	toplevelOnce.Do(func() {
		// TODO: Check whether there's a local file
		localCopy := filepath.Join(os.TempDir(), "tlds-alpha-by-domain.txt")
		if fi, err := os.Stat(localCopy); nil == err {
			if fi.ModTime().After(time.Now().Add(-7 * 24 * time.Hour)) {
				// Use the local copy to avoid network traffic
				toplevelDomains = make([]string, 0, 1024)

				inFile, err := os.Open(localCopy) //#nosec G304
				if nil != err {
					goto loadFromNet
				}
				defer inFile.Close()

				scanner := bufio.NewScanner(inFile)
				for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())
					if (0 == len(line)) || ("#" == string(line[0])) {
						// Ignore empty or comment lines
						continue
					}
					toplevelDomains = append(toplevelDomains, line)
				}
				toplevelDomains = slices.Clip(toplevelDomains)
				return
			}
		}

	loadFromNet:
		// Get the list of top-level domains from IANA
		toplevelDomains = make([]string, 0, 1024)
		// We need this entries for the hosts file format and unit-tests
		toplevelDomains = append(append(toplevelDomains, "localdomain"), "localhost")

		resp, err := http.Get("https://data.iana.org/TLD/tlds-alpha-by-domain.txt")
		if nil != err {
			return
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if (0 == len(line)) || ("#" == string(line[0])) {
				// Ignore empty or comment lines
				continue
			}

			// The IANA file contains uppercase entries but
			// we're only using lowercase entries in this package
			toplevelDomains = append(toplevelDomains, strings.ToLower(line))
		}
		toplevelDomains = slices.Clip(toplevelDomains)

		// Write `topLevelDomains` to `localCopy` for later use.
		file, err := os.OpenFile(localCopy, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) //#nosec G304
		if nil != err {
			return
		}
		defer file.Close()

		for _, tld := range toplevelDomains {
			_, _ = file.WriteString(tld + "\n")
		}
	})
} // init()

// ---------------------------------------------------------------------------
// `tABPLoader` method:

// `processABPLine()` processes a single line from an ABP filter list.
//
// The function returns the processed pattern and a boolean indicating
// whether the line was processed successfully.
//
// The function ignores all lines that contain invalid characters or
// are not valid hostnames.
//
// Parameters:
//   - `aLine`: The line to process.
//
// Returns:
//   - `rPattern`: The processed pattern.
//   - `rOK`: `true` if the line was processed successfully, `false` otherwise.
func processABPLine(aLine string) (rPattern string, rOK bool) {
	// Convert domain wildcard syntax
	if strings.HasPrefix(aLine, "||") {
		aLine = "*." + aLine[2:]
	}

	// Remove surrounding markers
	if aLine = strings.Trim(aLine, "|^/"); 0 == len(aLine) {
		return
	}

	// Strip protocol if present
	if strings.Contains(aLine, "://") {
		parts := strings.SplitN(aLine, "://", 2)
		if 2 > len(parts) {
			return
		}
		if aLine = strings.TrimSpace(parts[1]); 0 == len(aLine) {
			return
		}
	}

	// Reject lines with paths
	if strings.Contains(aLine, "/") {
		return
	}

	// Reject lines with wildcards
	if ("*." != aLine[:2]) && strings.Contains(aLine, "*") {
		// Reject `*analytics*.js` or `*.host.*.domain.tld`
		return
	}

	// Extract hostname and port
	rPattern = strings.SplitN(aLine, ":", 2)[0]
	if rPattern = strings.TrimSuffix(rPattern, "."); 0 == len(rPattern) {
		return
	}

	// Reject lines with invalid characters
	if strings.ContainsAny(rPattern, " /?#@[]") {
		rPattern = "" // clear pattern (for unit-tests)
		return
	}
	rOK = true

	return
} // processABPLine()

// `Load()` reads hostname patterns from the file and adds them
// to the node's tree.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aFilename`: The path/name to read the hostnames from.
//   - `aNode`: The node to add the patterns to.
//
// Returns:
//   - `error`: `nil` if the patterns were read successfully, the error otherwise.
func (al *tABPLoader) Load(aCtx context.Context, aFilename string, aNode *tNode) error {
	if (nil == al) || (nil == aNode) || "" == aFilename {
		return ErrLoaderNil
	}

	// Open the downloaded file
	inFile, err := os.Open(aFilename) //#nosec G304
	if nil != err {
		return err
	}
	defer inFile.Close()

	added := 0
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		// Check for timeout or cancellation
		if err := aCtx.Err(); nil != err {
			return err
		}
		line := strings.TrimSpace(scanner.Text())
		if 0 == len(line) {
			// Ignore empty lines
			continue
		}
		if "!" == string(line[0]) {
			// Ignore comment lines
			continue
		}

		if 2 > len(line) {
			// Ignore ABP specific lines
			continue
		}

		// Ignore ABP specific lines
		switch string(line[0:2]) {
		case "@@", "##", "[A":
			continue

		default:
			if strings.Contains(line, "$") {
				continue
			}
		}

		if "/" == string(line[0]) {
			switch filepath.Ext(line[1:]) {
			case
				".action",
				".asp",
				".aspx",
				".css",
				".cookie",
				".faces",
				".gif",
				".gxt",
				".htm",
				".html",
				".jpg",
				".js",
				".jsf",
				".json",
				".jsp",
				".php",
				".png",
				".portlet",
				".servlet",
				".svg",
				".txt",
				".wicket",
				".xml":
				continue

			default:
				// Not a file extension we want to skip
			}
		}

		if pattern, ok := processABPLine(line); ok {
			// Split the pattern into multiple patterns if it
			// contains a `,` or '|` and process them separately
			entries := strings.Split(pattern, ",")
			for _, entry := range entries {
				if !(isValidHostname(entry) || isValidWildcard(entry)) {
					continue
				}
				if parts := pattern2parts(entry); 0 < len(parts) {
					if aNode.add(aCtx, parts) {
						added++
					}
				}
			}
		}
	}
	if 0 == added {
		return ADlistError{fmt.Errorf("no valid patterns found in %q", aFilename)}
	}

	return scanner.Err()
} // Load()

// ---------------------------------------------------------------------------
// `tHostsLoader` methods:

// `Load()` reads hostnames from the file and adds them to the node's tree.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aFilename`: The path/name to read the hostnames from.
//   - `aNode`: The node to add the hostnames's patterns to.
//
// Returns:
//   - `error`: `nil` if the patterns were read successfully, the error otherwise.
func (hl *tHostsLoader) Load(aCtx context.Context, aFilename string, aNode *tNode) error {
	if (nil == hl) || (nil == aNode) {
		return ErrLoaderNil
	}

	// Open the downloaded file
	inFile, err := os.Open(aFilename) //#nosec G304
	if nil != err {
		return err
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		// Check for timeout or cancellation
		if err := aCtx.Err(); nil != err {
			return err
		}
		line := strings.TrimSpace(scanner.Text())
		if 0 == len(line) {
			// Ignore empty lines
			continue
		}

		switch string(line[0]) {
		case "#", ";":
			// Ignore comment lines
			continue
		default:
			// Not a comment line
		}

		// Split the line into fields: We need at least two
		// fields (IP address and hostname).
		if fields := strings.Fields(line); 1 < len(fields) {
			// Check if the IP is valid
			if nil == net.ParseIP(fields[0]) {
				continue
			}
			// Check the remaining parts of the line
			fields = fields[1:]
			for idx := range fields {
				if parts := pattern2parts(fields[idx]); 0 < len(parts) {
					aNode.add(aCtx, parts)
				}
			}
		}
	}

	return scanner.Err()
} // Load()

/*
// ---------------------------------------------------------------------------
// `tHostsSaver` methods:

// `Save()` writes all patterns currently in the node to the writer,
// one hostname pattern per line.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aWriter`: The writer to write the patterns to.
//   - `aNode`: The node to write the patterns from.
//
// Returns:
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
func (hs *tHostsSaver) Save(aCtx context.Context, aWriter io.Writer, aNode *tNode) error {
	if (nil == hs) || (nil == aWriter) || (nil == aNode) {
		return ErrLoaderNil
	}

	// We need to prepend the IP address to each pattern line.
	// This is a bit of a hack, but it works.
	return aNode.store(aCtx, aWriter, "127.0.0.1")
} // Save()
*/

// ---------------------------------------------------------------------------
// `tSimpleLoader` methods:

// `Load()` reads hostname patterns from the file and adds them
// to the node's tree.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aFilename`: The path/name to read the hostnames from.
//   - `aNode`: The node to add the patterns to.
//
// Returns:
//   - `error`: `nil` if the patterns were read successfully, the error otherwise.
func (sl *tSimpleLoader) Load(aCtx context.Context, aFilename string, aNode *tNode) error {
	if (nil == sl) || (nil == aNode) {
		return ErrLoaderNil
	}

	// Open the downloaded file
	inFile, err := os.Open(aFilename) //#nosec G304
	if nil != err {
		return err
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		// Check for timeout or cancellation
		if err := aCtx.Err(); nil != err {
			return err
		}
		line := strings.TrimSpace(scanner.Text())
		if 0 == len(line) {
			// Ignore empty lines
			continue
		}

		switch string(line[0]) {
		case "#", ";":
			// Ignore comment lines
			continue
		default:
			// Not a comment line
		}

		if parts := pattern2parts(line); 0 < len(parts) {
			aNode.add(aCtx, parts)
		}
	}

	return scanner.Err()
} // Load()

// ---------------------------------------------------------------------------
// `tSimpleSaver` methods:

// `Save()` writes all patterns currently in the node to the writer,
// one hostname pattern per line.
//
// Parameters:
//   - `aCtx`: The timeout context to use for the operation.
//   - `aWriter`: The writer to write the patterns to.
//   - `aNode`: The node to write the patterns from.
//
// Returns:
//   - `error`: `nil` if the patterns were written successfully, the error otherwise.
func (ss *tSimpleSaver) Save(aCtx context.Context, aWriter io.Writer, aNode *tNode) error {
	if (nil == ss) || (nil == aWriter) || (nil == aNode) {
		return ErrLoaderNil
	}

	return aNode.store(aCtx, aWriter)
} // Save()

// ---------------------------------------------------------------------------
// File related functions:

type (
	// `tArchiveFormat` represents a supported archive format.
	tArchiveFormat string
)

const (
	ArchiveBZ2     tArchiveFormat = "bz2"
	ArchiveGZ      tArchiveFormat = "gz"
	ArchiveRAR     tArchiveFormat = "rar"
	ArchiveTAR     tArchiveFormat = "tar"
	ArchiveXZ      tArchiveFormat = "xz"
	ArchiveZIP     tArchiveFormat = "zip"
	Archive7Z      tArchiveFormat = "7z"
	ArchiveUnknown tArchiveFormat = ""
)

var (
	// `ErrInvalidDir` is returned if a given directory is invalid.
	ErrInvalidDir = ADlistError{errors.New("Directory is invalid")}

	// `ErrInvalidFile` is returned if a given filename is invalid.
	ErrInvalidFile = ADlistError{errors.New("Filename is invalid")}

	// `ErrInvalidUrl` is returned if a given URL is invalid.
	ErrInvalidUrl = ADlistError{errors.New("URL is invalid")}

	// `ErrUnknownArchive` is returned if a given archive format is unknown.
	ErrUnknownArchive = ADlistError{errors.New("Unknown archive format")}

	// `ErrUnknownFileType` is returned if a given file type is unknown.
	ErrUnknownFileType = ADlistError{errors.New("Unknown file type")}

	// `ErrUnknownMimeType` is returned if a given MIME type is unknown.
	ErrUnknownMimeType = ADlistError{errors.New("Unknown MIME type")}

	// `ErrUnsupportedArchive` is returned if a given archive format is not supported.
	ErrUnsupportedArchive = ADlistError{errors.New("Unsupported archive format")}

	// `ErrUnsupportedMime` is returned if a given MIME type is not supported.
	ErrUnsupportedMime = ADlistError{errors.New("Unsupported MIME type")}

	// // `textMimeTypes` lists known text-based MIME types.
	// textMimeTypes = []string{
	// 	"text/plain",
	// 	"text/x-hosts",
	// 	"text/x-hostnames",
	// }

	// `archiveSignatures` lists known archive magic numbers and their offsets.
	archiveSignatures = []struct {
		Signature []byte
		Offset    int64
		Format    tArchiveFormat
	}{
		{[]byte{0x50, 0x4B, 0x03, 0x04},
			0, ArchiveZIP}, // ZIP
		{[]byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x00},
			0, ArchiveRAR}, // RAR v1.5
		{[]byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x01, 0x00},
			0, ArchiveRAR}, // RAR v5.0
		{[]byte{0x1F, 0x8B},
			0, ArchiveGZ}, // GZIP
		{[]byte{0x42, 0x5A, 0x68},
			0, ArchiveBZ2}, // BZ2
		{[]byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00},
			0, ArchiveXZ}, // XZ
		{[]byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C},
			0, Archive7Z}, // 7z
		// TAR has no fixed magic at byte 0; see [isBinary]
	}

	// `validHostnameRE` is a regular expression for hostname validation per RFC 952/1123.
	validHostnameRE = regexp.MustCompile(`^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*)$`)
)

// `detectFileType()` detects the type of a file based on its magic number.
//
// Parameters:
//   - `aFilename`: The file to detect the type of.
//
// Returns:
//   - `string`: The MIME type of the file.
//   - `error`: `nil` if the file type was detected successfully, the error otherwise.
func detectFileType(aFilename string) (rMime string, rErr error) {
	inFile, err := os.Open(aFilename) //#nosec G304
	if nil != err {
		rErr = ADlistError{fmt.Errorf("Failed to open file %q: %v",
			aFilename, err)}
		return
	}
	defer inFile.Close()

	if format, err := isBinary(inFile); (nil == err) &&
		(ArchiveUnknown != format) {
		// Ok, so we're done with a known archive format
		rMime = "application/x-" + string(format)
		return
	}

	// Check for text-based formats
	if isText(inFile) {
		if isHostsFile(inFile) {
			rMime = "text/x-hosts"
			return
		}

		if isHostnamesOnly(inFile) {
			rMime = "text/x-hostnames"
			return
		}

		if isABPfile(inFile) {
			rMime = "text/x-abp"
			return
		}

		// Ok, so we're done with some text file
		rMime = "text/plain"
		return
	}
	// Unknown file type
	rErr = ErrUnknownFileType

	return
} // detectFileType()

// `downloadFile()` downloads a file from the given URL and saves it
// in the specified directory with the given filename.
//
// Parameters:
//   - `aURL`: The URL to download the file from.
//   - `aFilename`: The filename to save the data as.
//
// Returns:
//   - `rFilename`: The absolute path/name of the downloaded file.
//   - `rErr`: `nil` if the file was downloaded and saved successfully, the error otherwise.
func downloadFile(aURL, aFilename string) (rFilename string, rErr error) {
	if aURL = strings.TrimSpace(aURL); 0 == len(aURL) {
		rErr = ErrInvalidUrl
		return
	}
	if aFilename, rErr = filepath.Abs(aFilename); nil != rErr {
		return
	}
	dir := filepath.Dir(aFilename)
	if 0 == len(dir) {
		dir = os.TempDir()
	}

	// Ensure the directory exists. If `dir` is already a directory,
	// `MkdirAll` does nothing and returns nil.
	if rErr = os.MkdirAll(dir, 0750); nil != rErr {
		rErr = ADlistError{fmt.Errorf("Failed to create directory: %v", rErr)}
		return
	}

	//
	//TODO: Check whether we have a local cop< already
	//

	// Build a tmp. filename
	tmpName := aFilename + "~"
	if _, err := os.Stat(tmpName); nil == err {
		_ = os.Remove(tmpName)
	}

	// Request the file
	resp, err := http.Get(aURL) //#nosec G107
	if nil != err {
		rErr = ADlistError{fmt.Errorf("Failed to download file: %v", err)}
		return
	}
	defer resp.Body.Close()

	// First write to the temporary file and later rename
	// it to the final name if no errors occurred
	tmpFile, err := os.OpenFile(tmpName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) //#nosec G304
	if nil != err {
		rErr = ADlistError{fmt.Errorf("Failed to create temporary file: %v", err)}
		return
	}
	defer tmpFile.Close()

	// Copy the content.
	if _, err = io.Copy(tmpFile, resp.Body); nil != err {
		_ = os.Remove(tmpName)
		rErr = ADlistError{fmt.Errorf("Failed to save file: %v", err)}
		return
	}

	// Replace `tmpName` if it exists
	if rErr = os.Rename(tmpName, aFilename); nil != rErr {
		_ = os.Remove(tmpName)
		rErr = ADlistError{fmt.Errorf("Failed to rename file: %v", rErr)}
		return
	}
	rFilename = aFilename

	return
} // downloadFile()

// `isABPfile()` checks whether the given file is an ABP filter list.
//
// Parameters:
//   - `aFile`: The file data to check.
//
// Returns:
//   - `rOK`: `true` if the file is an ABP filter list, `false` otherwise.
func isABPfile(aFile io.ReadSeeker) (rOK bool) {
	if nil == aFile {
		return
	}
	if _, err := aFile.Seek(0, io.SeekStart); nil != err {
		return
	}

	abps := 0
	scanner := bufio.NewScanner(aFile)
	for scanner.Scan() {
		// 32 ABP specific lines should be enough to assume it's an ABP
		// file, even with up to 20 comment headers at the beginning
		if 32 < abps {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		// Ignore lines too short to be ABP specific
		if 2 > len(line) {
			// Skip w/o counting and checking
			continue
		}

		switch string(line[0]) {
		case "!", "/", "&", "-", ".", "[":
			// Count ABP specific line starts
			abps++
			continue
		default:
			// Not an ABP specific line start
		}

		switch string(line[:2]) {
		case "@@", "##", "||", "[A":
			// Count ABP specific lines
			abps++
			continue
		default:
			// Not an ABP specific line start
			if "#" == string(line[0]) {
				// Skip a comment line from other file types
				continue
			}
		}

		// Reject lines with invalid characters
		if strings.ContainsAny(line, "/?#@$[]") {
			// Count ABP specific line
			abps++
			continue
		}
	}
	rOK = (0 < abps)

	return
} // isABPfile()

// `isBinary()` checks whether the given file is a known archive format.
//
// Parameters:
//   - `aFile`: The file data to check.
//
// Returns:
//   - `ArchiveFormat`: The archive format if the file is a known archive format,
//     `ArchiveUnknown` otherwise.
//   - `error`: `nil` if the file is a known archive format, the error otherwise.
func isBinary(aFile io.ReadSeeker) (tArchiveFormat, error) {
	if nil == aFile {
		return ArchiveUnknown, ErrInvalidFile
	}
	if _, err := aFile.Seek(0, io.SeekStart); nil != err {
		return ArchiveUnknown, err
	}
	buf := make([]byte, 520) // Enough for all known signatures and TAR check
	n, err := io.ReadFull(aFile, buf)
	if (nil != err) && (err != io.ErrUnexpectedEOF) && (err != io.EOF) {
		return ArchiveUnknown, err
	}
	if 0 == n {
		return ArchiveUnknown, ErrUnknownFileType
	}
	buf = buf[:n]

	// Check known signatures
	for _, sig := range archiveSignatures {
		sigLen := sig.Offset + int64(len(sig.Signature))
		if (int64(len(buf)) >= sigLen) &&
			slices.Equal(buf[sig.Offset:sigLen], sig.Signature) {
			return sig.Format, nil
		}
	}

	// Special case: TAR (ustar) magic at offset 257
	if (265 <= len(buf)) && ("ustar" == string(buf[257:262])) {
		return ArchiveTAR, nil
	}

	return ArchiveUnknown, ErrUnknownFileType
} // isBinary()

// `isHostnamesOnly()` checks whether the given file contains
// only hostname patterns or wildcards.
//
// Parameters:
//   - `aFile`: The file data to check.
//
// Returns:
//   - `rOK`: `true` if the file contains only hostnames or wildcards,
//     `false` otherwise.
func isHostnamesOnly(aFile io.ReadSeeker) (rOK bool) {
	if nil == aFile {
		return
	}
	if _, err := aFile.Seek(0, io.SeekStart); nil != err {
		return
	}

	loops := 0
	scanner := bufio.NewScanner(aFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Ignore empty lines and comment lines
		if (0 == len(line)) ||
			("#" == string(line[0])) ||
			(";" == string(line[0])) {
			continue
		}

		// Allow a single hostname or wildcard per line
		if !isValidHostname(line) && !isValidWildcard(line) {
			return
		}
		loops++
	}
	rOK = (0 < loops)

	return
} // isHostnamesOnly()

// `isHostsFile()` checks whether the given file is a `hosts(5)` file.
//
// Parameters:
//   - `aFile`: The file data to check.
//
// Returns:
//   - `rOK`: `true` if the file is a `hosts(5)` file, `false` otherwise.
func isHostsFile(aFile io.ReadSeeker) (rOK bool) {
	if nil == aFile {
		return
	}
	if _, err := aFile.Seek(0, io.SeekStart); nil != err {
		return
	}

	loops := 0
	scanner := bufio.NewScanner(aFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Ignore empty lines and comment lines
		if (0 == len(line)) ||
			("#" == string(line[0])) ||
			(";" == string(line[0])) {
			continue
		}

		// Split the line into fields
		fields := strings.Fields(line)
		if 2 > len(fields) {
			return
		}

		// Check if the IP is valid
		if nil == net.ParseIP(fields[0]) {
			return
		}

		// Check if the hostname patterns are valid
		for _, pattern := range fields[1:] {
			if !isValidHostname(pattern) && !isValidWildcard(pattern) {
				return
			}
		}

		loops++
	}
	rOK = (0 < loops)

	return
} // isHostsFile()

// `isText()` checks whether the given data is text.
//
// Parameters:
//   - `aFile`: The file data to check.
//
// Returns:
//   - `rOK`: `true` if the data is text, `false` otherwise.
func isText(aFile io.ReadSeeker) (rOK bool) {
	if nil == aFile {
		return
	}
	if _, err := aFile.Seek(0, io.SeekStart); nil != err {
		return
	}

	// Read the first 512 bytes for text detection
	header := make([]byte, 512)
	n, err := aFile.Read(header)
	if (nil != err) && (io.EOF != err) {
		return
	}
	if 0 == n {
		return
	}
	header = header[:n]

	loops := 0
	for _, b := range header {
		if (b < 0x09 || b > 0x0D) && (b < 0x20 || b > 0x7E) {
			return
		}

		loops++
	}
	rOK = (0 < loops)

	return
} // isText()

// `isValidHostname()` checks whether the given pattern is a valid hostname.
//
// Parameters:
//   - `aPattern`: The hostname pattern to check.
//
// Returns:
//   - `bool`: `true` if the pattern is a valid hostname, `false` otherwise.
func isValidHostname(aPattern string) bool {
	if aPattern = strings.TrimSpace(aPattern); (0 == len(aPattern)) || (253 < len(aPattern)) {
		return false
	}

	if 0 < len(toplevelDomains) {
		// Check for valid top-level domain
		tld := filepath.Ext(aPattern)
		if 0 < len(tld) {
			// Remove the leading dot
			tld = tld[1:]
		} else {
			// No top-level domain, use the whole pattern
			tld = aPattern
		}

		if ok := slices.Contains(toplevelDomains, tld); ok {
			return validHostnameRE.MatchString(aPattern)
		}

		return false
	}

	// Fallback: No top-level domain check available
	return validHostnameRE.MatchString(aPattern)
} // isValidHostname()

// `isValidWildcard()` checks whether the given pattern is a valid wildcard.
//
// Parameters:
//   - `aPattern`: The wildcard pattern to check.
//
// Returns:
//   - `bool`: `true` if the pattern is a valid wildcard, `false` otherwise.
func isValidWildcard(aPattern string) (rOK bool) {
	if aPattern = strings.TrimSpace(aPattern); 4 > len(aPattern) {
		return
	}

	// Check for `*.` at the start
	if "*." != aPattern[0:2] {
		return
	}

	// Check for valid hostname after `*.`
	rOK = isValidHostname(aPattern[2:])

	return
} // isValidWildcard()

/* _EoF_ */
