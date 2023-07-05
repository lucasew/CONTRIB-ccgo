// Copyright 2022 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ccgo implements the ccgo command.
package ccgo // import "modernc.org/ccgo/v4/lib"

//TODO SYS_getsid macro missing
//TODO support hidden
//TODO else { if ... } -> else if
//TODO TNucontext_t - TNucontext_t5
//TODO s/break; fallthrough//
//TODO s/goto <label>; fallthrough/goto <label>/

//  [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf

// -export-X, -unexport-X flags

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"modernc.org/cc/v4"
	"modernc.org/gc/v2"
	"modernc.org/opt"
	"modernc.org/strutil"
)

var (
	oTraceL = flag.Bool("trcl", false, "Print produced object files.")
	oTraceG = flag.Bool("trcg", false, "Print produced Go files.")

	isTesting bool
)

// Task represents a compilation job.
type Task struct {
	D              []string // -D
	I              []string // -I
	O              string   // -O
	U              []string // -U
	args           []string // command name in args[0]
	cfg            *cc.Config
	cfgArgs        []string
	compiledfFiles map[string]string // *.c -> *.c.go
	defs           string
	execCC         string // -exec-cc
	fs             fs.FS
	goABI          *gc.ABI
	goarch         string
	goos           string
	inputFiles     []string
	l              []string // -l
	linkFiles      []string
	o              string // -o
	packageName    string // --package-name
	//TODO prefixUnpinned        string // --prefix-unpinned <string>
	prefixAutomatic       string // --prefix-automatic <string>
	prefixCcgoAutomatic   string
	prefixDefine          string // --prefix-define <string>
	prefixEnumerator      string // --prefix-enumerator <string>
	prefixExternal        string // --prefix-external <string>
	prefixField           string // --prefix-field <string>
	prefixImportQualifier string // --prefix-import-qualifier <string>
	prefixMacro           string // --prefix-macro <string>
	prefixStaticInternal  string // --prefix-static-internal <string>
	prefixStaticNone      string // --prefix-static-none <string>
	prefixTaggedEnum      string // --prefix-tagfed-enum <string>
	prefixTaggedStruct    string // --prefix-tagged-struct <string>
	prefixTaggedUnion     string // --prefix-taged-union <string>
	prefixTypename        string // --prefix-typename <string>
	prefixUndefined       string // --prefix-undefined <string>
	std                   string // -std
	stderr                io.Writer
	stdout                io.Writer
	tlsQualifier          string

	intSize int

	E                         bool // -E
	ansi                      bool // -ansi
	c                         bool // -c
	debugLinkerSave           bool // -debug-linker-save, causes pre type checking save of the linker result.
	freeStanding              bool // -ffreestanding
	fullPaths                 bool // -full-paths
	header                    bool // -header
	ignoreAsmErrors           bool // -ignore-asm-errors
	ignoreHeaderFunctions     bool // -ignore-header-functions
	ignoreUnsupportedAligment bool // -ignore-unsupported-alignment
	ignoreVectorFunctions     bool // -ignore-vector-functions
	isExeced                  bool // -exec ...
	keepObjectFiles           bool // -keep-object-files
	noBuiltin                 bool // -fno-builtin
	noObjFmt                  bool // -no-object-file-format
	nostdinc                  bool // -nostdinc
	nostdlib                  bool // -nostdlib
	opt0                      bool // -O0
	packageNameSet            bool
	positions                 bool // -positions
	prefixDefineSet           bool // --prefix-define <string>
	pthread                   bool // -pthread
	strictISOMode             bool // -ansi or stc=c90
	verifyTypes               bool // -verify-types
}

// NewTask returns a newly created Task. args[0] is the command name.
func NewTask(goos, goarch string, args []string, stdout, stderr io.Writer, fs fs.FS) (r *Task) {
	var d []string
	switch goarch {
	case "arm", "386":
		// modernc.org/libc@v1/sys/types/Off_t is 64 bit
		d = []string{"-D_FILE_OFFSET_BITS=64"}
	}

	intSize := 8
	switch goarch {
	case "arm", "386":
		intSize = 4
	}
	return &Task{
		D:              d,
		args:           args,
		compiledfFiles: map[string]string{},
		fs:             fs,
		goarch:         goarch,
		goos:           goos,
		intSize:        intSize,
		stderr:         stderr,
		stdout:         stdout,
		tlsQualifier:   tag(importQualifier) + "libc.",
	}
}

