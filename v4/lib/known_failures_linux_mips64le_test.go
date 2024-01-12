// Copyright 2022 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccgo // import "modernc.org/ccgo/v4/lib"

var testExecKnownFails = map[string]struct{}{
	// ==== EXEC FAIL - compiles and builds but fails when executed.

	// Won't fix: setjmp/longjmp
	`assets/github.com/vnmakarov/mir/c-benchmarks/except.c`: {},
	`assets/github.com/vnmakarov/mir/c-tests/new/setjmp.c`:  {},
	`assets/github.com/vnmakarov/mir/c-tests/new/setjmp2.c`: {},

	// Won't fix: sigfpe
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/execute/20101011-1.c`:                 {},
	`assets/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/execute/20101011-1.c`: {},

	// Won't fix: Architecture specific conversion overflow.
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/execute/20031003-1.c`:                 {},
	`assets/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/execute/20031003-1.c`: {},

	// Won't fix: implementation defined bit packing
	`assets/github.com/vnmakarov/mir/c-tests/lacc/bitfield-basic.c`:         {},
	`assets/github.com/vnmakarov/mir/c-tests/lacc/bitfield-trailing-zero.c`: {},
	`assets/github.com/vnmakarov/mir/c-tests/lacc/bitfield-types-init.c`:    {},

	// Won't fix: long double
	`assets/github.com/vnmakarov/mir/c-tests/lacc/long-double-load.c`: {},

	// Won't fix: return_addr
	`assets/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/execute/return-addr.c`: {},

	// ==== BUILD FAIL - compiles but does not build.

	`assets/benchmarksgame-team.pages.debian.net/reverse-complement-4.c`: {}, // BUILD FAIL: "exit status 1"
	`assets/ccgo/bug/sqlite.c`: {}, // BUILD FAIL: "exit status 1"

	// ==== COMPILE FAIL - does not compile.

	`assets/benchmarksgame-team.pages.debian.net/fasta-4.c`:             {}, // COMPILE FAIL: "fasta-4.o.go:638:3: undefined: \"fwrite_unlocked\" external (all_test.go:449:1: all_test.go:536:testExec1: ccgo.go:193:Main: ccgo.go:587:main: link.go:302:link: link.go:826:link:)"
	`assets/benchmarksgame-team.pages.debian.net/mandelbrot-8.c`:        {}, // COMPILE FAIL: "\"mandelbrot-8.c:16:30: unsupported vector type: Vec (expr.go:1205:checkVolatileExpr: expr.go:101:expr: expr.go:537:expr0: expr.go:1689:unaryExpression: type.go:412:isValidType: type.go:445:isValidType..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20000211-1.c`: {}, // COMPILE FAIL: "\"TODO (compile.go:433:compile: decl.go:297:externalDeclaration: decl.go:607:declaration: decl.go:746:initDeclarator: type.go:18:typedef: type.go:320:typ0:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20000326-2.c`: {}, // COMPILE FAIL: "\"20000326-2.c:7:3: label declarations not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:353:functionDefinition0: stmt.go:324:compoundState..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20000405-3.c`: {}, // COMPILE FAIL: "\"20000405-3.c:1:1: unsupported alignment 32 of struct foo {entry array of 40 pointer to void} (decl.go:297:externalDeclaration: decl.go:597:declaration: type.go:564:defineStructType: type.go:499:struc..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20000518-1.c`: {}, // COMPILE FAIL: "\"20000518-1.c:7:2: label declarations not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:353:functionDefinition0: stmt.go:324:compoundState..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20010202-1.c`: {}, // COMPILE FAIL: "\"20010202-1.c:3:5: incomplete type: array of array of char (decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:343:functionDefinition0: type.go:464:isValidType1: type.go:388:isVa..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20010226-1.c`: {}, // COMPILE FAIL: "\"20010226-1.c:16:12: nested functions not supported (stmt.go:360:blockItem: stmt.go:42:statement: stmt.go:392:selectionStatement: stmt.go:536:bracedStatement: stmt.go:545:unbracedStatement: stmt.go:36..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20010605-1.c`: {}, // COMPILE FAIL: "\"20010605-1.c:9:9: nested functions not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:384:functionDefinition0: stmt.go:324:compoundStateme..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20010903-2.c`: {}, // COMPILE FAIL: "\"20010903-2.c:9:14: nested functions not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:384:functionDefinition0: stmt.go:324:compoundStatem..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20011023-1.c`: {}, // COMPILE FAIL: "\"20011023-1.c:8:8: nested functions not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:384:functionDefinition0: stmt.go:324:compoundStateme..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20020309-1.c`: {}, // COMPILE FAIL: "\"20020309-1.c:8:5: nested functions not supported (decl.go:384:functionDefinition0: stmt.go:324:compoundStatement: stmt.go:360:blockItem: stmt.go:26:statement: stmt.go:545:unbracedStatement: stmt.go:3..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20021108-1.c`: {}, // COMPILE FAIL: "\"TODO UnaryExpressionLabelAddr (expr.go:1185:additiveExpression: expr.go:1200:binopArgs: expr.go:1205:checkVolatileExpr: expr.go:101:expr: expr.go:537:expr0: expr.go:1698:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20021204-1.c`: {}, // COMPILE FAIL: "\"20021204-1.c:8:7: nested functions not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:384:functionDefinition0: stmt.go:324:compoundStateme..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20030224-1.c`: {}, // COMPILE FAIL: "\"20030224-1.c:6:25: invalid type size: -1 (decl.go:384:functionDefinition0: stmt.go:261:compoundStatement: decl.go:236:declareLocals: type.go:42:typ: type.go:65:typ0: type.go:483:isValidType1:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20030418-1.c`: {}, // COMPILE FAIL: "\"20030418-1.c:13:8: nested functions not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:384:functionDefinition0: stmt.go:324:compoundStatem..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20030716-1.c`: {}, // COMPILE FAIL: "\"20030716-1.c:3:6: incomplete type: array of int (decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:343:functionDefinition0: type.go:464:isValidType1: type.go:388:isValidParamTy..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20030903-1.c`: {}, // COMPILE FAIL: "\"TODO UnaryExpressionReal (expr.go:101:expr: expr.go:501:expr0: expr.go:3510:assignmentExpression: expr.go:101:expr: expr.go:537:expr0: expr.go:1724:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20030910-1.c`: {}, // COMPILE FAIL: "\"TODO UnaryExpressionReal (expr.go:101:expr: expr.go:501:expr0: expr.go:3515:assignmentExpression: expr.go:101:expr: expr.go:537:expr0: expr.go:1724:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20031011-1.c`: {}, // COMPILE FAIL: "\"20031011-1.c:15:8: nested functions not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:384:functionDefinition0: stmt.go:324:compoundStatem..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20040310-1.c`: {}, // COMPILE FAIL: "\"20040310-1.c:4:7: nested functions not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:384:functionDefinition0: stmt.go:324:compoundStateme..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20040317-1.c`: {}, // COMPILE FAIL: "\"20040317-1.c:1:5: incomplete type: array of array of char (decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:343:functionDefinition0: type.go:464:isValidType1: type.go:388:isVa..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20040317-3.c`: {}, // COMPILE FAIL: "\"20040317-3.c:4:7: nested functions not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:384:functionDefinition0: stmt.go:324:compoundStateme..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20040323-1.c`: {}, // COMPILE FAIL: "\"20040323-1.c:10:16: nested functions not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:384:functionDefinition0: stmt.go:324:compoundState..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20040614-1.c`: {}, // COMPILE FAIL: "\"TODO <nil> (asm_mips64x.s:648:goexit: compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:345:functionDefinition0: decl.go:102:newFnCtx:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20050113-1.c`: {}, // COMPILE FAIL: "\"20050113-1.c:11:20: unsupported vector type: V2SF (expr.go:531:expr0: expr.go:3898:primaryExpression: expr.go:4361:primaryExpressionFloatConst: type.go:60:verifyTyp: type.go:65:typ0: type.go:445:isVa..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20050119-1.c`: {}, // COMPILE FAIL: "\"20050119-1.c:7:8: nested functions not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:384:functionDefinition0: stmt.go:324:compoundStateme..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20050122-2.c`: {}, // COMPILE FAIL: "\"20050122-2.c:10:3: label declarations not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:353:functionDefinition0: stmt.go:324:compoundStat..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20050510-1.c`: {}, // COMPILE FAIL: "\"TODO <nil> (asm_mips64x.s:648:goexit: compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:345:functionDefinition0: decl.go:102:newFnCtx:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20070603-1.c`: {}, // COMPILE FAIL: "\"TODO UnaryExpressionReal (expr.go:1168:additiveExpression: expr.go:1200:binopArgs: expr.go:1205:checkVolatileExpr: expr.go:101:expr: expr.go:537:expr0: expr.go:1724:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20070603-2.c`: {}, // COMPILE FAIL: "\"TODO UnaryExpressionReal (expr.go:1168:additiveExpression: expr.go:1200:binopArgs: expr.go:1205:checkVolatileExpr: expr.go:101:expr: expr.go:537:expr0: expr.go:1724:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20070919-1.c`: {}, // COMPILE FAIL: "\"TODO exprUintptr (expr.go:101:expr: expr.go:537:expr0: expr.go:1612:unaryExpression: expr.go:101:expr: expr.go:529:expr0: expr.go:2126:postfixExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20090907-1.c`: {}, // COMPILE FAIL: "\"TODO (decl.go:297:externalDeclaration: decl.go:597:declaration: type.go:564:defineStructType: type.go:499:structLiteral: type.go:263:typ0: type.go:320:typ0:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20110131-1.c`: {}, // COMPILE FAIL: "\"-: TODO (expr.go:1205:checkVolatileExpr: expr.go:101:expr: expr.go:513:expr0: expr.go:921:conditionalExpression: expr.go:70:topExpr: expr.go:85:expr:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/20121107-1.c`: {}, // COMPILE FAIL: "\"-: TODO (expr.go:531:expr0: expr.go:3915:primaryExpression: expr.go:513:expr0: expr.go:921:conditionalExpression: expr.go:70:topExpr: expr.go:85:expr:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/920301-1.c`:   {}, // COMPILE FAIL: "\"TODO UnaryExpressionLabelAddr (init.go:263:initializerArray: init.go:93:initializer: expr.go:70:topExpr: expr.go:101:expr: expr.go:537:expr0: expr.go:1698:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/920415-1.c`:   {}, // COMPILE FAIL: "\"TODO <nil> (asm_mips64x.s:648:goexit: compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:345:functionDefinition0: decl.go:102:newFnCtx:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/920428-3.c`:   {}, // COMPILE FAIL: "\"TODO <nil> (asm_mips64x.s:648:goexit: compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:345:functionDefinition0: decl.go:102:newFnCtx:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/920501-1.c`:   {}, // COMPILE FAIL: "\"TODO UnaryExpressionLabelAddr (init.go:263:initializerArray: init.go:93:initializer: expr.go:70:topExpr: expr.go:101:expr: expr.go:537:expr0: expr.go:1698:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/920501-7.c`:   {}, // COMPILE FAIL: "\"TODO <nil> (asm_mips64x.s:648:goexit: compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:345:functionDefinition0: decl.go:102:newFnCtx:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/920502-1.c`:   {}, // COMPILE FAIL: "\"TODO <nil> (asm_mips64x.s:648:goexit: compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:345:functionDefinition0: decl.go:102:newFnCtx:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/920826-1.c`:   {}, // COMPILE FAIL: "\"TODO <nil> (asm_mips64x.s:648:goexit: compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:345:functionDefinition0: decl.go:102:newFnCtx:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/920831-1.c`:   {}, // COMPILE FAIL: "\"TODO <nil> (asm_mips64x.s:648:goexit: compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:345:functionDefinition0: decl.go:102:newFnCtx:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/930118-1.c`:   {}, // COMPILE FAIL: "\"930118-1.c:3:1: label declarations not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:353:functionDefinition0: stmt.go:324:compoundStateme..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/930506-2.c`:   {}, // COMPILE FAIL: "\"930506-2.c:5:9: nested functions not supported (decl.go:384:functionDefinition0: stmt.go:324:compoundStatement: stmt.go:360:blockItem: stmt.go:26:statement: stmt.go:545:unbracedStatement: stmt.go:363..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/941014-4.c`:   {}, // COMPILE FAIL: "\"TODO UnaryExpressionLabelAddr (init.go:32:initializerOuter: init.go:93:initializer: expr.go:70:topExpr: expr.go:101:expr: expr.go:537:expr0: expr.go:1698:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/950610-1.c`:   {}, // COMPILE FAIL: "\"950610-1.c:1:1: incomplete type: array of array of int (decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:343:functionDefinition0: type.go:464:isValidType1: type.go:388:isValid..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/950613-1.c`:   {}, // COMPILE FAIL: "\"TODO <nil> (asm_mips64x.s:648:goexit: compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:345:functionDefinition0: decl.go:102:newFnCtx:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/950919-1.c`:   {}, // COMPILE FAIL: "\"950919-1.c:2:10: assertions are a deprecated extension\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/951116-1.c`:   {}, // COMPILE FAIL: "\"951116-1.c:7:7: nested functions not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:384:functionDefinition0: stmt.go:324:compoundStatement..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/981001-4.c`:   {}, // COMPILE FAIL: "\"-: TODO (expr.go:531:expr0: expr.go:3915:primaryExpression: expr.go:513:expr0: expr.go:921:conditionalExpression: expr.go:70:topExpr: expr.go:85:expr:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/981006-1.c`:   {}, // COMPILE FAIL: "\"981006-1.c:14:3: label declarations not supported (compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:353:functionDefinition0: stmt.go:324:compoundStatem..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/981223-1.c`:   {}, // COMPILE FAIL: "\"TODO UnaryExpressionReal (expr.go:1273:equalityExpression: expr.go:1200:binopArgs: expr.go:1205:checkVolatileExpr: expr.go:101:expr: expr.go:537:expr0: expr.go:1724:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/991213-1.c`:   {}, // COMPILE FAIL: "\"TODO UnaryExpressionImag (expr.go:70:topExpr: expr.go:101:expr: expr.go:531:expr0: expr.go:3915:primaryExpression: expr.go:537:expr0: expr.go:1722:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/991213-3.c`:   {}, // COMPILE FAIL: "\"TODO <nil> (asm_mips64x.s:648:goexit: compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:345:functionDefinition0: decl.go:102:newFnCtx:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/complex-1.c`:  {}, // COMPILE FAIL: "\"TODO *cc.PredefinedType _Complex int _Complex int (expr.go:70:topExpr: expr.go:115:expr: expr.go:169:convert: expr.go:331:convertType: type.go:30:helper: type.go:159:typ0:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/complex-2.c`:  {}, // COMPILE FAIL: "\"TODO UnaryExpressionImag (expr.go:101:expr: expr.go:501:expr0: expr.go:3510:assignmentExpression: expr.go:101:expr: expr.go:537:expr0: expr.go:1722:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/complex-3.c`:  {}, // COMPILE FAIL: "\"TODO UnaryExpressionImag (expr.go:1168:additiveExpression: expr.go:1200:binopArgs: expr.go:1205:checkVolatileExpr: expr.go:101:expr: expr.go:537:expr0: expr.go:1722:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/complex-4.c`:  {}, // COMPILE FAIL: "\"TODO UnaryExpressionReal (expr.go:101:expr: expr.go:501:expr0: expr.go:3510:assignmentExpression: expr.go:101:expr: expr.go:537:expr0: expr.go:1724:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/complex-5.c`:  {}, // COMPILE FAIL: "\"TODO *cc.PredefinedType _Complex int _Complex int (expr.go:101:expr: expr.go:531:expr0: expr.go:3896:primaryExpression: expr.go:4340:primaryExpressionIntConst: type.go:30:helper: type.go:159:typ0:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/complex-6.c`:  {}, // COMPILE FAIL: "\"TODO *cc.PredefinedType _Complex int _Complex int (expr.go:101:expr: expr.go:531:expr0: expr.go:3896:primaryExpression: expr.go:4340:primaryExpressionIntConst: type.go:30:helper: type.go:159:typ0:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/ex.c`:         {}, // COMPILE FAIL: "\"ex.c:12:19: too few arguments to function 'foo', type 'function(int, int) returning int' in 'foo ()' (expr.go:3227:postfixExpressionCall: expr.go:70:topExpr: expr.go:101:expr: expr.go:529:expr0: expr..."
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/icfmatch.c`:   {}, // COMPILE FAIL: "\"icfmatch.c:4:14: unsupported vector type: v4qi (expr.go:101:expr: expr.go:531:expr0: expr.go:3804:primaryExpression: type.go:60:verifyTyp: type.go:65:typ0: type.go:445:isValidType1:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/labels-1.c`:   {}, // COMPILE FAIL: "\"TODO UnaryExpressionLabelAddr (init.go:32:initializerOuter: init.go:93:initializer: expr.go:70:topExpr: expr.go:101:expr: expr.go:537:expr0: expr.go:1698:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/labels-2.c`:   {}, // COMPILE FAIL: "\"TODO UnaryExpressionLabelAddr (init.go:398:initializerStruct: init.go:93:initializer: expr.go:70:topExpr: expr.go:101:expr: expr.go:537:expr0: expr.go:1698:unaryExpression:)\""
	`assets/gcc-9.1.0/gcc/testsuite/gcc.c-torture/compile/labels-3.c`:   {}, // COMPILE FAIL: "\"TODO <nil> (asm_mips64x.s:648:goexit: compile.go:433:compile: decl.go:295:externalDeclaration: decl.go:323:functionDefinition: decl.go:345:functionDefinition0: decl.go:102:newFnCtx:)\""
}
