// Copyright 2022 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccgo // import "modernc.org/ccgo/v4/lib"

import (
	"fmt"
	"sort"
	"strings"

	"modernc.org/cc/v4"
)

func (c *ctx) statement(w writer, n *cc.Statement) {
	sep := sep(n)
	if c.task.positions {
		sep = strings.TrimRight(sep, "\n\r\t ")
	}
	switch n.Case {
	case cc.StatementLabeled: // LabeledStatement
		w.w("%s%s", sep, c.posComment(n))
		c.labeledStatement(w, n.LabeledStatement)
	case cc.StatementCompound: // CompoundStatement
		//TODO- c.compoundStatement(w, n.CompoundStatement, false, "")
		c.unbracedStatement(w, n)
	case cc.StatementExpr: // ExpressionStatement
		w.w("%s%s", sep, c.posComment(n))
		e := n.ExpressionStatement.ExpressionList
		if e == nil {
			break
		}

		w.w("%s;", c.discardStr2(e, c.topExpr(w, e, nil, exprVoid)))
	case cc.StatementSelection: // SelectionStatement
		w.w("%s%s", sep, c.posComment(n))
		c.selectionStatement(w, n.SelectionStatement)
	case cc.StatementIteration: // IterationStatement
		w.w("%s%s", sep, c.posComment(n))
		c.iterationStatement(w, n.IterationStatement)
	case cc.StatementJump: // JumpStatement
		w.w("%s%s", sep, c.posComment(n))
		c.jumpStatement(w, n.JumpStatement)
	case cc.StatementAsm: // AsmStatement
		w.w("%s%s", sep, c.posComment(n))
		a := strings.Split(nodeSource(n.AsmStatement), "\n")
		w.w("\n// %s", strings.Join(a, "\n// "))
		p := c.pos(n)
		w.w("\n%s__assert_fail(%stls, %q, %q, %d, %q);", tag(external), tag(ccgo), "assembler statements not supported\x00", p.Filename+"\x00", p.Line, c.fn.Name()+"\x00")
		if !c.task.ignoreAsmErrors {
			c.err(errorf("%v: assembler statements not supported", c.pos(n)))
		}
	default:
		c.err(errorf("internal error %T %v", n, n.Case))
	}
}

