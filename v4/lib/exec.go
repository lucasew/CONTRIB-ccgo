// Copyright 2022 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccgo // import "modernc.org/ccgo/v4/lib"

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"modernc.org/opt"
	"modernc.org/strutil"
)

const (
	// xxEnvVars contain the paths to the binaries ccgo acts as a proxy for when
	// -exec is in effect.
	AREnvVar      = "CCGO_EXEC_AR"
	CCEnvVar      = "CCGO_EXEC_CC"
	ClangEnvVar   = "CCGO_EXEC_CLANG"
	GCCEnvVar     = "CCGO_EXEC_GCC"
	LIBTOOLEnvVar = "CCGO_EXEC_LIBTOOL"
	LNEnvVar      = "CCGO_EXEC_LN"
	MVEnvVar      = "CCGO_EXEC_MV"
	RMEnvVar      = "CCGO_EXEC_RM"
	cflagsEnvVar  = "CCGO_EXEC_CFLAGS"
	cflagsSep     = "|"
)

var (
	execBins = []struct{ nm, envVar string }{
		{"ar", AREnvVar},
		{"cc", CCEnvVar},
		{"clang", ClangEnvVar},
		{"gcc", GCCEnvVar},
		{"libtool", LIBTOOLEnvVar},
		{"ln", LNEnvVar},
		{"mv", MVEnvVar},
		{"rm", RMEnvVar},
	}
)

func (t *Task) exec(args []string) (err error) {
	if len(args) == 0 {
		return errorf("-exec: missing command")
	}

	for _, v := range execBins {
		if s := os.Getenv(v.envVar); s != "" {
			return errorf("-exec: env var %s already set: %q", v.envVar, s)
		}

		bin, err := exec.LookPath(v.nm)
		if err == nil {
			if err := os.Setenv(v.envVar, bin); err != nil {
				return errorf("cannot set env var %s: %v", v.envVar, err)
			}
		}
	}

	if !IsExecEnv() {
		return errorf("no supported C compiler found")
	}

	cflags := t.args[1 : (len(t.args))-len(args)-1] // -1 for the final "-exec"
	if err := os.Setenv(cflagsEnvVar, strutil.JoinFields(cflags, cflagsSep)); err != nil {
		return errorf("cannot set env var %s: %v", cflagsEnvVar, err)
	}

	self, err := os.Executable()
	if err != nil {
		return err
	}

	dirTemp, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}

	defer os.RemoveAll(dirTemp)

	path := os.Getenv("PATH")

	defer os.Setenv("PATH", path)

	if err = os.Setenv("PATH", fmt.Sprintf("%s%c%s", dirTemp, os.PathListSeparator, path)); err != nil {
		return err
	}

	ln := os.Getenv(LNEnvVar)
	for _, v := range execBins {
		switch {
		case v.nm == "libtool" && t.goos != "darwin":
			continue
		}

		symlink := filepath.Join(dirTemp, v.nm)
		out, err := exec.Command(ln, self, symlink).CombinedOutput()
		if err != nil {
			return errorf("%s\n%v", out, err)
		}
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = t.stdout
	cmd.Stderr = t.stderr
	return cmd.Run()
}

type strSlice []string

func (s *strSlice) add(v ...string) { *s = append(*s, v...) }

func (t *Task) execed(cflags []string) (err error) {
	if dmesgs {
		wd, err := os.Getwd()
		dmesg("%v: ==== task.execed ENTER: wd (%v, %v), CC=%q %s=%q, %s=%q, %s=%q\\\n%v", origin(1), wd, err, os.Getenv("CC"), CCEnvVar, os.Getenv(CCEnvVar), GCCEnvVar, os.Getenv(GCCEnvVar), ClangEnvVar, os.Getenv(ClangEnvVar), t.args)
	}

	defer func() {
		if e := recover(); e != nil && err == nil {
			err = errorf("PANIC: %v\n%s", e, debug.Stack())
		}
		if err != nil {
			dmesg("%v: ==== EXIT FAIL: %v\n", origin(1), err)
			return
		}

		dmesg("%v: ==== EXIT OK:", origin(1))
	}()

	if len(t.args) == 0 {
		return errorf("internal error: real CC=%q, GCC=%q, Clang=%q, faked args=%q", t.realCC, t.realGCC, t.realClang, t.args)
	}

	switch t.noExe(filepath.Base(t.args[0])) {
	case t.noExe(filepath.Base(t.realAR)):
		return t.ar()
	case t.noExe(filepath.Base(t.realCC)):
		return t.cc(t.realCC, cflags)
	case t.noExe(filepath.Base(t.realGCC)):
		return t.cc(t.realGCC, cflags)
	case t.noExe(filepath.Base(t.realClang)):
		return t.cc(t.realClang, cflags)
	case t.noExe(filepath.Base(t.realMV)):
		return t.mv()
	case t.noExe(filepath.Base(t.realRM)):
		return t.rm()
	case t.noExe(filepath.Base(t.realLN)):
		return t.ln()
	case t.noExe(filepath.Base(t.realLIBTOOL)):
		return t.libtool()
	default:
		return errorf("internal error: realAR=%q realCC=%q, realGCC=%q, realClang=%q, t.args=%q", t.realAR, t.realCC, t.realGCC, t.realClang, t.args)
	}
}