// Main executes task.
func (t *Task) Main() (err error) {
	if realCC := os.Getenv(CCEnvVar); realCC != "" {
		var flags []string
		if cflags := os.Getenv(cflagsEnvVar); cflags != "" {
			flags = strutil.SplitFields(cflags, cflagsSep)
		}
		return t.execed(realCC, flags)
	}

	return t.main()
}

func (t *Task) main() (err error) {
	if dmesgs {
		dmesg("%v: ==== enter %s", origin(1), t.args)
		defer func() {
			dmesg("%v: ==== exit: %v", origin(1), err)
		}()
	}

	switch len(t.args) {
	case 0:
		return errorf("invalid arguments")
	case 1:
		return errorf("no input files")
	}

	if t.goABI, err = gc.NewABI(t.goos, t.goarch); err != nil {
		return errorf("%v", err)
	}

	set := opt.NewSet()
	set.Arg("-package-name", false, func(arg, val string) error { t.packageName = val; t.packageNameSet = true; return nil })
	set.Arg("-prefix-automatic", false, func(arg, val string) error { t.prefixAutomatic = val; return nil })
	set.Arg("-prefix-define", false, func(arg, val string) error { t.prefixDefine = val; t.prefixDefineSet = true; return nil })
	set.Arg("-prefix-enumerator", false, func(arg, val string) error { t.prefixEnumerator = val; return nil })
	set.Arg("-prefix-external", false, func(arg, val string) error { t.prefixExternal = val; return nil })
	set.Arg("-prefix-field", false, func(arg, val string) error { t.prefixField = val; return nil })
	set.Arg("-prefix-import-qualifier", false, func(arg, val string) error { t.prefixImportQualifier = val; return nil })
	set.Arg("-prefix-macro", false, func(arg, val string) error { t.prefixMacro = val; return nil })
	set.Arg("-prefix-static-internal", false, func(arg, val string) error { t.prefixStaticInternal = val; return nil })
	set.Arg("-prefix-static-none", false, func(arg, val string) error { t.prefixStaticNone = val; return nil })
	set.Arg("-prefix-tagged-enum", false, func(arg, val string) error { t.prefixTaggedEnum = val; return nil })
	set.Arg("-prefix-tagged-struct", false, func(arg, val string) error { t.prefixTaggedStruct = val; return nil })
	set.Arg("-prefix-tagged-union", false, func(arg, val string) error { t.prefixTaggedUnion = val; return nil })
	set.Arg("-prefix-typename", false, func(arg, val string) error { t.prefixTypename = val; return nil })
	set.Arg("-prefix-undefined", false, func(arg, val string) error { t.prefixUndefined = val; return nil })
	//TODO set.Arg("-prefix-unpinned", false, func(arg, val string) error { t.prefixUnpinned = val; return nil })
	set.Arg("D", true, func(arg, val string) error { t.D = append(t.D, fmt.Sprintf("%s%s", arg, val)); return nil })
	set.Arg("I", true, func(arg, val string) error { t.I = append(t.I, val); return nil })
	set.Arg("O", true, func(arg, val string) error { t.O = fmt.Sprintf("%s%s", arg, val); t.opt0 = val == "0"; return nil })
	set.Arg("U", true, func(arg, val string) error { t.U = append(t.U, fmt.Sprintf("%s%s", arg, val)); return nil })
	set.Arg("exec-cc", false, func(arg, val string) error { t.execCC = val; return nil })
	set.Arg("l", true, func(arg, val string) error {
		t.l = append(t.l, val)
		t.linkFiles = append(t.linkFiles, arg+"="+val)
		return nil
	})
	set.Arg("o", true, func(arg, val string) error { t.o = val; return nil })
	set.Arg("std", true, func(arg, val string) error {
		t.std = fmt.Sprintf("%s=%s", arg, val)
		if val == "c90" {
			t.strictISOMode = true
		}
		return nil
	})
	set.Opt("E", func(arg string) error { t.E = true; return nil })
	set.Opt("ansi", func(arg string) error { t.ansi = true; t.strictISOMode = true; return nil })
	set.Opt("c", func(arg string) error { t.c = true; return nil })
	set.Opt("debug-linker-save", func(arg string) error { t.debugLinkerSave = true; return nil })
	set.Opt("exec", func(arg string) error { return opt.Skip(nil) })
	set.Opt("extended-errors", func(arg string) error { extendedErrors = true; gc.ExtendedErrors = true; return nil })
	set.Opt("ffreestanding", func(arg string) error { t.freeStanding = true; t.cfgArgs = append(t.cfgArgs, arg); return nil })
	set.Opt("fno-builtin", func(arg string) error { t.noBuiltin = true; t.cfgArgs = append(t.cfgArgs, arg); return nil })
	set.Opt("full-paths", func(arg string) error { t.fullPaths = true; return nil })
	set.Opt("header", func(arg string) error { t.header = true; return nil })
	set.Opt("ignore-asm-errors", func(arg string) error { t.ignoreAsmErrors = true; return nil })
	set.Opt("ignore-header-functions", func(arg string) error { t.ignoreHeaderFunctions = true; return nil })
	set.Opt("ignore-unsupported-alignment", func(arg string) error { t.ignoreUnsupportedAligment = true; return nil })
	set.Opt("ignore-vector-functions", func(arg string) error { t.ignoreVectorFunctions = true; return nil })
	set.Opt("keep-object-files", func(arg string) error { t.keepObjectFiles = true; return nil })
	set.Opt("mlong-double-64", func(arg string) error { t.cfgArgs = append(t.cfgArgs, arg); return nil })
	set.Opt("no-object-file-format", func(arg string) error { t.noObjFmt = true; return nil }) // now ignored
	set.Opt("nostdinc", func(arg string) error { t.nostdinc = true; t.cfgArgs = append(t.cfgArgs, arg); return nil })
	set.Opt("nostdlib", func(arg string) error { t.nostdlib = true; return nil })
	set.Opt("positions", func(arg string) error { t.positions = true; return nil })
	set.Opt("pthread", func(arg string) error { t.pthread = true; t.cfgArgs = append(t.cfgArgs, arg); return nil })
	set.Opt("verify-types", func(arg string) error { t.verifyTypes = true; return nil })

	// Ignored
	set.Arg("MF", true, func(arg, val string) error { return nil })
	set.Arg("MQ", true, func(arg, val string) error { return nil })
	set.Arg("MT", true, func(arg, val string) error { return nil })
	set.Opt("M", func(arg string) error { return nil })
	set.Opt("MD", func(arg string) error { return nil })
	set.Opt("MM", func(arg string) error { return nil })
	set.Opt("MMD", func(arg string) error { return nil })
	set.Opt("MP", func(arg string) error { return nil })
	set.Opt("Qunused-arguments", func(arg string) error { return nil })
	set.Opt("S", func(arg string) error { return nil })
	set.Opt("dynamiclib", func(arg string) error { return nil })
	set.Opt("herror_on_warning", func(arg string) error { return nil })
	set.Opt("pedantic", func(arg string) error { return nil })
	set.Opt("pipe", func(arg string) error { return nil })
	set.Opt("s", func(arg string) error { return nil })
	set.Opt("shared", func(arg string) error { return nil })
	set.Opt("static", func(arg string) error { return nil })
	set.Opt("w", func(arg string) error { return nil })

	if err := set.Parse(t.args[1:], func(arg string) error {
		if strings.HasPrefix(arg, "-") {
			return errorf(" unrecognized command-line option '%s'", arg)
		}

		if strings.HasSuffix(arg, ".c") || strings.HasSuffix(arg, ".h") {
			t.inputFiles = append(t.inputFiles, arg)
			t.linkFiles = append(t.linkFiles, arg)
			return nil
		}

		if strings.HasSuffix(arg, ".go") {
			t.linkFiles = append(t.linkFiles, arg)
			return nil
		}

		return errorf("unexpected argument %s", arg)
	}); err != nil {
		switch x := err.(type) {
		case opt.Skip:
			return t.exec([]string(x))
		default:
			return errorf("parsing %v: %v", t.args[1:], err)
		}
	}

	t.cfgArgs = append(t.cfgArgs, t.D...)
	t.cfgArgs = append(t.cfgArgs, t.U...)
	t.cfgArgs = append(t.cfgArgs,
		t.O,
		t.std,
	)

	ldflag := cc.LongDouble64Flag(t.goos, t.goarch)
	if ldflag != "" {
		t.cfgArgs = append(t.cfgArgs, ldflag)
	}

	if t.goos == "windows" && (t.goarch == "386" || t.goarch == "amd64") {
		t.cfgArgs = append(t.cfgArgs,
			"-mno-3dnow",
			"-mno-abm",
			"-mno-aes",
			"-mno-avx",
			"-mno-avx2",
			"-mno-avx512cd",
			"-mno-avx512er",
			"-mno-avx512f",
			"-mno-avx512pf",
			"-mno-bmi",
			"-mno-bmi2",
			"-mno-f16c",
			"-mno-fma",
			"-mno-fma4",
			"-mno-fsgsbase",
			"-mno-lwp",
			"-mno-lzcnt",
			"-mno-mmx",
			"-mno-pclmul",
			"-mno-popcnt",
			"-mno-prefetchwt1",
			"-mno-rdrnd",
			"-mno-sha",
			"-mno-sse",
			"-mno-sse2",
			"-mno-sse3",
			"-mno-sse4",
			"-mno-sse4.1",
			"-mno-sse4.2",
			"-mno-sse4a",
			"-mno-ssse3",
			"-mno-tbm",
			"-mno-xop",
		)
	}

	// trc("", t.cfgArgs)
	cfg, err := cc.NewConfig(t.goos, t.goarch, t.cfgArgs...)
	if err != nil {
		return err
	}

	if ldflag == "" {
		if err = cfg.AdjustLongDouble(); err != nil {
			return err
		}
	}

	if t.header {
		cfg.Header = true
	}
	t.I = t.I[:len(t.I):len(t.I)]
	cfg.IncludePaths = append([]string{""}, t.I...)
	cfg.IncludePaths = append(cfg.IncludePaths, cfg.HostIncludePaths...)
	cfg.IncludePaths = append(cfg.IncludePaths, cfg.HostSysIncludePaths...)
	cfg.SysIncludePaths = append(t.I, cfg.HostSysIncludePaths...)
	t.defs = buildDefs(t.D, t.U)
	// trc("", t.defs)
	cfg.FS = t.fs
	t.cfg = cfg
	if t.E {
		for _, ifn := range t.inputFiles {
			if err := cc.Preprocess(cfg, sourcesFor(cfg, ifn, t.defs), t.stdout); err != nil {
				return err
			}
		}
		return nil
	}

	if t.nostdlib || t.freeStanding {
		t.tlsQualifier = ""
	}
	if t.c {
		return t.compile(t.o)
	}

	if !t.nostdlib && !t.freeStanding {
		t.linkFiles = append(t.linkFiles, fmt.Sprintf("-l=%s", defaultLibc))
	}
	return t.link()
}

func sourcesFor(cfg *cc.Config, fn, defs string) (r []cc.Source) {
	sources := []cc.Source{
		{Name: "<predefined>", Value: cfg.Predefined},
		{Name: "<builtin>", Value: cc.Builtin},
	}
	if defs != "" {
		sources = append(sources, cc.Source{Name: "<command-line>", Value: defs})
	}
	return append(sources, cc.Source{Name: fn, FS: cfg.FS})
}

// -c
func (t *Task) compile(optO string) error {
	switch len(t.inputFiles) {
	case 0:
		return errorf("no input files")
	case 1:
		// ok
	default:
		if t.o != "" && t.c {
			return errorf("cannot specify '-o' with '-c' with multiple files")
		}
	}

	p := newParallel("")
	for _, ifn := range t.inputFiles {
		ifn := ifn
		ofn := optO
		if ofn == "" {
			ofn = filepath.Base(ifn) + ".go"
		}
		t.compiledfFiles[ifn] = ofn
		p.exec(func() error { return newCtx(t, p.eh).compile(ifn, ofn) })
	}
	return p.wait()
}
