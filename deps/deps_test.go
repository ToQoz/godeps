package deps

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDeps_WhenCircularReferenceExist(t *testing.T) {
	wd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	func() {
		// circularref -> circularref/b
		// circularref/b -> circularref
		pkgs := Packages(filepath.Join(wd, "testdata/circularref"))

		if len(pkgs) != 1 {
			t.Errorf("Unexpected package count: %v, %v expected", len(pkgs), 1)
			return
		}

		deps := pkgs[0].Deps(nil)

		if len(deps) != 2 {
			fmt.Printf("%q\n", deps)
			t.Errorf("Unexpected dep count: %v, %v expected", len(deps), 2)
			return
		}
	}()

	func() {
		// circularref2 -> circularref2/z/zb
		// circularref2/z/zb -> circularref2/z
		// circularref2/z -> circularref2/z/zb
		pkgs := Packages(filepath.Join(wd, "testdata/circularref2"))

		if len(pkgs) != 1 {
			t.Errorf("Unexpected package count: %v, %v expected", len(pkgs), 1)
			return
		}

		deps := pkgs[0].Deps(nil)

		if len(deps) != 3 {
			fmt.Printf("%v\n", deps)
			t.Errorf("Unexpected dep count: %v, %v expected", len(deps), 3)
			return
		}
	}()
}