func (t *Task) noExe(s string) string {
	const tag = ".exe"
	if t.goos != "windows" || !strings.HasSuffix(s, tag) {
		return s
	}

	return s[:len(s)-len(tag)]
}

func (t *Task) libtool() error {
	cmd := exec.Command(t.realLIBTOOL, t.args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if dmesgs {
			dmesg("NOTE: %s returns %v", t.realLIBTOOL, err.(*exec.ExitError).ExitCode())
		}
	}
	set := opt.NewSet()
	var args strSlice
	var outfn string
	set.Arg("o", true, func(arg, val string) error {
		if !strings.HasSuffix(val, ".a") {
			return errorf("unexpected -o argument: %s", val)
		}

		outfn = t.goFile(val)
		return nil
	})
	if err := set.Parse(t.args[1:], func(arg string) error {
		if strings.HasPrefix(arg, "-") {
			if dmesgs {
				dmesg("", errorf("unexpected/unsupported option: %q", arg))
			}
			return errorf("unexpected/unsupported option: %s", arg)
		}

		args.add(t.goFile(arg))
		return nil
	}); err != nil {
		return err
	}
	args2 := strSlice{"-cr", outfn}
	args2 = append(args2, args...)
	if dmesgs {
		dmesg("", args2)
	}
	cmd = exec.Command(t.realAR, args2...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if dmesgs {
			dmesg("SKIP2: %s returns %v", t.realAR, err.(*exec.ExitError).ExitCode())
		}
		return err
	}

	if dmesgs {
		dmesg("OK %v", args2)
	}
	return nil
}

func (t *Task) ln() error {
	cmd := exec.Command(t.realLN, t.args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if dmesgs {
			dmesg("NOTE: %s returns %v", t.realLN, err.(*exec.ExitError).ExitCode())
		}
	}
	set := opt.NewSet()
	var args []string
	files := 0
	set.Opt("s", func(arg string) error { args = append(args, arg); return nil })
	set.Opt("sf", func(arg string) error { args = append(args, arg); return nil })
	set.Opt("fs", func(arg string) error { args = append(args, arg); return nil })
	if err := set.Parse(t.args[1:], func(arg string) error {
		if strings.HasPrefix(arg, "-") {
			if dmesgs {
				dmesg("", errorf("unexpected/unsupported option: %q", arg))
			}
			return errorf("unexpected/unsupported option: %s", arg)
		}

		args = append(args, t.goFile(arg))
		files++
		return nil
	}); err != nil {
		return err
	}
	if files != 2 {
		return errorf("real LN=%q, faked args=%q", t.realLN, t.args)
	}

	if _, err := os.Stat(args[0]); err != nil {
		return nil
	}

	shell0(60*time.Second, true, t.realLN, args...)
	return nil
}

func (t *Task) mv() error {
	cmd := exec.Command(t.realMV, t.args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if dmesgs {
			dmesg("SKIP: %s returns %v", t.realMV, err.(*exec.ExitError).ExitCode())
		}
		return err
	}

	set := opt.NewSet()
	var args []string
	files := 0
	set.Opt("f", func(arg string) error { args = append(args, "-f"); return nil })
	if err := set.Parse(t.args[1:], func(arg string) error {
		if strings.HasPrefix(arg, "-") {
			if dmesgs {
				dmesg("", errorf("unexpected/unsupported option: %q", arg))
			}
			return errorf("unexpected/unsupported option: %s", arg)
		}

		args = append(args, t.goFile(arg))
		files++
		return nil
	}); err != nil {
		return err
	}

	if files != 2 {
		return errorf("real MV=%q, faked args=%q", t.realMV, t.args)
	}

	if _, err := os.Stat(args[0]); err != nil {
		return nil
	}

	shell0(60*time.Second, true, t.realMV, args...)
	return nil
}

