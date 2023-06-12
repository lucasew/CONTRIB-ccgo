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
)

const (
	realCCEnvVar = "CCGO_FAKE_CC"
)

func (t *Task) fake(args []string) (err error) {
	if s := os.Getenv(realCCEnvVar); s != "" {
		return fmt.Errorf("-fake: env var %s already set: %q", realCCEnvVar, s)
	}

	if t.fakeCC == "" {
		return fmt.Errorf("-fake: missing -fake-cc option")
	}

	if len(args) == 0 {
		return fmt.Errorf("-fake: missing command")
	}

	cc, err := exec.LookPath(t.fakeCC)
	if err != nil {
		return fmt.Errorf("-fake: %v", err)
	}

	if err := os.Setenv(realCCEnvVar, cc); err != nil {
		return fmt.Errorf("cannot set env var %s: %v", realCCEnvVar, err)
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

func (t *Task) faked(realCC string) (err error) {
	panic(todo("real CC=%q, args to ccgo=%q", realCC, t.args))
}
