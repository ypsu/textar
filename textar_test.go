package textar_test

import (
	"bytes"
	"testing"

	"github.com/ypsu/textar"
)

func TestTextar(t *testing.T) {
	srcArchive := []textar.File{
		{"file1", []byte("content 1")},
		{"file2", []byte("content 2\n")},
		{"file3", []byte("content 3\n== with separator\n")},
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
}

func TestIndent(t *testing.T) {
	src, indent := []byte("some\nXXXindented\nstring"), "XXX"
	dst := textar.Unindent(textar.Indent(src, indent), indent)
	t.Logf("Indented string: %q.", textar.Indent(src, indent))
	if !bytes.Equal(dst, src) {
		t.Errorf("Unindent() = %q, want %q.", dst, src)
	}
}
