// Package textar encodes a file list (key-value slice) into a human editable text file and vice versa.
// This is inspired by https://pkg.go.dev/golang.org/x/tools/txtar but this format can encode any content without issues.
// Each file in a textar is encoded via "[SEP] [NAME]\n[CONTENT]\n".
// The first SEP in the file determines the SEP for the whole file.
// SEP cannot contain space or newline.
// Example:
//
//	== file1
//	file1 content
//	== file2
//	file2 content
//	== # some comment title
//	some comment body.
//	== file3
//	file3 content.
//
// The separator here is == because that's how the archive starts.
// The [Format] function automatically picks a separator that is unique and won't conflict with existing file values.
// Use [Parse] to parse it back into a slice of [File] entries.
//
// By default textar skips parsing entries starting with #.
// They can be used as comments.
// This behavior can be altered with [ParseOptions].
//
// See https://github.com/ypsu/textar/blob/main/example/seq.textar for a longer example.
package textar

import (
	"bytes"
	"math"
	"strings"
)

// File represents a single entry in the text archive.
type File struct {
	Name string
	Data []byte
}

// ParseOptions allows customizing the parsing.
type ParseOptions struct {
	// If true, textar won't skip entries that have a name starting with # or have an empty name.
	ParseComments bool
	// Parse appends the resulting Files to this buffer.
	Buffer []File
}

// FormatOptions allows customizing the formatting.
type FormatOptions struct {
	// The byte which gets repeated and then used to separate the Files.
	// Defaults to '=' if unspecified or set to an invalid value.
	Separator byte

	// Format appends the resulting data to this buffer.
	// Use this to reduce memory allocations.
	// Don't use it for concatenating textars, it won't work.
	Buffer []byte
}

// Parse data with the default settings.
// Note by default textar skips entries that start with the # comment marker.
// Use [ParseOptions] to alter this.
func Parse(data []byte) []File {
	return ParseOptions{}.Parse(data)
}

// Parse data with custom settings.
func (po ParseOptions) Parse(data []byte) []File {
	var sep, name, filedata []byte
	archive := po.Buffer

	sep, data, _ = bytes.Cut(data, []byte(" "))
	sep = append(append([]byte("\n"), sep...), ' ')
	for len(data) > 0 {
		var ok bool
		name, data, ok = bytes.Cut(data, []byte("\n"))
		if !ok {
			return archive
		}
		filedata, data, _ = bytes.Cut(data, sep)
		if !po.ParseComments && (len(name) == 0 || name[0] == '#') {
			continue
		}
		archive = append(archive, File{string(name), filedata})
	}
	return archive
}

// Format an archive into a byte stream with the default settings.
func Format(archive []File) []byte {
	return FormatOptions{}.Format(archive)
}

// Format an archive into a byte stream with custom settings.
func (fo FormatOptions) Format(archive []File) []byte {
	if len(archive) == 0 {
		return fo.Buffer
	}

	var (
		separator []byte                       // the full separator starting with a newline and a run of the separator byte
		buffer    = bytes.NewBuffer(fo.Buffer) // the result
	)

	// Compute the separator.
	sepcnt, sepchar := 2, fo.Separator
	if sepchar == 0 || sepchar == '\n' {
		sepchar = '='
	}
	for _, f := range archive {
		run := math.MinInt
		for _, ch := range f.Data {
			switch ch {
			case '\n':
				run = 0
			case sepchar:
				run, sepcnt = run+1, max(sepcnt, run+2)
			default:
				run = math.MinInt
			}
		}
	}
	separator = append(bytes.Repeat([]byte{sepchar}, sepcnt), ' ')

	// Generate the archive.
	for i, f := range archive {
		if i != 0 {
			buffer.WriteByte('\n')
		}
		buffer.Write(separator)
		buffer.WriteString(strings.ReplaceAll(f.Name, "\n", `\n`))
		buffer.WriteByte('\n')
		buffer.Write(f.Data)
	}
	return buffer.Bytes()
}

// Indent is a convenience function to indent data.
// Note that textar doesn't need this but indenting makes a textar easier to read if it contains embedded textars.
func Indent(data []byte, indent string) []byte {
	if indent == "" || len(data) == 0 {
		return data
	}
	return append([]byte(indent), bytes.ReplaceAll(data, []byte("\n"), []byte("\n"+indent))...)
}

// Unindent is a convenience function to undo Indent.
func Unindent(data []byte, indent string) []byte {
	if indent == "" {
		return data
	}
	return bytes.TrimPrefix(bytes.ReplaceAll(data, []byte("\n"+indent), []byte("\n")), []byte(indent))
}
