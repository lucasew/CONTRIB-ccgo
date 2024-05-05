// Copyright 2022 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccgo // import "modernc.org/ccgo/v4/lib"

var testExecKnownFails = map[string]struct{}{
	// ==== EXEC FAIL - compiles and builds but fails when executed.

	// // Won't fix: setjmp/longjmp
	// `assets/github.com/vnmakarov/mir/c-benchmarks/except.c`: {},
	// `assets/github.com/vnmakarov/mir/c-tests/new/setjmp.c`:  {},
	// `assets/github.com/vnmakarov/mir/c-tests/new/setjmp2.c`: {},

	// // Won't fix: sigfpe
	// `assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/execute/20101011-1.c`:                    {},
	// `assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/execute/ieee/fp-cmp-1.c`:                 {},
	// `assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/execute/ieee/fp-cmp-2.c`:                 {},
	// `assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/execute/ieee/fp-cmp-3.c`:                 {},
	// `assets/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/execute/20101011-1.c`:    {},
	// `assets/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/execute/ieee/fp-cmp-1.c`: {},
	// `assets/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/execute/ieee/fp-cmp-2.c`: {},
	// `assets/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/execute/ieee/fp-cmp-3.c`: {},

	// // Won't fix: Architecture specific conversion overflow.
	// `assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/execute/20031003-1.c`:                 {},
	// `assets/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/execute/20031003-1.c`: {},

	// // Won't fix: implementation defined bit fields
	// `assets/github.com/vnmakarov/mir/c-tests/lacc/bitfield-basic.c`:         {},
	// `assets/github.com/vnmakarov/mir/c-tests/lacc/bitfield-pack-next.c`:     {},
	// `assets/github.com/vnmakarov/mir/c-tests/lacc/bitfield-trailing-zero.c`: {},
	// `assets/github.com/vnmakarov/mir/c-tests/lacc/bitfield-types-init.c`:    {},

	// // Won't fix: long double
	// `assets/github.com/vnmakarov/mir/c-tests/lacc/long-double-load.c`: {},

	// // Won't fix: return_addr
	// `assets/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/execute/return-addr.c`: {},

	// // Won't fix, puts particular return value isn't specified
	// `assets/github.com/vnmakarov/mir/c-tests/lacc/macro-paste.c`: {},
	// `assets/github.com/vnmakarov/mir/c-tests/lacc/whitespace.c`:  {},

	//TODO
	`assets/CompCert-3.6/test/c/mandelbrot.c`:               {}, // EXEC FAIL
	`assets/benchmarksgame-team.pages.debian.net/fasta-3.c`: {}, // EXEC FAIL: "assets/benchmarksgame-team.pages.debian.net/fasta-3.c: >ONE Homo sapiens alu"
	`assets/benchmarksgame-team.pages.debian.net/fasta-8.c`: {}, // EXEC FAIL: "assets/benchmarksgame-team.pages.debian.net/fasta-8.c: >ONE Homo sapiens alu"

	// ==== BUILD FAIL - compiles but does not build.

	// ==== COMPILE FAIL - does not compile.

	`assets/benchmarksgame-team.pages.debian.net/mandelbrot-3.c`: {}, // COMPILE FAIL: "\"mandelbrot-3.c:27:21: unsupported vector type: v2df (expr.go:541:expr0: expr.go:4174:primaryExpression: expr.go:4646:primaryExpressionFloatConst: type.go:60:verifyTyp: type.go:65:typ0: type.go:440:is..."
	`assets/benchmarksgame-team.pages.debian.net/mandelbrot-8.c`: {}, // COMPILE FAIL: "\"mandelbrot-8.c:16:30: unsupported vector type: Vec (expr.go:1216:checkVolatileExpr: expr.go:101:expr: expr.go:547:expr0: expr.go:1706:unaryExpression: type.go:407:isValidType: type.go:440:isValidType..."
	`assets/benchmarksgame-team.pages.debian.net/mandelbrot.c`:   {}, // COMPILE FAIL: "\"mandelbrot.c:23:15: unsupported vector type: v2df (expr.go:541:expr0: expr.go:4174:primaryExpression: expr.go:4646:primaryExp

}
