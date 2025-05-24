// Package textar encodes a file list (key-value slice) into a human editable text file and vice versa.
// This is inspired by https://pkg.go.dev/golang.org/x/tools/txtar but this format can encode any content perfectly without issues.
// Go's txtar doesn't handle newlines and content containing txtar markers well.
// In textar each file in a textar is encoded via "[SEP] [NAME]\n[CONTENT]\n".
// SEP is two or more = signs.
// The first SEP can be arbitrary length, the rest must be the same length.
// The first line beginning with == determines the separator length.
// The dynamic SEP-lengtha makes it possible to encode and decode anything perfectly, this is the main advantage over Go's txtar.
// Furthermore anything before the first SEP is free form comment.
// Example:
//
//	Some comments here.
//
//	=== file1
//	file1 content.
//
//	=== file2
//	file2 content.
//	== file3
//	this is a textar within textar.
//
// The separator here is === so this textar contains file1 and file2.
// file3 is not part of the main textar, it just shows that file2 could be a textar itself.
//
// The [Archive.Format] function automatically picks a separator length that is unique and won't conflict with existing file values.
// Use [Parse] to parse it back.
//
// See https://github.com/ypsu/textar/blob/main/example/seq.textar for a longer example.
// See the testdata directory of https://github.com/ypsu/pkgtrim for a more realistic example.
package textar

import (
	"bytes"
	"fmt"
	"iter"
	"math"
	"os"
	"strings"
	"testing/fstest"
)

// A File is a single file in an archive.
type File struct {
	Name string
	Data []byte
}

// An Archive is a collection of files.
type Archive struct {
	Comment []byte
	Files   []File
}

// Parse parses the serialized form of an Archive. The returned Archive holds slices of data.
func Parse(data []byte) *Archive {
	a, p := &Archive{}, data
	if len(data) <= 2 {
		a.Comment = p
		return a
	}

	// Find the separator string.
	sep := make([]byte, 0, 5)
	sep = append(sep, '\n', '=', '=')
	if p[0] == '=' && p[1] == '=' {
		p = p[2:]
	} else {
		var ok bool
		a.Comment, p, ok = bytes.Cut(p, []byte("\n=="))
		if !ok {
			// Empty textar, treat the whole file as a big comment.
			a.Comment = data
			return a
		}
	}
	for len(p) > 0 && p[0] == '=' {
		p, sep = p[1:], append(sep, '=')
	}
	if len(p) == 0 || p[0] != ' ' {
		// Invalid textar, treat the whole file as a big comment.
		a.Comment = data
		return a
	}
	sep, p = append(sep, ' '), p[1:]

	// Populate the Files field.
	for {
		var name, data []byte
		var ok bool
		name, p, ok = bytes.Cut(p, []byte("\n"))
		if !ok {
			break
		}
		data, p, _ = bytes.Cut(p, sep)
		a.Files = append(a.Files, File{string(name), data})
	}
	return a
}

// ParseFile parses the named file as an archive.
func ParseFile(file string) (*Archive, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("textar.ReadFile: %v", err)
	}
	return Parse(data), nil
}

// Format an archive into a byte stream with custom settings.
func (a *Archive) Format() []byte {
	if a == nil {
		return nil
	}

	// Compute the separator.
	separator := []byte{'\n'} // the full separator: newline, equal signs, space
	sepcnt := 2
	for _, f := range a.Files {
		run := math.MinInt
		for _, ch := range f.Data {
			switch ch {
			case '\n':
				run = 1
			case '=':
				run++
			case ' ':
				sepcnt = max(sepcnt, run)
			default:
				run = math.MinInt
			}
		}
	}
	separator = append(separator, bytes.Repeat([]byte{'='}, sepcnt)...)
	separator = append(separator, ' ')

	// Generate the archive.
	p := &bytes.Buffer{}
	p.Write(a.Comment)
	for i, f := range a.Files {
		if i == 0 && len(a.Comment) == 0 {
			p.Write(separator[1:])
		} else {
			p.Write(separator)
		}
		p.WriteString(strings.ReplaceAll(f.Name, "\n", `\n`))
		p.WriteByte('\n')
		p.Write(f.Data)
	}
	return p.Bytes()
}

// Range iterates over the Files.
func (a *Archive) Range() iter.Seq2[string, []byte] {
	return func(yield func(name string, data []byte) bool) {
		for _, file := range a.Files {
			if !yield(file.Name, file.Data) {
				return
			}
		}
	}
}

// FS returns an object implementing [io/fs.FS] built from the contents of an archive.
// This is a helper function for tests.
func (a *Archive) FS() fstest.MapFS {
	fs := fstest.MapFS{}
	for name, data := range a.Range() {
		fs[strings.TrimPrefix(name, "/")] = &fstest.MapFile{Data: data, Mode: 0644}
	}
	return fs
}
