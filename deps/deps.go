package deps

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	verbose = false
)

const (
	StopTracingOnReachingTopLevelDeps = iota
	StopTracingOnReachingStandardPackageOrLeaf
	StopTracingOnReachingLeaf
)

func Verbose(v bool) {
	verbose = v
}

func Packages(_pattern string) []*Package {
	var pkgs []*Package
	var pattern *regexp.Regexp

	var dir string

	if strings.Index(_pattern, "...") != -1 {
		i := strings.Index(_pattern, "...")
		dir, _ = filepath.Split(_pattern[:i])
	} else {
		dir = _pattern
	}

	_pattern = regexp.QuoteMeta(_pattern)
	_pattern = strings.Replace(_pattern, `\.\.\.`, `.*`, -1)
	_pattern = strings.Replace(_pattern, `/.*`, `(/.*)?`, -1)
	pattern = regexp.MustCompile(`^` + _pattern + `$`)

	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}

		// Skip file.
		if !f.IsDir() {
			if verbose {
				fmt.Fprintln(os.Stderr, "SKIP "+path+". It is a file.")
			}
			return nil
		}

		// Ignore .<DIR>, or _<DIR>, or testdata
		_, segment := filepath.Split(path)

		hiddenDir := strings.HasPrefix(segment, ".") && segment != "." && segment != ".."
		testdataDir := segment == "testdata"

		if hiddenDir || testdataDir {
			if verbose {
				fmt.Fprintln(os.Stderr, "SKIP "+path+". It match to ignore rule.")
			}
			return filepath.SkipDir
		}

		// Match
		if pattern != nil && !pattern.MatchString(path) {
			if verbose {
				fmt.Fprintln(os.Stderr, "SKIP "+path+". It don't match to %v\n", pattern)
			}
			return filepath.SkipDir
		}

		// Check dir contains go file
		files, err := ioutil.ReadDir(path)

		if err != nil {
			panic(err)
		}

		goFound := false

		for _, f := range files {
			if f.IsDir() {
				continue
			}

			if filepath.Ext(f.Name()) == ".go" {
				goFound = true
			}
		}

		if goFound {
			pkg := &Package{FilePath: path, ImportPath: mustConvertPathToImportPath(path)}
			pkgs = append(pkgs, pkg)
			if verbose {
				fmt.Fprintln(os.Stderr, "OK "+path+".")
			}
			// OK
			return nil
		}

		if verbose {
			fmt.Fprintln(os.Stderr, "IGNORE "+path+". There are no go files.")
		}

		// Don't skip dir. Subpackages may be exists.
		return nil
	})

	return pkgs
}

type Package struct {
	FilePath   string
	ImportPath string
}

type Dep struct {
	From *Package
	To   *Package
}

func (p *Package) Deps(stack []*Dep, trackingMode int) (deps []*Dep) {
	deps = stack

	depImportPaths, err := p.DepImportPaths()

	if err != nil {
		return
	}

	for _, depImportPath := range depImportPaths {
		depPkg := &Package{FilePath: mustConvertImportPathToPath(depImportPath), ImportPath: depImportPath}
		dep := &Dep{From: p, To: depPkg}

		if containsDep(deps, dep) {
			continue
		}

		deps = append(deps, dep)

		skipSubDeps := false

		switch trackingMode {
		case StopTracingOnReachingTopLevelDeps:
			skipSubDeps = true
		case StopTracingOnReachingStandardPackageOrLeaf:
			skipSubDeps = depPkg.StandardPkg()
		case StopTracingOnReachingLeaf:
			skipSubDeps = false
		default:
			panic(fmt.Sprintf("Unknown tracing-mode. %s\n", trackingMode))
		}

		if skipSubDeps {
			continue
		}

		// Skip if already registered.
		for _, depDep := range depPkg.Deps(deps, trackingMode) {
			if containsDep(deps, depDep) {
				continue
			}

			deps = append(deps, depDep)
		}
	}

	return
}

func (p *Package) StandardPkg() bool {
	segments := strings.Split(p.ImportPath, "/")

	if len(segments) > 0 {
		return strings.Index(segments[0], ".") == -1
	}

	return true
}

func (p *Package) DepImportPaths() ([]string, error) {
	pkg, err := build.Default.Import(p.ImportPath, "", build.AllowBinary)

	if err != nil {
		return nil, err
	}

	return pkg.Imports, nil
}

func mustConvertPathToImportPath(path string) (importPath string) {
	importPath = path

	// remove $GOROOT/src/pkg
	if strings.HasPrefix(importPath, filepath.Join(build.Default.GOROOT, "src", "pkg")) {
		importPath = strings.Replace(importPath, filepath.Join(build.Default.GOROOT, "src", "pkg"), "", 1)
	}

	// remove $GOPATH/src
	for _, gopath := range filepath.SplitList(build.Default.GOPATH) {
		// * see build.Context.gopath()
		if gopath == "" || gopath == build.Default.GOROOT {
			continue
		}

		// * see build.Context.gopath()
		if strings.HasPrefix(gopath, "~") {
			continue
		}

		if strings.HasPrefix(importPath, filepath.Join(gopath, "src")) {
			importPath = strings.Replace(importPath, filepath.Join(gopath, "src"), "", 1)
			break
		}
	}

	if path == importPath {
		panic(fmt.Errorf("mustConvertPathToImportPath: %v must be in $GOROOT nor $GOPATH.", path))
	}

	// remove slash of head
	if strings.HasPrefix(importPath, "/") {
		importPath = strings.Replace(importPath, "/", "", 1)
	}

	return
}

func mustConvertImportPathToPath(importPath string) (path string) {
	path = filepath.Join(importPath, filepath.Join(build.Default.GOROOT, "src"))

	_, err := os.Stat(path)

	if err != nil {
		return path
	}

	for _, gopath := range filepath.SplitList(build.Default.GOPATH) {
		// * see build.Context.gopath()
		if gopath == "" || gopath == build.Default.GOROOT {
			continue
		}

		// * see build.Context.gopath()
		if strings.HasPrefix(gopath, "~") {
			continue
		}

		path = filepath.Join(importPath, filepath.Join(gopath, "src", "pkg"))

		_, err := os.Stat(path)

		if err != nil {
			return path
		}
	}

	panic(fmt.Errorf("mustConvertImportPathPath: %v must be in $GOROOT nor $GOPATH.", path))
}

func containsDep(deps []*Dep, dep *Dep) bool {
	for _, a := range deps {
		if a.From.ImportPath == dep.From.ImportPath && a.To.ImportPath == dep.To.ImportPath {
			return true
		}
	}
	return false
}
