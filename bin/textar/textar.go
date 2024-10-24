// textar converts to and extracts from textar files.
// Use -help to see the usage.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ypsu/textar"
)

var (
	flagC = flag.String("c", "", "Create a textar file in this file. Use - to write to stdout.")
	flagU = flag.String("u", "", "Update or extend the specified textar file with the files specified as arguments. Preserves comments.")
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

	var ar []textar.File
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
			ar = append(ar, textar.File{filename, data})
			return nil
		})
		if err != nil {
			return err // walk returns an io error which contains both the action and error
		}
	}

	data := textar.Format(ar)
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

	ar := textar.ParseOptions{ParseComments: true}.Parse(data)
	index := map[string]int{}
	for i, f := range ar {
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
				ar[i].Data = data
			} else {
				index[arg], ar = len(arg), append(ar, textar.File{filename, data})
			}
			return nil
		})
		if err != nil {
			return err // walk returns an io error which contains both the action and error
		}
	}

	fo := textar.FormatOptions{}
	if len(data) > 0 {
		fo.Separator = data[0]
	}
	return os.WriteFile(fn, fo.Format(ar), 0644) // io error contains both the action and error
}

func extract(fn string, args []string) error {
	data, err := os.ReadFile(fn)
	if err != nil {
		return err // io error contains both the action and error
	}

	selection := map[string]bool{}
	for _, arg := range args {
		selection[arg] = true
	}

	for _, f := range textar.Parse(data) {
		if len(selection) > 0 && !selection[f.Name] {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(f.Name), 0755); err != nil {
			return err // io error contains both the action and error
		}
		if err := os.WriteFile(f.Name, f.Data, 0644); err != nil {
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
