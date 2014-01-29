# godeps

[![Build Status](https://travis-ci.org/ToQoz/godeps.png?branch=master)](https://travis-ci.org/ToQoz/godeps)

github.com/ToQoz/godeps reveal go pkg dependencies.

    $ cd a-go-pkg
    $ godeps .
    $ dot -Tpng godeps.dot -o godeps.png

You can embed to your project's README like this.

## godeps

![Dependencies graph](godeps.png?raw=true)

## TODO

godoc like web version. And it provide badge like travis-ci.

## See also

- http://golang.org/pkg/go/build/
- http://golang.org/cmd/go/#hdr-List_packages
