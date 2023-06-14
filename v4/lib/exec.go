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
	"strings"

	"modernc.org/opt"
	"modernc.org/strutil"
)

const (
	realCCEnvVar = "CCGO_EXEC_CC"
	cflagsEnvVar = "CCGO_EXEC_CFLAGS"
	cflagsSep    = "|"
)

func (t *Task) exec(args []string) (err error) {
	if s := os.Getenv(realCCEnvVar); s != "" {
		return fmt.Errorf("-fake: env var %s already set: %q", realCCEnvVar, s)
	}

	if t.execCC == "" {
		return fmt.Errorf("-fake: missing -fake-cc option")
	}

	if len(args) == 0 {
		return fmt.Errorf("-fake: missing command")
	}

	cc, err := exec.LookPath(t.execCC)
	if err != nil {
		return fmt.Errorf("-fake: %v", err)
	}

	if err := os.Setenv(realCCEnvVar, cc); err != nil {
		return fmt.Errorf("cannot set env var %s: %v", realCCEnvVar, err)
	}

	cflags := t.args[1 : (len(t.args))-len(args)-1] // -1 for the final "-fake"
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
		return fmt.Errorf("-fake not yet supported on Windows")
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
		dmesg("%v: ==== ENTER: wd %v, %v \\\n%v", origin(1), wd, err, t.args)
	}

	defer func() {
		if e := recover(); e != nil && err == nil {
			err = fmt.Errorf("PANIC: %v", e)
		}
		if err != nil {
			dmesg("%v: ==== EXIT FAIL: %v", origin(1), err)
			return
		}

		dmesg("%v: ==== EXIT OK:", origin(1))
	}()

	ccBase := filepath.Base(realCC)
	if len(t.args) == 0 || t.args[0] != ccBase {
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
	set.Arg("o", true, func(arg, val string) error { args.add(arg, val+".go"); return nil })
	set.Arg("std", true, func(arg, val string) error { args.add(fmt.Sprintf("%s=%s", arg, val)); return nil })
	set.Opt("c", func(arg string) error { args.add(arg); return nil })
	set.Opt("nostdinc", func(arg string) error { args.add(arg); return nil })
	set.Opt("pipe", func(arg string) error { return nil })
	if err := set.Parse(t.args[1:], func(arg string) error {
		if strings.HasPrefix(arg, "-f") { // eg. -ffreestanding
			return nil
		}

		if strings.HasPrefix(arg, "-W") { // eg. -Wa,--noexecstack
			return nil
		}

		if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("unexpected/unsupported option: %s", arg)
		}

		if strings.HasSuffix(arg, ".c") || strings.HasSuffix(arg, ".h") {
			args.add(arg)
			return nil
		}

		if strings.HasSuffix(arg, ".s") {
			return opt.Skip(nil)
		}

		return fmt.Errorf("unexpected/unsupported argument: %s", arg)
	}); err != nil {
		if _, ok := err.(opt.Skip); ok {
			return nil
		}

		return err
	}

	return NewTask(t.goos, t.goarch, args, t.stdout, t.stderr, t.fs).main()
}
