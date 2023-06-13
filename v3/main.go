// Copyright 2020 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main // import "modernc.org/ccgo/v3"

import (
	"fmt"
	"os"
	"runtime"

	ccgo3 "modernc.org/ccgo/v3/lib"
	ccgo4 "modernc.org/ccgo/v4/lib"
	_ "modernc.org/libc"
)

//TODO parallel

//TODO CPython
//TODO Cython
//TODO gmp
//TODO gofrontend
//TODO gsl
//TODO gtk
//TODO hdf5
//TODO minigmp
//TODO mpc
//TODO mpfr
//TODO pcre
//TODO pcre2
//TODO quickjs
//TODO redis
//TODO tcl/tk
//TODO wolfssl
//TODO zdat
//TODO zlib

func main() {
	var err error
	switch {
	case len(os.Args) > 1 && os.Args[1] == "-v4":
		goarch := env("TARGET_GOARCH", env("GOARCH", runtime.GOARCH))
		goos := env("TARGET_GOOS", env("GOOS", runtime.GOOS))
		err = ccgo4.NewTask(goos, goarch, append([]string{os.Args[0]}, os.Args[2:]...), os.Stdout, os.Stderr, nil).Main()
	case len(os.Args) > 1 && os.Args[1] == "-v3":
		err = ccgo3.NewTask(append([]string{os.Args[0]}, os.Args[2:]...), os.Stdout, os.Stderr).Main()
	default:
		err = ccgo3.NewTask(os.Args, os.Stdout, os.Stderr).Main()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func env(name, deflt string) (r string) {
	r = deflt
	if s := os.Getenv(name); s != "" {
		r = s
	}
	return r
}
