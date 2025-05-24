// textar converts to and extracts from textar files.
// Use -help to see the usage.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ypsu/textar"
)

var (
	flagC = flag.String("c", "", "Create a textar file in this file. Use - to write to stdout.")
	flagU = flag.String("u", "", "Update or extend the specified textar file with the files specified as arguments.")
	flagX = flag.String("x", "", "Extract all or the specified files to the current directory. Use - to decompress stdin.")
)

func usage() {
	fmt.Fprintln(flag.CommandLine.Output(), "Manipulate .textar files.")
	fmt.Fprintln(flag.CommandLine.Output(), "")
	fmt.Fprintln(flag.CommandLine.Output(), "Create a textar file:  textar -c=archive.textar file1 file2 file3")
	fmt.Fprintln(flag.CommandLine.Output(), "Extract a textar file: textar -x=archive.textar")
	fmt.Fprintln(flag.CommandLine.Output(), "Update a single filee: textar -u=archive.textar file2")
	fmt.Fprintln(flag.CommandLine.Output(), "")
	fmt.Fprintln(flag.CommandLine.Output(), "If an argument refers to a directory then textar selects all files from under it.")
	fmt.Fprintln(flag.CommandLine.Output(), "")
	fmt.Fprintln(flag.CommandLine.Output(), "Flags:")
	flag.PrintDefaults()
}

func create(fn string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("main.EmptyArgs (-o needs some filenames to put into the textar file)")
	}

	var ar textar.Archive
	for _, arg := range args {
		err := filepath.Walk(arg, func(filename string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			data, err := os.ReadFile(filename)
			if err != nil {
				return err // io error contains both the action and error
			}
			ar.Files = append(ar.Files, textar.File{filename, data})
			return nil
		})
		if err != nil {
			return err // walk returns an io error which contains both the action and error
		}
	}

	data := ar.Format()
	if fn == "-" {
		fn = "/dev/stdout"
	}
	return os.WriteFile(fn, data, 0644) // io error contains both the action and error
}

func update(fn string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("main.EmptyArgs (-u needs some filenames to put into the textar file)")
	}
	data, err := os.ReadFile(fn)
	if err != nil {
		return err // io error contains both the action and error
	}

	ar := textar.Parse(data)
	index := map[string]int{}
	for i, f := range ar.Files {
		index[f.Name] = i
	}

	for _, arg := range args {
		err := filepath.Walk(arg, func(filename string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			data, err := os.ReadFile(filename)
			if err != nil {
				return err // io error contains both the action and error
			}
			if i, exists := index[filename]; exists {
				ar.Files[i].Data = data
			} else {
				index[arg], ar.Files = len(arg), append(ar.Files, textar.File{filename, data})
			}
			return nil
		})
		if err != nil {
			return err // walk returns an io error which contains both the action and error
		}
	}

	return os.WriteFile(fn, ar.Format(), 0644) // io error contains both the action and error
}

func subdir(dir, file string) bool {
	if strings.HasSuffix(dir, "/") {
		return strings.HasPrefix(file, dir)
	}
	r, found := strings.CutPrefix(file, dir)
	return found && strings.HasPrefix(r, "/")
}

func extract(fn string, args []string) error {
	if fn == "-" {
		fn = "/dev/stdin"
	}
	data, err := os.ReadFile(fn)
	if err != nil {
		return err // io error contains both the action and error
	}

	for name, data := range textar.Parse(data).Range() {
		ok := true
		if len(args) > 0 {
			ok = false
			for _, a := range args {
				if name == a || subdir(a, name) {
					ok = true
					break
				}
			}
		}
		if !ok {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			return err // io error contains both the action and error
		}
		if err := os.WriteFile(name, data, 0644); err != nil {
			return err // io error contains both the action and error
		}
	}

	return nil
}

func run() error {
	if len(os.Args) <= 1 {
		usage()
	}
	flag.Usage = usage
	flag.Parse()

	flagcnt := 0
	for _, v := range []string{*flagC, *flagU, *flagX} {
		if v != "" {
			flagcnt++
		}
	}
	if flagcnt != 1 {
		return fmt.Errorf("main.NotOneAction count=%d (must have exactly one of -o, -u, or -x flags set)", flagcnt)
	}

	switch {
	case *flagC != "":
		return create(*flagC, flag.Args())
	case *flagU != "":
		return update(*flagU, flag.Args())
	case *flagX != "":
		return extract(*flagX, flag.Args())
	}

	return fmt.Errorf("main.CannotHappen (because the switch above handled all cases)")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
