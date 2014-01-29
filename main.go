package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/ToQoz/godeps/deps"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	verboseFlag := flag.Bool("v", false, "verbose flag.")
	deepFlag := flag.Bool("deep", false, "deep flag. If this is set, godeps don't stop tracing dependency even if it reach to standard package.")

	flag.Parse()

	deps.Verbose(*verboseFlag)
	deps.Deep(*deepFlag)

	dir := flag.Arg(0)

	if dir == "" {
		dir = "."
	}

	if !strings.HasPrefix(dir, "/") {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		dir = filepath.Join(wd, dir)
	}

	pkgs := deps.Packages(dir)

	if pkgs == nil {
		fmt.Fprintln(os.Stderr, "ERROR: package not found")
		os.Exit(1)
	}

	for _, pkg := range pkgs {
		buf := []byte{}
		buffer := bytes.NewBuffer(buf)

		buffer.WriteString(fmt.Sprintf(`digraph "godeps-of-%v" {`+"\n", pkg.ImportPath))
		buffer.WriteString("    size=13.0;\n")

		for _, dep := range pkg.Deps(nil) {
			attrList := ""

			if !dep.To.StandardPkg() {
				attrList = "[color=red]"
			}

			buffer.WriteString(fmt.Sprintf(`    "%s" -> "%s"%s;`+"\n", dep.From.ImportPath, dep.To.ImportPath, attrList))
		}

		buffer.WriteString("}\n")

		// truncating it if it already exists
		f, err := os.Create(filepath.Join(pkg.FilePath, "godeps.dot"))

		if err != nil {
			panic(err)
		}

		f.Write(buffer.Bytes())
	}

	os.Exit(0)
}
