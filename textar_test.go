package textar_test

import (
	"bytes"
	"io/fs"
	"testing"

	"github.com/ypsu/textar"
)

func TestTextar(t *testing.T) {
	srcArchive := []textar.File{
		{"file1", []byte("content 1")},
		{"file2", []byte("content 2\n")},
		{"somedir/file3", []byte("content 3\n== with separator\n")},
		{"/file4", nil},
	}
	data := textar.Format(srcArchive)
	t.Logf("Encoded textar:\n%s", data)

	dstArchive := textar.Parse(data)
	if len(dstArchive) != len(srcArchive) {
		t.Fatalf("Archive size = %d, want %d.", len(dstArchive), len(srcArchive))
	}
	for i := range dstArchive {
		if dstArchive[i].Name != srcArchive[i].Name || !bytes.Equal(dstArchive[i].Data, srcArchive[i].Data) {
			t.Errorf("dst[i] = {%q, %q}, want {%q, %q}.", dstArchive[i].Name, dstArchive[i].Data, srcArchive[i].Name, srcArchive[i].Data)
		}
	}

	var dir fs.FS
	dir = textar.FS(srcArchive)
	_, err := fs.ReadFile(dir, "somedir/nonexistent")
	if err == nil {
		t.Errorf("Reading somedir/nonexistent didn't return an error.")
	}
	contents, err := fs.ReadFile(dir, "somedir/file3")
	if err != nil || bytes.Compare(contents, srcArchive[2].Data) != 0 {
		t.Errorf("Inconsistent content in somedir/file3.")
	}
	if matches, _ := fs.Glob(dir, "*"); len(matches) != 4 {
		t.Errorf("Glob() = %d, want %d.\n", len(matches), 4)
	}
}

func TestIndent(t *testing.T) {
	src, indent := []byte("some\nXXindented\nstring"), "XX"
	dst := textar.Unindent(textar.Indent(src, indent), indent)
	t.Logf("Indented string: %q.", textar.Indent(src, indent))
	if !bytes.Equal(dst, src) {
		t.Errorf("Unindent() = %q, want %q.", dst, src)
	}
}