func (c *ctx) mustConsume(n cc.ExpressionNode) (r bool) {
	// defer func() { trc("%v: %T %v %s: %v", n.Position(), n, n.Type().Kind(), cc.NodeSource(n), r) }()
	switch x := n.(type) {
	case *cc.AdditiveExpression:
		switch x.Case {
		case cc.AdditiveExpressionMul:
			return c.mustConsume(x.MultiplicativeExpression)
		default:
			return true
		}
	case *cc.AndExpression:
		switch x.Case {
		case cc.AndExpressionEq:
			return c.mustConsume(x.EqualityExpression)
		default:
			return true
		}
	case *cc.AssignmentExpression:
		switch x.Case {
		case cc.AssignmentExpressionCond:
			return c.mustConsume(x.ConditionalExpression)
		default:
			return false
		}
	case *cc.CastExpression:
		switch x.Case {
		case cc.CastExpressionUnary:
			return c.mustConsume(x.UnaryExpression)
		default:
			return c.mustConsume(x.CastExpression)
		}
	case *cc.ConditionalExpression:
		switch x.Case {
		case cc.ConditionalExpressionCond:
			return x.Type().Kind() != cc.Void
		default:
			return true
		}
	case *cc.EqualityExpression:
		switch x.Case {
		case cc.EqualityExpressionRel:
			return c.mustConsume(x.RelationalExpression)
		default:
			return true
		}
	case *cc.ExclusiveOrExpression:
		switch x.Case {
		case cc.ExclusiveOrExpressionAnd:
			return c.mustConsume(x.AndExpression)
		default:
			return true
		}
	case *cc.ExpressionList:
		for ; ; x = x.ExpressionList {
			if x.ExpressionList == nil {
				return c.mustConsume(x.AssignmentExpression)
			}
		}
	case *cc.InclusiveOrExpression:
		switch x.Case {
		case cc.InclusiveOrExpressionXor:
			return c.mustConsume(x.ExclusiveOrExpression)
		default:
			return true
		}
	case *cc.LogicalAndExpression:
		switch x.Case {
		case cc.LogicalAndExpressionOr:
			return c.mustConsume(x.InclusiveOrExpression)
		default:
			return true
		}
	case *cc.LogicalOrExpression:
		switch x.Case {
		case cc.LogicalOrExpressionLAnd:
			return c.mustConsume(x.LogicalAndExpression)
		default:
			return true
		}
	case *cc.MultiplicativeExpression:
		switch x.Case {
		case cc.MultiplicativeExpressionCast:
			return c.mustConsume(x.CastExpression)
		default:
			return true
		}
	case *cc.PostfixExpression:
		switch x.Case {
		case cc.PostfixExpressionCall, cc.PostfixExpressionDec, cc.PostfixExpressionInc:
			return false
		default:
			return true
		}
	case *cc.PrimaryExpression:
		switch x.Case {
		case cc.PrimaryExpressionExpr:
			return c.mustConsume(x.ExpressionList)
		case cc.PrimaryExpressionStmt:
			return n.Type().Kind() != cc.Void
		default:
			return true
		}
	case *cc.RelationalExpression:
		switch x.Case {
		case cc.RelationalExpressionShift:
			return c.mustConsume(x.ShiftExpression)
		default:
			return true
		}
	case *cc.ShiftExpression:
		switch x.Case {
		case cc.ShiftExpressionAdd:
			return c.mustConsume(x.AdditiveExpression)
		default:
			return true
		}
	case *cc.UnaryExpression:
		switch x.Case {
		case cc.UnaryExpressionDec, cc.UnaryExpressionInc:
			return false
		default:
			return true
		}
	case nil:
		return false
	default:
		//trc("%v: %T %v, %s: %v", n.Position(), n, n.Type().Kind(), cc.NodeSource(n), r) //TODO-DBG
		panic(todo("%T", x))
	}
}

func (c *ctx) labeledStatement(w writer, n *cc.LabeledStatement) {
	switch n.Case {
	case cc.LabeledStatementLabel: // IDENTIFIER ':' Statement
		w.w("%s%s:", tag(preserve), n.Token.Src()) //TODO use nameSpace
		c.statement(w, n.Statement)
	case cc.LabeledStatementCaseLabel: // "case" ConstantExpression ':' Statement
		switch {
		case len(c.switchCtx) != 0:
			w.w("%s:", c.switchCtx[n])
		default:
			if n.CaseOrdinal() != 0 {
				w.w("fallthrough;")
			}
			w.w("case %s:", c.topExpr(nil, n.ConstantExpression, cc.IntegerPromotion(n.Switch().ExpressionList.Type()), exprDefault))
		}
		c.unbracedStatement(w, n.Statement)
	case cc.LabeledStatementRange: // "case" ConstantExpression "..." ConstantExpression ':' Statement
		if n.CaseOrdinal() != 0 {
			w.w("fallthrough;")
		}
		c.err(errorf("TODO %v", n.Case))
	case cc.LabeledStatementDefault: // "default" ':' Statement
		switch {
		case len(c.switchCtx) != 0:
			w.w("%s:", c.switchCtx[n])
		default:
			if n.CaseOrdinal() != 0 {
				w.w("fallthrough;")
			}
			w.w("default:")
		}
		c.unbracedStatement(w, n.Statement)
	default:
		c.err(errorf("internal error %T %v", n, n.Case))
	}
}

