// The seq command demonstrates the usage of the textar library, see seq.textar for the details.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	_ "embed"

	"github.com/ypsu/textar"
)

func run() error {
	archive, err := textar.ParseFile("seq.textar")
	if err != nil {
		return fmt.Errorf("seq.ParseFile: %v", err)
	}
	for i, f := range archive.Files {
		if strings.HasPrefix(f.Name, "#") {
			continue
		}
		args := strings.Split(f.Name, " ")
		output, err := exec.Command(args[0], args[1:]...).Output()
		if err != nil {
			return err
		}
		archive.Files[i].Data = output
	}
	return os.WriteFile("seq.textar", archive.Format(), 0644)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Command failed: %v.\n", err)
		os.Exit(1)
	}
}
