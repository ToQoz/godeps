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

	trackingModeFlag := flag.Int("tracking-mode", 1, `0 = track only deps that depth is 1
1 = stop tracing when godeps reach standard package
2 = track recursively`)

	flag.Parse()

	deps.Verbose(*verboseFlag)

	var trackingMode int

	switch *trackingModeFlag {
	case deps.StopTracingOnReachingTopLevelDeps, deps.StopTracingOnReachingStandardPackageOrLeaf, deps.StopTracingOnReachingLeaf:
		trackingMode = *trackingModeFlag
	default:
		trackingMode = deps.StopTracingOnReachingStandardPackageOrLeaf
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

	pkgs := deps.Packages(dir)

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