func (c *ctx) compoundStatement(w writer, n *cc.CompoundStatement, fnBlock bool, value string) {
	defer func() { c.compoundStmtValue = value }()

	_, flat := c.f.flatScopes[n.LexicalScope()]
	switch {
	case fnBlock:
		if c.pass != 2 {
			break
		}

		w.w("{\n")
		switch {
		case c.f.tlsAllocs+int64(c.f.maxVaListSize) != 0:
			c.f.tlsAllocs = roundup(c.f.tlsAllocs, 8)
			v := c.f.tlsAllocs
			if c.f.maxVaListSize != 0 {
				v += c.f.maxVaListSize + 8
			}
			v = roundup(v, 16)
			w.w("%sbp := %[1]stls.%sAlloc(%d); /* tlsAllocs %v maxVaListSize %v */", tag(ccgo), tag(preserve), v, c.f.tlsAllocs, c.f.maxVaListSize)
			w.w("defer %stls.%sFree(%d);", tag(ccgo), tag(preserve), v)
			for _, v := range c.f.t.Parameters() {
				if d := v.Declarator; d != nil && c.f.declInfos.info(d).pinned() {
					w.w("*(*%s)(%s) = %s_%s;", c.typ(n, d.Type()), unsafePointer(bpOff(c.f.declInfos.info(d).bpOff)), tag(ccgo), d.Name())
				}
			}
			w.w("%s%s", sep(n.Token), c.posComment(n))
		default:
			w.w("%s%s", strings.TrimSpace(sep(n.Token)), c.posComment(n))
		}
		w.w("%s", c.f.declareLocals())
		if c.f.vlaSizes != nil {
			var a []string
			for d := range c.f.vlaSizes {
				a = append(a, c.f.locals[d])
			}
			sort.Strings(a)
			w.w("defer func() {")
			for _, v := range a {
				w.w("%srealloc(%stls, %s, 0);", tag(external), tag(ccgo), v)
			}
			w.w("}();")
		}
		if c.f.callsAlloca {
			w.w("defer %stls.FreeAlloca();", tag(ccgo))
		}
	default:
		if !flat {
			w.w(" {\n %s%s", sep(n.Token), c.posComment(n))
		}
	}
	var bi *cc.BlockItem
	for l := n.BlockItemList; l != nil; l = l.BlockItemList {
		bi = l.BlockItem
		if l.BlockItemList == nil && value != "" {
			switch bi.Case {
			case cc.BlockItemStmt:
				s := bi.Statement
				for s.Case == cc.StatementLabeled {
					s = s.LabeledStatement.Statement
				}
				switch s.Case {
				case cc.StatementExpr:
					switch e := s.ExpressionStatement.ExpressionList; x := e.(type) {
					case *cc.PrimaryExpression:
						switch x.Case {
						case cc.PrimaryExpressionStmt:
							c.blockItem(w, bi)
							w.w("%s = %s;", value, c.compoundStmtValue)
						default:
							w.w("%s = ", value)
							c.blockItem(w, bi)
						}
					default:
						w.w("%s = ", value)
						c.blockItem(w, bi)
					}
				default:
					// trc("%v: %s", bi.Position(), cc.NodeSource(bi))
					c.err(errorf("%v: TODO %v", s.Position(), s.Case))
				}
			default:
				c.err(errorf("TODO %v", bi.Case))
			}
			break
		}

		c.blockItem(w, bi)
	}
	switch {
	case fnBlock && c.f.t.Result().Kind() != cc.Void && !c.isReturn(bi):
		s := sep(n.Token2)
		if strings.Contains(s, "\n") {
			w.w("%s", s)
			s = ""
		}
		w.w("return %s%s;%s", tag(ccgo), retvalName, s)
		if !flat {
			w.w("};")
		}
	default:
		if !flat {
			w.w("%s", sep(n.Token2))
			w.w("};")
		}
	}
}

func (c *ctx) isReturn(n *cc.BlockItem) bool {
	if n == nil || n.Case != cc.BlockItemStmt {
		return false
	}

	return n.Statement.Case == cc.StatementJump && n.Statement.JumpStatement.Case == cc.JumpStatementReturn
}

