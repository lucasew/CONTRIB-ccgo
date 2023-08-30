// Copyright 2022 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ccgo implements the ccgo command.
package ccgo // import "modernc.org/ccgo/v4/lib"

//TODO TestSQLite linux/arm64
//TODO SYS_getsid macro missing
//TODO support hidden
//TODO Tucontext_t - Tucontext_t5
//TODO acosh u does not need to be pinned
//TODO tests += staticcheck
//TODO volatile handling of 'volatile struct vs s;', [0], pg. 73
//TODO add inlining infinite recursion protection

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
)

// Task represents a compilation job.
type Task struct {
	D                     []string // -D
	I                     []string // -I
	O                     string   // -O
	U                     []string // -U
	args                  []string // command name in args[0]
	cfg                   *cc.Config
	cfgArgs               []string
	compiledfFiles        map[string]string // *.c -> *.c.go
	defs                  string
	execCC                string // -exec-cc
	fs                    fs.FS
	goABI                 *gc.ABI
	goarch                string
	goos                  string
	hidden                nameSet  // -hide <string>
	idirafter             []string // -idirafter
	ignoreFile            nameSet  // -ignore-file=comma separated file list
	imports               []string // -import=comma separated import list
	inputFiles            []string
	iquote                []string // -iquote
	isystem               []string // -isystem
	l                     []string // -l
	linkFiles             []string
	o                     string   // -o
	packageName           string   // --package-name
	predef                []string // --predef
	prefixAnonType        string
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

	E                            bool // -E
	ansi                         bool // -ansi
	c                            bool // -c
	debugLinkerSave              bool // -debug-linker-save, causes pre type checking save of the linker result.
	freeStanding                 bool // -ffreestanding
	absolutePaths                bool // -absolute-paths
	fullPaths                    bool // -full-paths
	header                       bool // -header
	ignoreAsmErrors              bool // -ignore-asm-errors
	ignoreUnsupportedAligment    bool // -ignore-unsupported-alignment
	ignoreUnsupportedAtomicSizes bool // -ignore-unsupported-atomic-sizes
	ignoreVectorFunctions        bool // -ignore-vector-functions
	isExeced                     bool // -exec ...
	keepObjectFiles              bool // -keep-object-files
	noBuiltin                    bool // -fno-builtin
	noObjFmt                     bool // -no-object-file-format
	nostdinc                     bool // -nostdinc
	nostdlib                     bool // -nostdlib
	opt0                         bool // -O0
	packageNameSet               bool
	positions                    bool // -positions
	prefixDefineSet              bool // --prefix-define <string>
	pthread                      bool // -pthread
	strictISOMode                bool // -ansi or stc=c90
	verifyTypes                  bool // -verify-types
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
		prefixAnonType: "_",
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
		dmesg("%v: ==== enter %s CC=%q %s=%q(=realCC)", origin(1), t.args, os.Getenv("CC"), CCEnvVar, os.Getenv(CCEnvVar))
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
	set.Arg("-predef", false, func(arg, val string) error { t.predef = append(t.predef, val); return nil })
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
	set.Arg("D", true, func(arg, val string) error { t.D = append(t.D, fmt.Sprintf("%s%s", arg, val)); return nil })
	set.Arg("I", true, func(arg, val string) error { t.I = append(t.I, val); return nil })
	set.Arg("O", true, func(arg, val string) error { t.O = fmt.Sprintf("%s%s", arg, val); t.opt0 = val == "0"; return nil })
	set.Arg("U", true, func(arg, val string) error { t.U = append(t.U, fmt.Sprintf("%s%s", arg, val)); return nil })
	set.Arg("exec-cc", false, func(arg, val string) error { t.execCC = val; return nil })
	set.Arg("hide", false, func(arg, val string) error {
		for _, v := range strings.Split(val, ",") {
			t.hidden.add(v)
		}
		return nil
	})
	set.Arg("idirafter", true, func(arg, val string) error { t.idirafter = append(t.idirafter, val); return nil })
	set.Arg("ignore-file", false, func(arg, val string) error {
		for _, v := range strings.Split(val, ",") {
			t.ignoreFile.add(v)
		}
		return nil
	})
	set.Arg("import", false, func(arg, val string) error {
		t.imports = append(t.imports, strings.Split(val, ",")...)
		return nil
	})
	set.Arg("iquote", true, func(arg, val string) error { t.iquote = append(t.iquote, val); return nil })
	set.Arg("isystem", true, func(arg, val string) error { t.isystem = append(t.isystem, val); return nil })
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
	set.Opt("absolute-paths", func(arg string) error { t.absolutePaths = true; return nil })
	set.Opt("ansi", func(arg string) error { t.ansi = true; t.strictISOMode = true; return nil })
	set.Opt("c", func(arg string) error { t.c = true; return nil })
	set.Opt("debug-linker-save", func(arg string) error { t.debugLinkerSave = true; return nil })
	set.Opt("exec", func(arg string) error { return opt.Skip(nil) })
	set.Opt("extended-errors", func(arg string) error { extendedErrors = true; gc.ExtendedErrors = true; return nil })
	set.Opt("ffreestanding", func(arg string) error {
		t.freeStanding = true
		t.noBuiltin = true
		t.cfgArgs = append(t.cfgArgs, arg)
		return nil
	})
	set.Opt("fno-builtin", func(arg string) error { t.noBuiltin = true; t.cfgArgs = append(t.cfgArgs, arg); return nil })
	set.Opt("full-paths", func(arg string) error { t.fullPaths = true; return nil })
	set.Opt("header", func(arg string) error { t.header = true; return nil })
	set.Opt("ignore-asm-errors", func(arg string) error { t.ignoreAsmErrors = true; return nil })
	set.Opt("ignore-unsupported-alignment", func(arg string) error { t.ignoreUnsupportedAligment = true; return nil })
	set.Opt("ignore-unsupported-atomic-sizes", func(arg string) error { t.ignoreUnsupportedAtomicSizes = true; return nil })
	set.Opt("ignore-vector-functions", func(arg string) error { t.ignoreVectorFunctions = true; return nil })
	set.Opt("keep-object-files", func(arg string) error { t.keepObjectFiles = true; return nil })
	set.Opt("mlong-double-64", func(arg string) error { t.cfgArgs = append(t.cfgArgs, arg); return nil })
	set.Opt("no-object-file-format", func(arg string) error { t.noObjFmt = true; return nil })
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
			if dmesgs {
				dmesg("", errorf("unexpected/unsupported option: %q", arg))
			}
			return errorf("unexpected/unsupported option: %s", arg)
		}

		if t.ignoreFile.has(arg) {
			return nil
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

	if len(t.isystem) == 0 && !t.freeStanding {
		isystem, err := isystem(t.goos, t.goarch, defaultLibc)
		if err != nil {
			return err
		}

		if isystem != "" {
			t.isystem = []string{isystem}
			t.D = append(t.D, "-D_GNU_SOURCE")
		}
	}

	t.D = append(t.D, "-D__CCGO__")
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
	sv := os.Getenv("CC")
	if s := os.Getenv(CCEnvVar); s != "" {
		os.Setenv("CC", s)
	}
	cfg, err := cc.NewConfig(t.goos, t.goarch, t.cfgArgs...)
	os.Setenv("CC", sv)
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

	if t.nostdinc {
		cfg.HostIncludePaths = nil
		cfg.HostSysIncludePaths = nil
	}

	// --------------------------------------------------------------------
	// https://gcc.gnu.org/onlinedocs/gcc/Directory-Options.html
	//
	// Directories specified with -iquote apply only to the quote form of the
	// directive, #include "file". Directories specified with -I, -isystem, or
	// -idirafter apply to lookup for both the #include "file" and #include <file>
	// directives.
	//
	// You can specify any number or combination of these options on the command
	// line to search for header files in several directories. The lookup order is
	// as follows:

	cfg.IncludePaths = nil
	cfg.SysIncludePaths = nil

	// 1 For the quote form of the include directive, the directory of the current
	//   file is searched first.
	cfg.IncludePaths = append(cfg.IncludePaths, "")

	// 2 For the quote form of the include directive, the directories specified by
	//   -iquote options are searched in left-to-right order, as they appear on the
	//   command line.
	cfg.IncludePaths = append(cfg.IncludePaths, t.iquote...)

	// 3 Directories specified with -I options are scanned in left-to-right order.
	cfg.IncludePaths = append(cfg.IncludePaths, t.I...)
	cfg.SysIncludePaths = append(cfg.SysIncludePaths, t.I...)

	// 4 Directories specified with -isystem options are scanned in left-to-right
	//   order.
	cfg.IncludePaths = append(cfg.IncludePaths, t.isystem...)
	cfg.SysIncludePaths = append(cfg.SysIncludePaths, t.isystem...)

	// 5 Standard system directories are scanned.
	cfg.IncludePaths = append(cfg.IncludePaths, cfg.HostIncludePaths...)
	cfg.IncludePaths = append(cfg.IncludePaths, cfg.HostSysIncludePaths...)
	cfg.SysIncludePaths = append(cfg.SysIncludePaths, cfg.HostIncludePaths...)
	cfg.SysIncludePaths = append(cfg.SysIncludePaths, cfg.HostSysIncludePaths...)

	// 6 Directories specified with -idirafter options are scanned in left-to-right
	//   order.
	cfg.IncludePaths = append(cfg.IncludePaths, t.idirafter...)
	cfg.SysIncludePaths = append(cfg.SysIncludePaths, t.idirafter...)
	// --------------------------------------------------------------------

	t.defs = buildDefs(t.D, t.U)
	cfg.FS = t.fs
	t.cfg = cfg
	if t.E {
		for _, ifn := range t.inputFiles {
			if err := cc.Preprocess(cfg, sourcesFor(cfg, ifn, t), t.stdout); err != nil {
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

func sourcesFor(cfg *cc.Config, fn string, t *Task) (r []cc.Source) {
	predef := cfg.Predefined
	if len(t.predef) != 0 {
		predef += "\n" + strings.Join(t.predef, "\n")
	}
	sources := []cc.Source{
		{Name: "<predefined>", Value: predef},
		{Name: "<builtin>", Value: cc.Builtin},
	}
	if t.defs != "" {
		sources = append(sources, cc.Source{Name: "<command-line>", Value: t.defs})
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
			switch filepath.Ext(ifn) {
			case ".c":
				ofn = filepath.Base(ifn)
				ofn = ofn[:len(ofn)-len(".c")] + ".o.go"
			default:
				ofn = filepath.Base(ifn) + ".go"
			}
		}
		t.compiledfFiles[ifn] = ofn
		p.exec(func() error { return newCtx(t, p.eh).compile(ifn, ofn) })
	}
	return p.wait()
}