func (t *Task) rm() error {
	cmd := exec.Command(t.realRM, t.args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if dmesgs {
			dmesg("SKIP: %s returns %v", t.realRM, err.(*exec.ExitError).ExitCode())
		}
		return err
	}

	rf := false
	set := opt.NewSet()
	set.Opt("r", func(arg string) error { return nil })
	set.Opt("f", func(arg string) error { return nil })
	set.Opt("rf", func(arg string) error { rf = true; return nil })
	set.Opt("fr", func(arg string) error { rf = true; return nil })
	return set.Parse(t.args[1:], func(arg string) error {
		if strings.HasPrefix(arg, "-") {
			if dmesgs {
				dmesg("", errorf("unexpected/unsupported option: %q", arg))
			}
			return errorf("unexpected/unsupported option: %s", arg)
		}

		switch {
		case rf:
			// nop
		default:
			os.Remove(t.goFile(arg))
		}
		return nil
	})
}

func (t *Task) goFile(s string) string {
	switch filepath.Ext(s) {
	case ".lo", ".o":
		return s + ".go"
	default:
		return s + "go"
	}
}

func (t *Task) cc(realCC string, cflags []string) error {
	cmd := exec.Command(realCC, t.args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if dmesgs {
			dmesg("NOTE: %s returns %v", t.realCC, err.(*exec.ExitError).ExitCode())
		}
	}

	optE := false
	args := append(strSlice{t.args[0]}, cflags...)
	set := opt.NewSet()
	ignore := 0
	set.Arg("-libc", false, func(arg, val string) error { args.add(fmt.Sprintf("%s=%s", arg, val)); return nil })
	set.Arg("D", true, func(arg, val string) error { args.add(arg + val); return nil })
	set.Arg("I", true, func(arg, val string) error { args.add(arg + val); return nil })
	set.Arg("L", true, func(arg, val string) error { args.add(arg + val); return nil })
	set.Arg("MD", true, func(arg, val string) error { return nil })
	set.Arg("MF", true, func(arg, val string) error { return nil })
	set.Arg("MT", true, func(arg, val string) error { return nil })
	set.Arg("O", true, func(arg, val string) error { args.add(arg + val); return nil })
	set.Arg("U", true, func(arg, val string) error { args.add(arg + val); return nil })
	set.Arg("current_version", false, func(arg, val string) error { return nil })
	set.Arg("gz", true, func(arg, val string) error { args.add(fmt.Sprintf("%s=%s", arg, val)); return nil })
	set.Arg("idirafter", true, func(arg, val string) error { args.add(fmt.Sprintf("%s=%s", arg, val)); return nil })
	set.Arg("iquote", true, func(arg, val string) error { args.add(fmt.Sprintf("%s=%s", arg, val)); return nil })
	set.Arg("isystem", true, func(arg, val string) error { args.add(fmt.Sprintf("%s=%s", arg, val)); return nil })
	set.Arg("l", true, func(arg, val string) error { args.add(arg + val); return nil })
	set.Arg("march", true, func(arg, val string) error { args.add(fmt.Sprintf("%s=%s", arg, val)); return nil })
	set.Arg("mtune", true, func(arg, val string) error { args.add(fmt.Sprintf("%s=%s", arg, val)); return nil })
	set.Arg("o", true, func(arg, val string) error { args.add(arg, val+".go"); return nil })
	set.Arg("sectcreate", false, func(arg, val string) error { ignore = 2; return nil })
	set.Arg("std", true, func(arg, val string) error { args.add(fmt.Sprintf("%s=%s", arg, val)); return nil })
	set.Opt("-version", func(arg string) error { args.add(arg); return nil })
	set.Opt("E", func(arg string) error { optE = true; return nil })
	set.Opt("MMD", func(arg string) error { return nil })
	set.Opt("Qunused-arguments", func(arg string) error { args.add(arg); return nil })
	set.Opt("c", func(arg string) error { args.add(arg); return nil })
	set.Opt("dumpmachine", func(arg string) error { args.add(arg); return nil })
	set.Opt("dynamiclib", func(arg string) error { return nil })
	set.Opt("ffreestanding", func(arg string) error { args.add(arg); return nil })
	set.Opt("fno-builtin", func(arg string) error { args.add(arg); return nil })
	set.Opt("g", func(arg string) error { return nil })
	set.Opt("headerpad_max_install_names", func(arg string) error { args.add(arg); return nil })
	set.Opt("ignore-link-errors", func(arg string) error { args.add(arg); return nil })
	set.Opt("m32", func(arg string) error { args.add(arg); return nil })
	set.Opt("m64", func(arg string) error { args.add(arg); return nil })
	set.Opt("mdynamic-no-pic", func(arg string) error { return nil })
	set.Opt("mlong-double-64", func(arg string) error { args.add(arg); return nil })
	set.Opt("mconsole", func(arg string) error { args.add(arg); return nil })
	set.Opt("municode", func(arg string) error { args.add(arg); return nil })
	set.Opt("nostdinc", func(arg string) error { args.add(arg); return nil })
	set.Opt("nostdlib", func(arg string) error { args.add(arg); return nil })
	set.Opt("pedantic", func(arg string) error { args.add(arg); return nil })
	set.Opt("pedantic-errors", func(arg string) error { args.add(arg); return nil })
	set.Opt("pipe", func(arg string) error { return nil })
	set.Opt("s", func(arg string) error { args.add(arg); return nil })
	set.Opt("shared", func(arg string) error { args.add(arg); return nil })
	set.Opt("static", func(arg string) error { args.add(arg); return nil })
	set.Opt("static-libgcc", func(arg string) error { args.add(arg); return nil })
	set.Opt("v", func(arg string) error { args.add(arg); return nil })
	set.Opt("w", func(arg string) error { args.add(arg); return nil })
	files := 0
	if err := set.Parse(t.args[1:], func(arg string) error {
		if ignore > 0 {
			ignore--
			return nil
		}

		if optE {
			return nil
		}

		if strings.HasPrefix(arg, "-f") {
			return nil
		}

		if strings.HasPrefix(arg, "-W") { // eg. -Wa,--noexecstack
			return nil
		}

		if strings.HasPrefix(arg, "-") {
			if dmesgs {
				dmesg("", errorf("unexpected/unsupported option: %q", arg))
			}
			return errorf("unexpected/unsupported option: %s", arg)
		}

		switch filepath.Ext(arg) {
		case ".c", ".h":
			args.add(arg)
			files++
			return nil
		case ".s", ".S":
			return nil
		case ".o", ".lo":
			nm := arg + ".go"
			nm2 := ""
			if strings.HasSuffix(arg, ".lo") {
				nm2 = arg[:len(arg)-len(".lo")] + ".o.go"
			}
			switch {
			case t.fs != nil:
				if _, err := t.fs.Open(nm); err != nil {
					nm = nm2
					if _, err := t.fs.Open(nm); err != nil {
						return nil
					}
				}
			default:
				if _, err := os.Stat(nm); err != nil {
					nm = nm2
					if _, err := os.Stat(nm); err != nil {
						return nil
					}
				}
			}
			args.add(nm)
			files++
			return nil
		case ".a":
			args.add(arg)
			files++
			return nil
		}

		return errorf("unexpected/unsupported argument: %s", arg)
	}); err != nil {
		return err
	}

	if files == 0 || optE {
		return nil
	}

	t = NewTask(t.goos, t.goarch, args, t.stdout, t.stderr, t.fs)
	t.isExeced = true
	return t.main()
}