func (c *ctx) blockItem(w writer, n *cc.BlockItem) {
	switch n.Case {
	case cc.BlockItemDecl: // Declaration
		c.declaration(w, n.Declaration, false)
	case cc.BlockItemLabel: // LabelDeclaration
		c.err(errorf("TODO %v", n.Case))
	case cc.BlockItemStmt: // Statement
		c.statement(w, n.Statement)
	case cc.BlockItemFuncDef: // DeclarationSpecifiers Declarator CompoundStatement
		if c.pass == 2 {
			c.err(errorf("%v: nested functions not supported", c.pos(n.Declarator)))
			// c.functionDefinition0(w, sep(n), n, n.Declarator, n.CompoundStatement, true) //TODO does not really work yet
		}
	default:
		c.err(errorf("internal error %T %v", n, n.Case))
	}
}

func (c *ctx) isEmptyStatment(n *cc.Statement) bool {
	return n == nil || n.Case == cc.StatementExpr && n.ExpressionStatement.ExpressionList == nil
}

func (c *ctx) selectionStatement(w writer, n *cc.SelectionStatement) {
	if _, ok := c.f.flatScopes[n.LexicalScope()]; ok {
		c.selectionStatementFlat(w, n)
		return
	}

	switch n.Statement.Case {
	case cc.StatementCompound:
		if _, ok := c.f.flatScopes[n.Statement.CompoundStatement.LexicalScope()]; ok {
			c.selectionStatementFlat(w, n)
			return
		}
	}

	switch n.Case {
	case cc.SelectionStatementIf: // "if" '(' ExpressionList ')' Statement
		w.w("if %s", c.expr(w, n.ExpressionList, nil, exprBool))
		c.bracedStatement(w, n.Statement)
	case cc.SelectionStatementIfElse: // "if" '(' ExpressionList ')' Statement "else" Statement
		switch {
		case c.isEmptyStatment(n.Statement):
			w.w("if !(%s) {", c.expr(w, n.ExpressionList, nil, exprBool))
			c.unbracedStatement(w, n.Statement2)
			w.w("};")
		default:
			w.w("if %s {", c.expr(w, n.ExpressionList, nil, exprBool))
			c.unbracedStatement(w, n.Statement)
			w.w("} else {")
			c.unbracedStatement(w, n.Statement2)
			w.w("};")
		}
	case cc.SelectionStatementSwitch: // "switch" '(' ExpressionList ')' Statement
		for _, v := range n.LabeledStatements() {
			if v.Case == cc.LabeledStatementLabel {
				c.selectionStatementFlat(w, n)
				return
			}

			if v.Statement == nil {
				continue
			}

			switch v.Statement.Case {
			case cc.StatementIteration:
				c.f.flatScopes[v.Statement.IterationStatement.LexicalScope()] = struct{}{}
				c.selectionStatementFlat(w, n)
				return
			}
		}

		ok := false
		switch s := n.Statement; s.Case {
		case cc.StatementCompound:
			if l := n.Statement.CompoundStatement.BlockItemList; l != nil {
				switch bi := l.BlockItem; bi.Case {
				case cc.BlockItemStmt:
					switch s := bi.Statement; s.Case {
					case cc.StatementLabeled:
						switch ls := s.LabeledStatement; ls.Case {
						case cc.LabeledStatementCaseLabel, cc.LabeledStatementDefault:
							// ok
							ok = true
						}
					}
				}
			}

		}
		if !ok {
			c.selectionStatementFlat(w, n)
			return
		}

		defer c.setSwitchCtx(nil)()
		defer c.setBreakCtx("")()

		w.w("switch %s", c.nonTopExpr(w, n.ExpressionList, cc.IntegerPromotion(n.ExpressionList.Type()), exprDefault))
		c.bracedStatement(w, n.Statement)
	default:
		c.err(errorf("internal error %T %v", n, n.Case))
	}
}

