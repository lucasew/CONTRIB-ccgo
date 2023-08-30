// Copyright 2022 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccgo // import "modernc.org/ccgo/v4/lib"

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"modernc.org/opt"
	"modernc.org/strutil"
)

const (
	// CCEnvVar contains the path to the C compiler ccgo acts as a proxy
	// for when -exec is in effect.
	CCEnvVar     = "CCGO_EXEC_CC"
	cflagsEnvVar = "CCGO_EXEC_CFLAGS"
	cflagsSep    = "|"
)

func (t *Task) exec(args []string) (err error) {
	if s := os.Getenv(CCEnvVar); s != "" {
		return fmt.Errorf("-exec: env var %s already set: %q", CCEnvVar, s)
	}

	if t.execCC == "" {
		return fmt.Errorf("-exec: missing -exec-cc option")
	}

	if len(args) == 0 {
		return fmt.Errorf("-exec: missing command")
	}

	cc, err := exec.LookPath(t.execCC)
	if err != nil {
		return fmt.Errorf("-exec: %v", err)
	}

	if err := os.Setenv(CCEnvVar, cc); err != nil {
		return fmt.Errorf("cannot set env var %s: %v", CCEnvVar, err)
	}

	cflags := t.args[1 : (len(t.args))-len(args)-1] // -1 for the final "-exec"
	if err := os.Setenv(cflagsEnvVar, strutil.JoinFields(cflags, cflagsSep)); err != nil {
		return fmt.Errorf("cannot set env var %s: %v", cflagsEnvVar, err)
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

	switch runtime.GOOS {
	case "windows":
		return fmt.Errorf("-exec not yet supported on Windows")
	default:
		symlink := filepath.Join(dirTemp, filepath.Base(cc))
		path := os.Getenv("PATH")

		defer os.Setenv("PATH", path)

		if err = os.Setenv("PATH", fmt.Sprintf("%s%c%s", dirTemp, os.PathListSeparator, path)); err != nil {
			return err
		}

		out, err := exec.Command("ln", self, symlink).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s\n%v", out, err)
		}
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = t.stdout
	cmd.Stderr = t.stderr
	return cmd.Run()
}

type strSlice []string

func (s *strSlice) add(v ...string) { *s = append(*s, v...) }

func (t *Task) execed(realCC string, cflags []string) (err error) {
	if dmesgs {
		wd, err := os.Getwd()
		dmesg("%v: ==== ENTER: wd (%v, %v), CC=%q %s=%q(=realCC)\\\n%v", origin(1), wd, err, os.Getenv("CC"), CCEnvVar, os.Getenv(CCEnvVar), t.args)
	}

	defer func() {
		if e := recover(); e != nil && err == nil {
			err = fmt.Errorf("PANIC: %v\n%s", e, debug.Stack())
		}
		if err != nil {
			dmesg("%v: ==== EXIT FAIL: %v\n", origin(1), err)
			return
		}

		dmesg("%v: ==== EXIT OK:", origin(1))
	}()

	if len(t.args) == 0 || filepath.Base(t.args[0]) != filepath.Base(realCC) {
		return fmt.Errorf("%v: internal error: real CC=%q, faked args=%q", origin(1), realCC, t.args)
	}

	cmd := exec.Command(realCC, t.args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if dmesgs {
			dmesg("SKIP: host C compiler returns %v", err.(*exec.ExitError).ExitCode())
		}
		return err
	}

	args := append(strSlice{t.args[0]}, cflags...)
	set := opt.NewSet()
	set.Arg("D", true, func(arg, val string) error { args.add(arg + val); return nil })
	set.Arg("I", true, func(arg, val string) error { args.add(arg + val); return nil })
	set.Arg("O", true, func(arg, val string) error { args.add(arg + val); return nil })
	set.Arg("U", true, func(arg, val string) error { args.add(arg + val); return nil })
	set.Arg("idirafter", true, func(arg, val string) error { args.add(fmt.Sprintf("%s=%s", arg, val)); return nil })
	set.Arg("iquote", true, func(arg, val string) error { args.add(fmt.Sprintf("%s=%s", arg, val)); return nil })
	set.Arg("isystem", true, func(arg, val string) error { args.add(fmt.Sprintf("%s=%s", arg, val)); return nil })
	set.Arg("l", true, func(arg, val string) error { return nil })
	set.Arg("o", true, func(arg, val string) error { args.add(arg, val+".go"); return nil })
	set.Arg("std", true, func(arg, val string) error { args.add(fmt.Sprintf("%s=%s", arg, val)); return nil })
	set.Opt("c", func(arg string) error { args.add(arg); return nil })
	set.Opt("ffreestanding", func(arg string) error { args.add(arg); return nil })
	set.Opt("fno-builtin", func(arg string) error { args.add(arg); return nil })
	set.Opt("g", func(arg string) error { return nil })
	set.Opt("mlong-double-64", func(arg string) error { args.add(arg); return nil })
	set.Opt("nostdinc", func(arg string) error { args.add(arg); return nil })
	set.Opt("nostdlib", func(arg string) error { args.add(arg); return nil })
	set.Opt("pipe", func(arg string) error { return nil })
	set.Opt("shared", func(arg string) error { args.add(arg); return nil })
	files := 0
	if err := set.Parse(t.args[1:], func(arg string) error {
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
		case ".s":
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
		}

		return fmt.Errorf("unexpected/unsupported argument: %s", arg)
	}); err != nil {
		if _, ok := err.(opt.Skip); ok {
			return nil
		}

		return err
	}

	if files == 0 {
		return nil
	}

	t = NewTask(t.goos, t.goarch, args, t.stdout, t.stderr, t.fs)
	t.isExeced = true
	return t.main()
}