func (t *Task) ar() error {
	if dmesgs {
		dmesg("AR %v", t.args)
	}
	cmd := exec.Command(t.realAR, t.args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if dmesgs {
			dmesg("SKIP: %s returns %v", t.realAR, err.(*exec.ExitError).ExitCode())
		}
		return err
	}

	set := opt.NewSet()
	var argN, members int
	args := strSlice{t.args[0]}
	if err := set.Parse(t.args[1:], func(arg string) error {
		if strings.HasPrefix(arg, "-") {
			if dmesgs {
				dmesg("", errorf("unexpected/unsupported option: %q", arg))
			}
			return errorf("unexpected/unsupported option: %s", arg)
		}

		argN++
		switch argN {
		case 1: // keyletters
			var out string
			for _, c := range arg {
				switch sc := string(c); sc {
				case
					"c", // create the archive
					"q", // quick append
					"r", // insert member
					"u": // update

					out += sc
				case "s": // add index
					// nop
				default:
					return errorf("TODO #%d: %q: real AR=%q, faked args=%q", argN, arg, t.realAR, t.args)
				}
			}
			args.add(out)
			return nil
		case 2: // archive name
			if !strings.HasSuffix(arg, ".a") {
				return errorf("TODO #%d: %q: real AR=%q, faked args=%q", argN, arg, t.realAR, t.args)
			}

			args.add(arg + "go") // archive.ago
			return nil
		default:
			basenames := map[string]string{} // base: path
			switch filepath.Ext(arg) {
			case ".lo", ".o":
				nm := arg + ".go"
				if _, err := os.Stat(nm); err == nil {
					bn := filepath.Base(nm)
					if ex, ok := basenames[bn]; ok {
						return errorf("duplicate basename %s: %s", ex, nm)
					}

					members++
					args.add(nm)
				}
				return nil
			default:
				return errorf("TODO #%d: %q: real AR=%q, faked args=%q", argN, arg, t.realAR, t.args)
			}
		}

		return errorf("unexpected/unsupported argument: %s", arg)
	}); err != nil {
		return err
	}

	cmd = exec.Command(t.realAR, []string(args[1:])...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if dmesgs {
			dmesg("SKIP2: %s returns %v", t.realAR, err.(*exec.ExitError).ExitCode())
		}
		return err
	}

	return nil
}