func (c *ctx) selectionStatementFlat(w writer, n *cc.SelectionStatement) {
	switch n.Case {
	case cc.SelectionStatementIf: // "if" '(' ExpressionList ')' Statement
		//	if !expr goto a
		//	stmt
		// a:
		a := c.label()
		w.w("if !(%s) { goto %s };", c.expr(w, n.ExpressionList, nil, exprBool), a)
		c.unbracedStatement(w, n.Statement)
		w.w("%s:", a)
	case cc.SelectionStatementIfElse: // "if" '(' ExpressionList ')' Statement "else" Statement
		//	if !expr goto a
		//	stmt
		//	goto b
		// a:	stmt2
		// b:
		a := c.label()
		b := c.label()
		w.w("if !(%s) { goto %s };", c.expr(w, n.ExpressionList, nil, exprBool), a)
		c.unbracedStatement(w, n.Statement)
		w.w("goto %s; %s:", b, a)
		c.unbracedStatement(w, n.Statement2)
		w.w("%s:", b)
	case cc.SelectionStatementSwitch: // "switch" '(' ExpressionList ')' Statement
		//	switch expr {
		//	case 1:
		//		goto label1
		//	case 2:
		//		goto label2
		//	...
		//	default:
		//		goto labelN
		//	}
		//	goto brk
		//	label1:
		//		statements in case 1
		//	label2:
		//		statements in case 2
		//	...
		//	labelN:
		//		statements in default
		//	brk:
		t := cc.IntegerPromotion(n.ExpressionList.Type())
		w.w("switch %s {", c.nonTopExpr(w, n.ExpressionList, t, exprDefault))
		labels := map[*cc.LabeledStatement]string{}
		for _, v := range n.LabeledStatements() {
			switch v.Case {
			case cc.LabeledStatementLabel: // IDENTIFIER ':' Statement
				// nop
			case cc.LabeledStatementCaseLabel: // "case" ConstantExpression ':' Statement
				label := c.label()
				labels[v] = label
				w.w("case %s: goto %s;", c.topExpr(nil, v.ConstantExpression, t, exprDefault), label)
			case cc.LabeledStatementRange: // "case" ConstantExpression "..." ConstantExpression ':' Statement
				c.err(errorf("TODO %v", n.Case))
			case cc.LabeledStatementDefault: // "default" ':' Statement
				label := c.label()
				labels[v] = label
				w.w("default: goto %s;", label)
			default:
				c.err(errorf("internal error %T %v", n, n.Case))
			}
		}
		brk := c.label()
		w.w("\n}; goto %s;", brk)

		defer c.setSwitchCtx(labels)()
		defer c.setBreakCtx(brk)()

		c.unbracedStatement(w, n.Statement)
		w.w("%s:", brk)
	default:
		c.err(errorf("internal error %T %v", n, n.Case))
	}
}

func (c *ctx) bracedStatement(w writer, n *cc.Statement) {
	w.w("{")
	c.unbracedStatement(w, n)
	w.w("};")
}

func (c *ctx) unbracedStatement(w writer, n *cc.Statement) {
	switch n.Case {
	case cc.StatementCompound:
		w.w("%s", sep(n))
		for l := n.CompoundStatement.BlockItemList; l != nil; l = l.BlockItemList {
			c.blockItem(w, l.BlockItem)
		}
		w.w("%s", sep(n.CompoundStatement.Token2))
	default:
		c.statement(w, n)
	}
}

