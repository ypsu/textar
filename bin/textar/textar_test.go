package main

import (
	"os"
	"testing"
)

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestCreate(t *testing.T) {
	os.Chdir(t.TempDir())

	check(t, os.WriteFile("testfile1", []byte("testcontent1"), 0644))
	check(t, os.WriteFile("testfile2", []byte("testcontent2"), 0644))
	check(t, os.MkdirAll("testdir1/testdir2", 0755))
	check(t, os.WriteFile("testdir1/testdir2/testfile3", []byte("testcontent3"), 0644))
	check(t, create("ar.textar", []string{"testfile1", "testdir1"}))

	data, err := os.ReadFile("ar.textar")
	check(t, err)
	if got, want := string(data), "== testfile1\ntestcontent1\n== testdir1/testdir2/testfile3\ntestcontent3"; got != want {
		t.Fatalf("Invalid create() result:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestUpdate(t *testing.T) {
	os.Chdir(t.TempDir())

	check(t, os.WriteFile("ar.textar", []byte("== testfile1\ntestcontent1\n== testdir1/testdir2/testfile3\ntestcontent3"), 0644))
	check(t, os.MkdirAll("testdir1/testdir2", 0755))
	check(t, os.WriteFile("testdir1/testdir2/testfile3", []byte("newcontent3"), 0644))
	check(t, os.WriteFile("testfile2", []byte("newcontent2"), 0644))
	check(t, update("ar.textar", []string{"testdir1", "testfile2"}))

	data, err := os.ReadFile("ar.textar")
	check(t, err)
	if got, want := string(data), "== testfile1\ntestcontent1\n== testdir1/testdir2/testfile3\nnewcontent3\n== testfile2\nnewcontent2"; got != want {
		t.Fatalf("Invalid update() result:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestExtract(t *testing.T) {
	os.Chdir(t.TempDir())
	check(t, os.WriteFile("ar.textar", []byte("== testdir/testfile1\ntestcontent1\n== testdir/testdir2/testfile3\ntestcontent3"), 0644))

	check(t, extract("ar.textar", []string{"testdir/testdir2/testfile3"}))
	check(t, create("test.textar", []string{"testdir"}))
	data, err := os.ReadFile("test.textar")
	check(t, err)
	if got, want := string(data), "== testdir/testdir2/testfile3\ntestcontent3"; got != want {
		t.Fatalf("Invalid extract() result:\ngot:  %q\nwant: %q", got, want)
	}

	check(t, extract("ar.textar", nil))
	check(t, create("test.textar", []string{"testdir"}))
	data, err = os.ReadFile("test.textar")
	check(t, err)
	if got, want := string(data), "== testdir/testdir2/testfile3\ntestcontent3\n== testdir/testfile1\ntestcontent1"; got != want {
		t.Fatalf("Invalid extract() result:\ngot:  %q\nwant: %q", got, want)
	}
}
