package textar_test

import (
	"bytes"
	"io/fs"
	"testing"

	"github.com/ypsu/textar"
)

func TestTextar(t *testing.T) {
	srcArchive := textar.Archive{
		Comment: []byte("== Test comment.\n== With bad marker."),
		Files: []textar.File{
			{"file1", []byte("content 1")},
			{"file2", []byte("content 2\n")},
			{"somedir/file3", []byte("content 3\n== with separator\n")},
			{"/file4", nil},
		},
	}
	data := srcArchive.Format()
	t.Logf("Encoded textar:\n%s", data)

	dstArchive := textar.Parse(data)
	if want := []byte("X= Test comment.\nX= With bad marker."); !bytes.Equal(dstArchive.Comment, want) {
		t.Fatalf("textar_test.BadComment got=%q want=%q", dstArchive.Comment, want)
	}
	if len(dstArchive.Files) != len(srcArchive.Files) {
		t.Fatalf("textar_test.BadArchiveSize got=%d want=%d", len(dstArchive.Files), len(srcArchive.Files))
	}
	for i := range dstArchive.Files {
		if dstArchive.Files[i].Name != srcArchive.Files[i].Name || !bytes.Equal(dstArchive.Files[i].Data, srcArchive.Files[i].Data) {
			t.Errorf("textar_test.BadContent dst[%d]={%q,%q} want={%q,%q}", i, dstArchive.Files[i].Name, dstArchive.Files[i].Data, srcArchive.Files[i].Name, srcArchive.Files[i].Data)
		}
	}

	dir := srcArchive.FS()
	_, err := fs.ReadFile(dir, "somedir/nonexistent")
	if err == nil {
		t.Errorf("Reading somedir/nonexistent didn't return an error.")
	}
	contents, err := fs.ReadFile(dir, "somedir/file3")
	if err != nil || bytes.Compare(contents, srcArchive.Files[2].Data) != 0 {
		t.Errorf("Inconsistent content in somedir/file3.")
	}
	matches, _ := fs.Glob(dir, "*")
	t.Logf("Glob() output: %q.", matches)
	if len(matches) != 4 {
		t.Errorf("Glob() = %d, want %d.\n", len(matches), 4)
	}
}