func (c *ctx) iterationStatement(w writer, n *cc.IterationStatement) {
	defer c.setBreakCtx("")()
	defer c.setContinueCtx("")()

	if _, ok := c.f.flatScopes[n.LexicalScope()]; ok {
		c.iterationStatementFlat(w, n)
		return
	}

	switch n.Statement.Case {
	case cc.StatementCompound:
		if _, ok := c.f.flatScopes[n.Statement.CompoundStatement.LexicalScope()]; ok {
			c.iterationStatementFlat(w, n)
			return
		}
	}

	var cont string
	switch n.Case {
	case cc.IterationStatementWhile: // "while" '(' ExpressionList ')' Statement
		var a buf
		switch b := c.expr(&a, n.ExpressionList, nil, exprBool); {
		case a.len() != 0:
			w.w("for {")
			w.w("%s", a.bytes())
			w.w("\nif !(%s) { break };", b)
			c.unbracedStatement(w, n.Statement)
			w.w("\n};")
		default:
			w.w("for %s", b)
			c.bracedStatement(w, n.Statement)
		}
	case cc.IterationStatementDo: // "do" Statement "while" '(' ExpressionList ')' ';'
		if c.isSafeZero(n.ExpressionList) {
			c.unbracedStatement(w, n.Statement)
			break
		}

		var a buf
		switch b := c.expr(&a, n.ExpressionList, nil, exprBool); {
		case a.len() != 0:
			w.w("for %sfirst := true; ; %[1]sfirst = false {", tag(ccgo))
			w.w("\nif !first {")
			w.w("%s", a.bytes())
			w.w("\nif !(%s) { break };", b)
			w.w("\n};")
			c.unbracedStatement(w, n.Statement)
			w.w("};")
		default:
			w.w("for %scond := true; %[1]scond; %[1]scond = %s", tag(ccgo), b)
			c.bracedStatement(w, n.Statement)
		}
	case cc.IterationStatementFor: // "for" '(' ExpressionList ';' ExpressionList ';' ExpressionList ')' Statement
		var a, a2, a3 buf
		var b2 []byte
		var b4 string
		b := c.expr(&a, n.ExpressionList, nil, exprVoid)
		if b.len() == 0 && a.len() != 0 {
			s := string(a.bytes())
			s = strings.TrimSpace(s)
			s = strings.TrimRight(s, ";")
			if !strings.Contains(s, ";") {
				b.w("%s", s)
				a.reset()
			}
		}
		if n.ExpressionList2 != nil {
			b2 = c.expr(&a2, n.ExpressionList2, nil, exprBool).bytes()
		}
		b3 := c.expr(&a3, n.ExpressionList3, nil, exprVoid)
		if b3.len() != 0 {
			switch {
			case c.mustConsume(n.ExpressionList3):
				b4 = fmt.Sprintf("%s_ = %s", tag(preserve), b3)
			default:
				b4 = fmt.Sprintf("%s", b3)
			}
		}
		if a3.len() != 0 {
			cont = c.label()
			defer c.setContinueCtx(cont)()
		}
		switch {
		case a.len() == 0 && a2.len() == 0 && a3.len() == 0:
			w.w("for %s; %s; %s {", b, b2, b3)
			c.unbracedStatement(w, n.Statement)
			w.w("};")
		case a.len() == 0 && a2.len() == 0 && a3.len() != 0:
			w.w("for %s; %s;  {", b, b2)
			c.unbracedStatement(w, n.Statement)
			w.w("\ngoto %s; %[1]s:", cont)
			w.w("\n%s%s};", a3.bytes(), b4)
		case a.len() == 0 && a2.len() != 0 && a3.len() == 0:
			w.w("for %s; ; %s {", b, b3)
			w.w("\n%s", a2.bytes())
			w.w("\nif !(%s) { break };", b2)
			c.unbracedStatement(w, n.Statement)
			w.w("};")
		case a.len() == 0 && a2.len() != 0 && a3.len() != 0:
			w.w("for %s; ; {", b)
			w.w("\n%s", a2.bytes())
			w.w("\nif !(%s) { break };", b2)
			c.unbracedStatement(w, n.Statement)
			w.w("\ngoto %s; %[1]s:", cont)
			w.w("\n%s%s};", a3.bytes(), b4)
		case a.len() != 0 && a2.len() == 0 && a3.len() == 0:
			w.w("%s%s", a.bytes(), b.bytes())
			w.w("\nfor ;%s; %s{", b2, b3)
			c.unbracedStatement(w, n.Statement)
			w.w("};")
		case a.len() != 0 && a2.len() == 0 && a3.len() != 0:
			w.w("%s%s", a.bytes(), b.bytes())
			w.w("\nfor %s {", b2)
			c.unbracedStatement(w, n.Statement)
			w.w("\ngoto %s; %[1]s:", cont)
			w.w("\n%s%s};", a3.bytes(), b4)
		case a.len() != 0 && a2.len() != 0 && a3.len() == 0:
			w.w("%s%s", a.bytes(), b.bytes())
			w.w("\nfor ; ; %s {", b3)
			w.w("\n%s", a2.bytes())
			w.w("\nif !(%s) { break };", b2)
			c.unbracedStatement(w, n.Statement)
			w.w("};")
		case a.len() != 0 && a2.len() != 0 && a3.len() != 0:
			w.w("%s%s", a.bytes(), b.bytes())
			w.w("\nfor ; ; %s {", b3)
			w.w("\n%s", a2.bytes())
			w.w("\nif !(%s) { break };", b2)
			c.unbracedStatement(w, n.Statement)
			w.w("\ngoto %s; %[1]s:", cont)
			w.w("\n%s%s};", a3.bytes(), b4)
		}
	case cc.IterationStatementForDecl: // "for" '(' Declaration ExpressionList ';' ExpressionList ')' Statement
		c.declaration(w, n.Declaration, false)
		var a, a2 buf
		var b []byte
		if n.ExpressionList != nil {
			b = c.expr(&a, n.ExpressionList, nil, exprBool).bytes()
		}
		b2 := c.expr(&a2, n.ExpressionList2, nil, exprVoid)
		if a2.len() != 0 {
			cont = c.label()
			defer c.setContinueCtx(cont)()
		}
		switch {
		case a.len() == 0 && a2.len() == 0:
			w.w("for ; %s; %s {", b, b2)
			c.unbracedStatement(w, n.Statement)
			w.w("};")
		case a.len() == 0 && a2.len() != 0:
			w.w("for %s  {", b)
			c.unbracedStatement(w, n.Statement)
			w.w("\ngoto %s; %[1]s:", cont)
			w.w("\n%s%s};", a2.bytes(), b2.bytes())
		case a.len() != 0 && a2.len() == 0:
			w.w("for ; ; %s {", b2)
			w.w("\n%s", a.bytes())
			w.w("\nif !(%s) { break };", b)
			c.unbracedStatement(w, n.Statement)
			w.w("};")
		default: // case a.len() != 0 && a2.len() != 0:
			w.w("for {")
			w.w("\n%s", a.bytes())
			w.w("\nif !(%s) { break };", b)
			c.unbracedStatement(w, n.Statement)
			w.w("\ngoto %s; %[1]s:", cont)
			w.w("\n%s%s};", a2.bytes(), b2.bytes())
		}
	default:
		c.err(errorf("internal error %T %v", n, n.Case))
	}
}

