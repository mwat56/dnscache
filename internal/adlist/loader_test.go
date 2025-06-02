/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package adlist

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_tABPLoader_Load(t *testing.T) {
	loader := &tABPLoader{}
	tmpDir := t.TempDir()
	tests := []struct {
		name     string
		al       *tABPLoader
		filename string
		node     *tNode
		wantErr  bool
	}{
		/* */
		{
			name:     "01 - nil loader",
			al:       nil,
			filename: filepath.Join(tmpDir, "01_abp.txt"),
			node:     newNode(),
			wantErr:  true,
		},
		{
			name:     "02 - empty filename",
			al:       loader,
			filename: "",
			node:     newNode(),
			wantErr:  true,
		},
		{
			name:     "03 - nil node",
			al:       loader,
			filename: filepath.Join(tmpDir, "03_abp.txt"),
			node:     nil,
			wantErr:  true,
		},
		{
			name:     "04 - non existing file",
			al:       loader,
			filename: filepath.Join(tmpDir, "04_abp.txt"),
			node:     newNode(),
			wantErr:  true,
		},
		{
			name: "05 - reader with comments",
			al:   loader,
			filename: func() string {
				fName := filepath.Join(tmpDir, "05_abp.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("# comment\n; comment\n# the next line is no comment\n comment")
				_ = f.Close()
				return fName
			}(),
			node:    newNode(),
			wantErr: true,
		},
		{
			name: "06 - reader with empty lines",
			al:   loader,
			filename: func() string {
				fName := filepath.Join(tmpDir, "06_abp.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("\n\n\n")
				_ = f.Close()
				return fName
			}(),
			node:    newNode(),
			wantErr: true,
		},
		{
			name: "07 - reader with valid data",
			al:   loader,
			filename: func() string {
				fName := filepath.Join(tmpDir, "07_abp.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("||example.com^")
				_ = f.Close()
				return fName
			}(),
			node:    newNode(),
			wantErr: false,
		},
		{
			name: "08 - reader with invalid data",
			al:   loader,
			filename: func() string {
				fName := filepath.Join(tmpDir, "08_abp.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("Hello, World!")
				_ = f.Close()
				return fName
			}(),
			node:    newNode(),
			wantErr: true,
		},
		/* */
		{
			name:     "09 - reader with local ABP file",
			al:       loader,
			filename: "/home/matthias/devel/Go/src/github.com/mwat56/dnscache/internal/adlist/fanboy-annoyance.abp.hosts",
			node:     newNode(),
			wantErr:  false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.al.Load(context.TODO(), tc.filename, tc.node)

			if (err != nil) != tc.wantErr {
				t.Errorf("tABPLoader.Load() error = '%v', wantErr '%v'",
					err, tc.wantErr)
			}
		})
	}
} // Test_tABPLoader_Load()

func Test_tHostsLoader_Load(t *testing.T) {
	loader := &tHostsLoader{}
	tmpDir := t.TempDir()
	tests := []struct {
		name    string
		hl      *tHostsLoader
		file    string
		node    *tNode
		wantErr bool
	}{
		/* */
		{
			name: "01 - nil loader",
			hl:   nil,
			file: func() string {
				fName := filepath.Join(tmpDir, "01_hosts.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("0.0.0.1 tld")
				_ = f.Close()
				return fName
			}(),
			node:    newNode(),
			wantErr: true,
		},
		{
			name:    "02 - empty filename",
			hl:      loader,
			file:    "",
			node:    newNode(),
			wantErr: true,
		},
		{
			name: "03 - nil node",
			hl:   loader,
			file: func() string {
				fName := filepath.Join(tmpDir, "03_hosts.txt")
				f, _ := os.Create(fName)
				_ = f.Close()
				return fName
			}(),
			node:    nil,
			wantErr: true,
		},
		{
			name: "04 - non existing file",
			hl:   loader,
			file: func() string {
				fName := filepath.Join(tmpDir, "04_hosts.txt")
				return fName
			}(),
			node:    newNode(),
			wantErr: true,
		},
		{
			name: "05 - reader with comments",
			hl:   loader,
			file: func() string {
				fName := filepath.Join(tmpDir, "05_hosts.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("# comment\n; comment\n# the next line is no comment\n0.0.0.1 comment")
				_ = f.Close()
				return fName
			}(),
			node:    newNode(),
			wantErr: false,
		},
		{
			name: "06 - reader with empty lines",
			hl:   loader,
			file: func() string {
				fName := filepath.Join(tmpDir, "06_hosts.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("\n\n\n")
				_ = f.Close()
				return fName
			}(),
			node:    newNode(),
			wantErr: false,
		},
		{
			name: "07 - reader with valid data",
			hl:   loader,
			file: func() string {
				fName := filepath.Join(tmpDir, "07_hosts.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("0.0.0.1 tld\n0.0.0.2 domain.tld\n0.0.0.3 host.domain.tld\n\n# The next line is an invalid hots entry\ninvalid\n\n0.0.0.4 *.domain.tld")
				_ = f.Close()
				return fName
			}(),
			node:    newNode(),
			wantErr: false,
		},
		/* */
		{
			name: "08 - file with valid data and multiple IP addresses",
			hl:   loader,
			file: func() string {
				fName := filepath.Join(tmpDir, "08_hosts.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("0.0.0.1 tld\n0.0.0.2 domain.tld\n0.0.0.3 host.domain.tld\n\n# The next line is an invalid hots entry\ninvalid\n\n0.0.0.4 domain.tld\n0.0.0.5 h5.domain.tld\n0.0.0.6 h6.domain.tld\n0.0.0.7 h7.domain.tld\n0.0.0.8 h8.domain.tld\n0.0.0.9 h9.domain.tld\nh10.domain.tld\n")
				_ = f.Close()
				return fName
			}(),
			node:    newNode(),
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.hl.Load(context.TODO(), tc.file, tc.node)
			if (nil != err) != tc.wantErr {
				t.Errorf("tHostsLoader.Load() error = '%v', wantErr '%v'",
					err, tc.wantErr)
			}
		})
	}
} // Test_tHostsLoader_Load()

func Test_tSimpleLoader_Load(t *testing.T) {
	loader := &tSimpleLoader{}
	tmpDir := t.TempDir()
	tests := []struct {
		name    string
		sl      *tSimpleLoader
		file    string
		node    *tNode
		wantErr bool
	}{
		/* */
		{
			name:    "01 - nil loader",
			sl:      nil,
			file:    "tld",
			node:    newNode(),
			wantErr: true,
		},
		{
			name:    "02 - empty filename",
			sl:      loader,
			file:    "",
			node:    newNode(),
			wantErr: true,
		},
		{
			name:    "03 - nil node",
			sl:      loader,
			file:    "tld",
			node:    nil,
			wantErr: true,
		},
		{
			name:    "04 - non existing file",
			sl:      loader,
			file:    "doesnotexist.txt",
			node:    newNode(),
			wantErr: true,
		},
		{
			name: "05 - file with comments",
			sl:   loader,
			file: func() string {
				fName := filepath.Join(tmpDir, "05_hosts.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("# comment\n; comment\n# the next line is no comment\n comment")
				_ = f.Close()
				return fName
			}(),
			node:    newNode(),
			wantErr: false,
		},
		{
			name: "06 - file with empty lines",
			sl:   loader,
			file: func() string {
				fName := filepath.Join(tmpDir, "06_hosts.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("\n\n\n")
				_ = f.Close()
				return fName
			}(),
			node:    newNode(),
			wantErr: false,
		},
		{
			name: "07 - file with valid data",
			sl:   loader,
			file: func() string {
				fName := filepath.Join(tmpDir, "07_hosts.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("tld\ndomain.tld\nhost.domain.tld\ninvalid\n*.domain.tld")
				_ = f.Close()
				return fName
			}(),
			node:    newNode(),
			wantErr: false,
		},
		{
			name: "08 - file with valid data and multiple IP addresses",
			sl:   loader,
			file: func() string {
				fName := filepath.Join(tmpDir, "08_hosts.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("tld\ndomain.tld\nhost.domain.tld\n\n# The next line is an invalid hots entry\ninvalid\n\n domain2.tld\n*.domain2.tld\ndomain.tld2\n*.domain2.tld3\ndomain.tld3\n*.domain.tld3\nhost.domain.tld4\n*.domain.tld4\n*.domain.tld\n")
				_ = f.Close()
				return fName
			}(),
			node:    newNode(),
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.sl.Load(context.TODO(), tc.file, tc.node)
			if (nil != err) != tc.wantErr {
				t.Errorf("tSimpleLoader.Load() error = '%v', wantErr '%v'",
					err, tc.wantErr)
			}
		})
	}
} // Test_tSimpleLoader_Load()

/*
func Test_tHostsSaver_Save(t *testing.T) {
	tests := []struct {
		name     string
		hs       *tHostsSaver
		node     *tNode
		wantText string
		wantErr  bool
	}{
		{
			name:     "01 - nil saver",
			hs:       nil,
			node:     newNode(),
			wantText: "",
			wantErr:  true,
		},
		{
			name:     "02 - nil node",
			hs:       &tHostsSaver{},
			node:     nil,
			wantText: "",
			wantErr:  true,
		},
		{
			name:     "03 - empty node",
			hs:       &tHostsSaver{},
			node:     newNode(),
			wantText: "",
			wantErr:  false,
		},
		{
			name: "04 - node with child",
			hs:   &tHostsSaver{},
			node: func() *tNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain"})
				return n
			}(),
			wantText: "127.0.0.1 domain.tld\n",
			wantErr:  false,
		},
		{
			name: "05 - node with wildcard",
			hs:   &tHostsSaver{},
			node: func() *tNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "*"})
				return n
			}(),
			wantText: "127.0.0.1 *.tld\n",
			wantErr:  false,
		},
		{
			name: "06 - node with children",
			hs:   &tHostsSaver{},
			node: func() *tNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain", "host"})
				return n
			}(),
			wantText: "127.0.0.1 host.domain.tld\n",
			wantErr:  false,
		},
		{
			name: "07 - node with wildcard and children",
			hs:   &tHostsSaver{},
			node: func() *tNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain", "*"})
				n.add(context.TODO(), tPartsList{"tld", "domain", "host"})
				return n
			}(),
			wantText: "127.0.0.1 *.domain.tld\n127.0.0.1 host.domain.tld\n",
			wantErr:  false,
		},
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			aWriter := &bytes.Buffer{}
			err := tc.hs.Save(context.TODO(), aWriter, tc.node)

			if (nil != err) != tc.wantErr {
				t.Errorf("tHostsSaver.Save() error = '%v', wantErr '%v'",
					err, tc.wantErr)
				return
			}
			if gotText := aWriter.String(); gotText != tc.wantText {
				t.Errorf("tHostsSaver.Save() =\n%q\nwant\n%q",
					gotText, tc.wantText)
			}
		})
	}
} // Test_tHostsSaver_Save()
*/

func Test_tSimpleSaver_Save(t *testing.T) {
	tests := []struct {
		name     string
		ss       *tSimpleSaver
		node     *tNode
		wantText string
		wantErr  bool
	}{
		/* */
		{
			name:     "01 - nil saver",
			ss:       nil,
			node:     newNode(),
			wantText: "",
			wantErr:  true,
		},
		{
			name:     "02 - nil node",
			ss:       &tSimpleSaver{},
			node:     nil,
			wantText: "",
			wantErr:  true,
		},
		{
			name:     "03 - empty node",
			ss:       &tSimpleSaver{},
			node:     newNode(),
			wantText: "",
			wantErr:  false,
		},
		{
			name: "04 - node with child",
			ss:   &tSimpleSaver{},
			node: func() *tNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain"})
				return n
			}(),
			wantText: "domain.tld\n",
			wantErr:  false,
		},
		{
			name: "05 - node with wildcard",
			ss:   &tSimpleSaver{},
			node: func() *tNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "*"})
				return n
			}(),
			wantText: "*.tld\n",
			wantErr:  false,
		},
		{
			name: "06 - node with children",
			ss:   &tSimpleSaver{},
			node: func() *tNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain", "host"})
				return n
			}(),
			wantText: "host.domain.tld\n",
			wantErr:  false,
		},
		{
			name: "07 - node with wildcard and children",
			ss:   &tSimpleSaver{},
			node: func() *tNode {
				n := newNode()
				n.add(context.TODO(), tPartsList{"tld", "domain", "*"})
				n.add(context.TODO(), tPartsList{"tld", "domain", "host"})
				return n
			}(),
			wantText: "*.domain.tld\nhost.domain.tld\n",
			wantErr:  false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			aWriter := &bytes.Buffer{}
			err := tc.ss.Save(context.TODO(), aWriter, tc.node)

			if (nil != err) != tc.wantErr {
				t.Errorf("tSimpleSaver.Save() error = '%v', wantErr '%v'",
					err, tc.wantErr)
				return
			}
			if gotText := aWriter.String(); gotText != tc.wantText {
				t.Errorf("tSimpleSaver.Save() =\n%q\nwant\n%q",
					gotText, tc.wantText)
			}

		})
	}
} // Test_tSimpleSaver_Save()

func Test_downloadFile(t *testing.T) {
	tmpDir := t.TempDir()
	tests := []struct {
		name     string
		url      string
		filename string
		wantErr  bool
	}{
		/* */
		{
			name:     "01 - empty url",
			url:      "",
			filename: filepath.Join(tmpDir, "NilUrl.txt"),
			wantErr:  true,
		},
		{
			name:     "02 - nil dir",
			url:      "https://example.com/",
			filename: filepath.Join(tmpDir, "NilDir.txt"),
			wantErr:  false,
		},
		{
			name:     "03 - empty filename",
			url:      "https://example.com/",
			filename: "",
			wantErr:  true,
		},
		{
			name:     "04 - valid url, and filename",
			url:      "https://example.com/",
			filename: filepath.Join(tmpDir, "Valid.txt"),
			wantErr:  false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			name, err := downloadFile(tc.url, tc.filename)

			if (nil != err) != tc.wantErr {
				t.Errorf("DownloadFile() error =\n'%v'\nwantErr '%v'",
					err, tc.wantErr)
			}

			if !tc.wantErr && (name != tc.filename) {
				t.Errorf("DownloadFile() =\n%q\nwant\n%q",
					name, tc.filename)
			}
		})
	}
} // Test_downloadFile()

func Test_isABPfile(t *testing.T) {
	tests := []struct {
		name   string
		file   io.ReadSeeker
		wantOK bool
	}{
		/* */
		{
			name:   "01 - nil file",
			file:   nil,
			wantOK: false,
		},
		{
			name:   "02 - empty file",
			file:   bytes.NewReader([]byte{}),
			wantOK: false,
		},
		{
			name:   "03 - non-empty file",
			file:   bytes.NewReader([]byte("Hello, World!")),
			wantOK: false,
		},
		{
			name:   "04 - ABP file",
			file:   bytes.NewReader([]byte("||example.com^")),
			wantOK: true,
		},
		{
			name:   "05 - newlines file",
			file:   bytes.NewReader([]byte("\n\n\n\n")),
			wantOK: false,
		},
		/* */
		{
			name:   "06 - many comments file",
			file:   bytes.NewReader([]byte("\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n! comment\n")),
			wantOK: true,
		},
		/* */
		{
			name:   "07 - ABP file",
			file:   bytes.NewReader([]byte("[Adblock Plus 1.1]\n! Version: 202505221917\n! Title: EasyList Germany\n! Last modified: 22 May 2025 19:17 UTC\n! Expires: 1 days (update frequency)\n! Homepage: https://easylist.to/\n! Licence: https://easylist.to/pages/licence.html\n!\n! Bitte melde ungeblockte Werbung und fälschlicherweise geblockte Dinge:\n! Forum: https://forums.lanik.us/viewforum.php?f=90\n! E-Mail: easylist.germany@gmail.com\n! GitHub issues: https://github.com/easylist/easylistgermany/issues\n! GitHub pull requests: https://github.com/easylist/easylistgermany/pulls\n!\n! ----------------Allgemeine Regeln zum Blockieren von Werbung-----------------!\n! *** easylistgermany:easylistgermany/easylistgermany_general_block.txt \n**\n-Bannerwerbung-\n-werb_hori.\n-werb_vert.\n-Werbebanner-\n-werbebanner.$domain=~merkur-werbebanner.de\n-Werbebannerr_\n.at/werbung/\n.com/de/ad/\n.com/werbung")),
			wantOK: true,
		},
		/* */
		{
			name:   "08 - hosts file",
			file:   bytes.NewReader([]byte("#=====================================\n# Title: Hosts contributed by Steven Black\n# http://stevenblack.com\n#=====================================\n 0.0.0.0 www.30-day-change.com\n0.0.0.0 mclean.f.360.cn\n0.0.0.0 mvconf.f.360.cn\n0.0.0.0 care.help.360.cn\n0.0.0.0 eul.s.360.cn\n0.0.0.0 g.s.360.cn\n0.0.0.0 p.s.360.cn ")),
			wantOK: false,
		},
		{
			name:   "09 - ;comments file",
			file:   bytes.NewReader([]byte("\n;\n:\n;\n;")),
			wantOK: false,
		},
		{
			name:   "10 - #comments file",
			file:   bytes.NewReader([]byte("\n#\n#\n#\n#")),
			wantOK: false,
		},
		/* */
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := isABPfile(tc.file)
			if gotOK != tc.wantOK {
				t.Errorf("isABPfile() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_isABPfile()

func Test_isBinary(t *testing.T) {
	tests := []struct {
		name    string
		file    io.ReadSeeker
		want    tArchiveFormat
		wantErr bool
	}{
		/* */
		{
			name:    "01 - nil file",
			file:    nil,
			want:    ArchiveUnknown,
			wantErr: true,
		},
		{
			name:    "02 - empty file",
			file:    bytes.NewReader([]byte{}),
			want:    ArchiveUnknown,
			wantErr: true,
		},
		{
			name:    "03 - non-empty file",
			file:    bytes.NewReader([]byte("Hello, World!")),
			want:    ArchiveUnknown,
			wantErr: true,
		},
		{
			name:    "04 - ZIP file",
			file:    bytes.NewReader([]byte{0x50, 0x4B, 0x03, 0x04}),
			want:    ArchiveZIP,
			wantErr: false,
		},
		{
			name:    "05 - RAR file",
			file:    bytes.NewReader([]byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x00}),
			want:    ArchiveRAR,
			wantErr: false,
		},
		{
			name:    "06 - GZIP file",
			file:    bytes.NewReader([]byte{0x1F, 0x8B}),
			want:    ArchiveGZ,
			wantErr: false,
		},
		{
			name:    "07 - BZ2 file",
			file:    bytes.NewReader([]byte{0x42, 0x5A, 0x68}),
			want:    ArchiveBZ2,
			wantErr: false,
		},
		{
			name:    "08 - XZ file",
			file:    bytes.NewReader([]byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}),
			want:    ArchiveXZ,
			wantErr: false,
		},
		{
			name:    "09 - 7z file",
			file:    bytes.NewReader([]byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}),
			want:    Archive7Z,
			wantErr: false,
		},
		{
			name: "10 - TAR file",
			file: func() io.ReadSeeker {
				buf := make([]byte, 512)
				copy(buf[257:262], []byte{0x75, 0x73, 0x74, 0x61, 0x72, 0x00})

				return bytes.NewReader(buf)
			}(),
			want:    ArchiveTAR,
			wantErr: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := isBinary(tc.file)

			if (nil != err) != tc.wantErr {
				t.Errorf("isBinary() error = '%v', wantErr '%v'",
					err, tc.wantErr)
				return
			}
			if got != tc.want {
				t.Errorf("isBinary() = %q, want %q",
					got, tc.want)
			}
		})
	}
} // Test_isBinary()

func Test_isHostnamesOnly(t *testing.T) {
	tests := []struct {
		name   string
		file   io.ReadSeeker
		wantOK bool
	}{
		/* */
		{
			name:   "01 - nil file",
			file:   nil,
			wantOK: false,
		},
		{
			name:   "02 - empty file",
			file:   bytes.NewReader([]byte{}),
			wantOK: false,
		},
		{
			name:   "03 - non-empty file",
			file:   bytes.NewReader([]byte("Hello, World!")),
			wantOK: false,
		},
		{
			name:   "04 - hosts file",
			file:   bytes.NewReader([]byte("127.0.0.1 localhost")),
			wantOK: false,
		},
		{
			name:   "05 - hosts file",
			file:   bytes.NewReader([]byte("127.0.0.1 localhost\n# Comment")),
			wantOK: false,
		},
		{
			name:   "06 - hosts file",
			file:   bytes.NewReader([]byte("127.0.0.1 localhost\n127.0.0.1 localhost")),
			wantOK: false,
		},
		{
			name:   "07 - hosts file",
			file:   bytes.NewReader([]byte("127.0.0.0 localdomain\n127.0.0.1 localhost\n# Comment")),
			wantOK: false,
		},
		{
			name:   "08 - empty lines only",
			file:   bytes.NewReader([]byte("\n\n\n")),
			wantOK: false,
		},
		{
			name:   "09 - comments only",
			file:   bytes.NewReader([]byte("\n# comment1\n# comment2\n; comment3\n")),
			wantOK: false,
		},
		{
			name:   "10 - comments and empty lines",
			file:   bytes.NewReader([]byte("\n# comment1\n\n# comment2\n\n; comment3\n\n\n\n")),
			wantOK: false,
		},
		{
			name:   "11 - pattern file",
			file:   bytes.NewReader([]byte("www.example.com\n# Comment")),
			wantOK: true,
		},
		{
			name:   "12 - email file",
			file:   bytes.NewReader([]byte("test@example.com\n# Comment")),
			wantOK: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := isHostnamesOnly(tc.file)
			if gotOK != tc.wantOK {
				t.Errorf("isHostnamesOnly() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_isHostnamesOnly()

func Test_isHostsFile(t *testing.T) {
	tests := []struct {
		name   string
		file   io.ReadSeeker
		wantOK bool
	}{
		/* */
		{
			name:   "01 - nil file",
			file:   nil,
			wantOK: false,
		},
		{
			name:   "02 - empty file",
			file:   bytes.NewReader([]byte{}),
			wantOK: false,
		},
		{
			name:   "03 - non-empty file",
			file:   bytes.NewReader([]byte("Hello, World!")),
			wantOK: false,
		},
		{
			name:   "04 - hosts file",
			file:   bytes.NewReader([]byte("127.0.0.1 localhost.localdomain")),
			wantOK: true,
		},
		{
			name:   "05 - hosts file",
			file:   bytes.NewReader([]byte("127.0.0.1 localhost.localdomain\n# Comment")),
			wantOK: true,
		},
		{
			name:   "06 - hosts file",
			file:   bytes.NewReader([]byte("127.0.0.1 localhost.localdomain\n127.0.0.2 secondhost.localdomain")),
			wantOK: true,
		},
		/* */
		{
			name:   "07 - hosts file",
			file:   bytes.NewReader([]byte("127.0.0.0 localdomain\n127.0.0.1 localhost\n# Comment")),
			wantOK: true,
		},
		/* */
		{
			name:   "08 - empty lines only",
			file:   bytes.NewReader([]byte("\n\n\n")),
			wantOK: false,
		},
		{
			name:   "09 - comments only",
			file:   bytes.NewReader([]byte("\n# comment1\n# comment2\n; comment3\n")),
			wantOK: false,
		},
		{
			name:   "10 - comments and empty lines",
			file:   bytes.NewReader([]byte("\n# comment1\n\n# comment2\n\n; comment3\n\n\n\n")),
			wantOK: false,
		},
		{
			name:   "11 - pattern file",
			file:   bytes.NewReader([]byte("www.example.com\n# Comment")),
			wantOK: false,
		},
		{
			name:   "12 - email file",
			file:   bytes.NewReader([]byte("0.0.0.0 test@example.com\n# Comment")),
			wantOK: false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := isHostsFile(tc.file)
			if gotOK != tc.wantOK {
				t.Errorf("isHostsFile() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_isHostsFile()

func Test_isText(t *testing.T) {
	tests := []struct {
		name   string
		file   io.ReadSeeker
		wantOK bool
	}{
		/* */
		{
			name:   "01 - nil file",
			file:   nil,
			wantOK: false,
		},
		{
			name:   "02 - empty file",
			file:   bytes.NewReader([]byte{}),
			wantOK: false,
		},
		{
			name:   "03 - non-empty file",
			file:   bytes.NewReader([]byte("Hello, World!")),
			wantOK: true,
		},
		{
			name:   "04 - binary file",
			file:   bytes.NewReader([]byte{0x00, 0x01, 0x02, 0x03}),
			wantOK: false,
		},
		{
			name:   "05 - text file",
			file:   bytes.NewReader([]byte("Hello, World!\n# Comment")),
			wantOK: true,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := isText(tc.file)
			if gotOK != tc.wantOK {
				t.Errorf("isText() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_isText()

func Test_isValidHostname(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    bool
	}{
		/* */
		{
			name:    "01 - empty pattern",
			pattern: "",
			want:    false,
		},
		{
			name:    "02 - too long pattern",
			pattern: "a" + strings.Repeat("a", 253),
			want:    false,
		},
		{
			name:    "03 - valid hostname",
			pattern: "example.com",
			want:    true,
		},
		{
			name:    "04 - valid hostname with subdomain",
			pattern: "sub.example.com",
			want:    true,
		},
		{
			name:    "05 - valid hostname with multiple subdomains",
			pattern: "host.sub.example.com",
			want:    true,
		},
		{
			name:    "06 - invalid hostname with wildcard",
			pattern: "\n# not allowed here:\n*.sub.example.com",
			want:    false,
		},
		{
			name:    "07 - invalid hostname with invalid character",
			pattern: "example.com!uucp",
			want:    false,
		},
		{
			name:    "08 - invalid hostname with too long label",
			pattern: "a" + strings.Repeat("a", 64) + ".example.com",
			want:    false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isValidHostname(tc.pattern)
			if got != tc.want {
				t.Errorf("isValidHostname() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_isValidHostname()

func Test_isValidWildcard(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantOK  bool
	}{
		/* */
		{
			name:    "01 - empty pattern",
			pattern: "",
			wantOK:  false,
		},
		{
			name:    "02 - too short pattern",
			pattern: "*.",
			wantOK:  false,
		},
		{
			name:    "03 - valid wildcard",
			pattern: "*.example.com",
			wantOK:  true,
		},
		{
			name:    "04 - invalid wildcard with too long label",
			pattern: "*." + strings.Repeat("a", 64) + ".example.com",
			wantOK:  false,
		},
		{
			name:    "05 - invalid wildcard with invalid character",
			pattern: "*.example.com!uucp",
			wantOK:  false,
		},
		{
			name:    "06 - invalid wildcard",
			pattern: "host.*.domain.tld",
			wantOK:  false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotOK := isValidWildcard(tc.pattern)
			if gotOK != tc.wantOK {
				t.Errorf("isValidWildcard() = '%v', want '%v'",
					gotOK, tc.wantOK)
			}
		})
	}
} // Test_isValidWildcard()

func Test_detectFileType(t *testing.T) {
	tmpDir := t.TempDir()
	tests := []struct {
		name     string
		filename string
		wantMime string
		wantErr  bool
	}{
		/* */
		{
			name:     "01 - nil filename",
			filename: "",
			wantMime: "",
			wantErr:  true,
		},
		{
			name:     "02 - non-existent file",
			filename: "doesnotexist.txt",
			wantMime: "",
			wantErr:  true,
		},
		{
			name: "03 - empty file",
			filename: func() string {
				fName := filepath.Join(tmpDir, "empty.txt")
				f, _ := os.Create(fName)
				_ = f.Close()
				return fName
			}(),
			wantMime: "",
			wantErr:  true,
		},
		{
			name: "04 - hosts file",
			filename: func() string {
				fName := filepath.Join(tmpDir, "hosts.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("127.0.0.1 localhost.localdomain\n")
				_ = f.Close()
				return fName
			}(),
			wantMime: "text/x-hosts",
			wantErr:  false,
		},
		{
			name: "05 - hostnames file",
			filename: func() string {
				fName := filepath.Join(tmpDir, "hostnames.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("localhost.localdomain\n")
				_ = f.Close()
				return fName
			}(),
			wantMime: "text/x-hostnames",
			wantErr:  false,
		},
		/* */
		{
			name: "06 - plain text file",
			filename: func() string {
				fName := filepath.Join(tmpDir, "plain.txt")
				f, _ := os.Create(fName)
				_, _ = f.WriteString("Oh dear, this is just\na plain text\n\n# and a comment\n")
				_ = f.Close()
				return fName
			}(),
			wantMime: "text/plain",
			wantErr:  false,
		},
		{
			name: "07 - ZIP file",
			filename: func() string {
				fName := filepath.Join(tmpDir, "test.zip")
				_ = os.WriteFile(fName, []byte{0x50, 0x4B, 0x03, 0x04}, 0644)
				return fName
			}(),
			wantMime: "application/x-zip",
			wantErr:  false,
		},
		/* */
		// TODO: Add test cases.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotMime, err := detectFileType(tc.filename)
			if (nil != err) != tc.wantErr {
				t.Errorf("detectFileType() error = %q, wantErr '%v'",
					err, tc.wantErr)
				return
			}
			if gotMime != tc.wantMime {
				t.Errorf("detectFileType() = %q, want %q",
					gotMime, tc.wantMime)
			}
		})
	}
} // Test_detectFileType()

func Test_ProcessABPLine(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		valid    bool
	}{
		{"||example.com^", "*.example.com", true},
		{"|http://example.com/|", "example.com", true},
		{"http://example.com:8080", "example.com", true},
		{"|http://example.com/index.html|", "", false},
		{"||ads.example.com/banner.gif^", "", false},
		{"/ad/*", "", false},
		{"*analytics*.js", "", false},
		{"malformed/domain", "", false},
		{"||host.*example.com^", "*.host.*example.com", true},
		{"host.domain.tld", "host.domain.tld", true},
		{"host.*.domain.tld", "", false},
		{"test@domain.tld", "", false},
		{"|https://host.domain.tld?redirect=https://example.com", "", false},
		{"|https:// ", "", false},
	}

	for _, tc := range tests {
		pattern, ok := processABPLine(tc.input)
		if ok != tc.valid || pattern != tc.expected {
			t.Errorf("Input %q: Expected (%q, %t), Got (%q, %t)",
				tc.input, tc.expected, tc.valid, pattern, ok)
		}
	}
} // Test_processABPLine()

/* _EoF_ */
