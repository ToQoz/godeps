package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/ToQoz/godeps/godeps"
	"os"
	"path/filepath"
	"strings"
)

func usage() {
	banner := "usage: godeps [flags] [packages(default=.)]\n"

	fmt.Fprintf(os.Stderr, banner)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func main() {
	flag.Usage = usage

	verboseFlag := flag.Bool("v", false, "verbose")

	trackingModeFlag := flag.Int("tracking-mode", 1, `
	0 = track only deps that depth is 1
	1 = stop tracking when godeps reach standard package
	2 = track recursively`)

	flag.Parse()

	godeps.Verbose(*verboseFlag)

	var trackingMode int

	switch *trackingModeFlag {
	case godeps.StopTracingOnReachingTopLevelDeps, godeps.StopTracingOnReachingStandardPackageOrLeaf, godeps.StopTracingOnReachingLeaf:
		trackingMode = *trackingModeFlag
	default:
		trackingMode = godeps.StopTracingOnReachingStandardPackageOrLeaf
	}

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

	pkgs := godeps.Packages(dir)

	if pkgs == nil {
		fmt.Fprintln(os.Stderr, "ERROR: package not found in %s", dir)
		os.Exit(1)
	}

	for _, pkg := range pkgs {
		buf := []byte{}
		buffer := bytes.NewBuffer(buf)

		buffer.WriteString(fmt.Sprintf(`digraph "godeps-of-%v" {`+"\n", pkg.ImportPath))
		buffer.WriteString("    size=13.0;\n")

		for _, dep := range pkg.Deps(nil, trackingMode) {
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