func (c *ctx) iterationStatementFlat(w writer, n *cc.IterationStatement) {
	brk := c.label()
	cont := c.label()

	defer c.setBreakCtx(brk)()
	defer c.setContinueCtx(cont)()

	switch n.Case {
	case cc.IterationStatementWhile: // "while" '(' ExpressionList ')' Statement
		// cont:
		//	if !expr goto brk
		//	stmt
		//	goto cont
		// brk:
		w.w("%s:", cont)
		w.w("if !(%s) { goto %s };", c.expr(w, n.ExpressionList, nil, exprBool), brk)
		c.unbracedStatement(w, n.Statement)
		w.w("goto %s; %s:", cont, brk)
	case cc.IterationStatementDo: // "do" Statement "while" '(' ExpressionList ')' ';'
		// cont:
		//	stmt
		//	if expr goto cont
		// brk:
		w.w("%s:", cont)
		c.unbracedStatement(w, n.Statement)
		w.w("if (%s) { goto %s }; goto %s; %[3]s:", c.expr(w, n.ExpressionList, nil, exprBool), cont, brk)
	case cc.IterationStatementFor: // "for" '(' ExpressionList ';' ExpressionList ';' ExpressionList ')' Statement
		//	expr1
		// a:	if !expr2 goto brk
		//	stmt
		// cont:
		//	expr3
		//	goto a
		// brk:
		a := c.label()
		b := c.expr(w, n.ExpressionList, nil, exprVoid)
		if b.len() != 0 {
			w.w("%s;", b)
		}
		w.w("%s: ", a)
		if n.ExpressionList2 != nil {
			w.w("if !(%s) { goto %s };", c.expr(w, n.ExpressionList2, nil, exprBool), brk)
		}
		c.unbracedStatement(w, n.Statement)
		w.w("goto %s; %[1]s: ", cont)
		w.w("%s;", c.expr(w, n.ExpressionList3, nil, exprVoid))
		w.w("goto %s; goto %s; %[2]s:", a, brk)
	case cc.IterationStatementForDecl: // "for" '(' Declaration ExpressionList ';' ExpressionList ')' Statement
		//	decl
		// a:	if !expr goto brk
		//	stmt
		// cont:
		//	expr2
		//	goto a
		// brk:
		a := c.label()
		c.declaration(w, n.Declaration, false)
		w.w("%s: if !(%s) { goto %s };", a, c.expr(w, n.ExpressionList, nil, exprBool), brk)
		c.unbracedStatement(w, n.Statement)
		w.w("goto %s; %[1]s: ", cont)
		w.w("%s;", c.expr(w, n.ExpressionList2, nil, exprVoid))
		w.w("goto %s; %s:", a, brk)
	default:
		c.err(errorf("internal error %T %v", n, n.Case))
	}
}

func (c *ctx) label() string { return fmt.Sprintf("%s_%d", tag(ccgo), c.id()) }

func (c *ctx) jumpStatement(w writer, n *cc.JumpStatement) {
	switch n.Case {
	case cc.JumpStatementGoto: // "goto" IDENTIFIER ';'
		w.w("goto %s%s;", tag(preserve), n.Token2.Src())
	case cc.JumpStatementGotoExpr: // "goto" '*' ExpressionList ';'
		c.err(errorf("TODO %v", n.Case))
	case cc.JumpStatementContinue: // "continue" ';'
		if c.continueCtx != "" {
			w.w("goto %s;", c.continueCtx)
			break
		}

		w.w("continue;")
	case cc.JumpStatementBreak: // "break" ';'
		if c.breakCtx != "" {
			w.w("goto %s;", c.breakCtx)
			break
		}

		w.w("break;")
	case cc.JumpStatementReturn: // "return" ExpressionList ';'
		if nfo := c.f.inlineInfo; nfo != nil {
			switch ft := nfo.fd.Declarator.Type().(*cc.FunctionType); {
			case n.ExpressionList != nil:
				switch {
				case nfo.mode == exprVoid:
					if nfo.exit == "" {
						nfo.exit = c.label()
					}
					w.w("%s_ = %s;", tag(preserve), c.topExpr(w, n.ExpressionList, nil, exprDefault))
				default:
					if nfo.exit == "" {
						nfo.result = c.f.newAutovar(nfo.fd, ft.Result())
						nfo.exit = c.label()
					}
					w.w("%s = %s;", nfo.result, c.topExpr(w, n.ExpressionList, ft.Result(), exprDefault))
				}
			}
			if nfo.exit == "" {
				nfo.exit = c.label()
			}
			w.w("goto %s;", nfo.exit)
			return
		}

		switch {
		case n.ExpressionList != nil:
			switch {
			case c.f.t.Result().Kind() == cc.Void:
				w.w("%s; return;", c.expr(w, n.ExpressionList, nil, exprVoid))
			default:
				w.w("return %s;", c.checkVolatileExpr(w, n.ExpressionList, c.f.t.Result(), exprDefault))
			}
		default:
			w.w("return;")
		}
	default:
		c.err(errorf("internal error %T %v", n, n.Case))
	}
}
