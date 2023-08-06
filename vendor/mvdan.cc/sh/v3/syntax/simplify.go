// Copyright (c) 2017, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package syntax

import "bytes"

// Simplify modifies a node to remove redundant pieces of syntax, and returns
// whether any changes were made.
//
// The changes currently applied are:
//
//	Remove clearly useless parentheses       $(( (expr) ))
//	Remove dollars from vars in exprs        (($var))
//	Remove duplicate subshells               $( (stmts) )
//	Remove redundant quotes                  [[ "$var" == str ]]
//	Merge negations with unary operators     [[ ! -n $var ]]
//	Use single quotes to shorten literals    "\$foo"
func Simplify(n Node) bool {
	s := simplifier{}
	Walk(n, s.visit)
	return s.modified
}

type simplifier struct {
	modified bool
}

func (s *simplifier) visit(node Node) bool {
	switch x := node.(type) {
	case *Assign:
		x.Index = s.removeParensArithm(x.Index)
		// Don't inline params, as x[i] and x[$i] mean
		// different things when x is an associative
		// array; the first means "i", the second "$i".
	case *ParamExp:
		x.Index = s.removeParensArithm(x.Index)
		// don't inline params - same as above.

		if x.Slice == nil {
			break
		}
		x.Slice.Offset = s.removeParensArithm(x.Slice.Offset)
		x.Slice.Offset = s.inlineSimpleParams(x.Slice.Offset)
		x.Slice.Length = s.removeParensArithm(x.Slice.Length)
		x.Slice.Length = s.inlineSimpleParams(x.Slice.Length)
	case *ArithmExp:
		x.X = s.removeParensArithm(x.X)
		x.X = s.inlineSimpleParams(x.X)
	case *ArithmCmd:
		x.X = s.removeParensArithm(x.X)
		x.X = s.inlineSimpleParams(x.X)
	case *ParenArithm:
		x.X = s.removeParensArithm(x.X)
		x.X = s.inlineSimpleParams(x.X)
	case *BinaryArithm:
		x.X = s.inlineSimpleParams(x.X)
		x.Y = s.inlineSimpleParams(x.Y)
	case *CmdSubst:
		x.Stmts = s.inlineSubshell(x.Stmts)
	case *Subshell:
		x.Stmts = s.inlineSubshell(x.Stmts)
	case *Word:
		x.Parts = s.simplifyWord(x.Parts)
	case *TestClause:
		x.X = s.removeParensTest(x.X)
		x.X = s.removeNegateTest(x.X)
	case *ParenTest:
		x.X = s.removeParensTest(x.X)
		x.X = s.removeNegateTest(x.X)
	case *BinaryTest:
		x.X = s.unquoteParams(x.X)
		x.X = s.removeNegateTest(x.X)
		if x.Op == TsMatchShort {
			s.modified = true
			x.Op = TsMatch
		}
		switch x.Op {
		case TsMatch, TsNoMatch:
			// unquoting enables globbing
		default:
			x.Y = s.unquoteParams(x.Y)
		}
		x.Y = s.removeNegateTest(x.Y)
	case *UnaryTest:
		x.X = s.unquoteParams(x.X)
	}
	return true
}

func (s *simplifier) simplifyWord(wps []WordPart) []WordPart {
parts:
	for i, wp := range wps {
		dq, _ := wp.(*DblQuoted)
		if dq == nil || len(dq.Parts) != 1 {
			break
		}
		lit, _ := dq.Parts[0].(*Lit)
		if lit == nil {
			break
		}
		var buf bytes.Buffer
		escaped := false
		for _, r := range lit.Value {
			switch r {
			case '\\':
				escaped = !escaped
				if escaped {
					continue
				}
			case '\'':
				continue parts
			case '$', '"', '`':
				escaped = false
			default:
				if escaped {
					continue parts
				}
				escaped = false
			}
			buf.WriteRune(r)
		}
		newVal := buf.String()
		if newVal == lit.Value {
			break
		}
		s.modified = true
		wps[i] = &SglQuoted{
			Left:   dq.Pos(),
			Right:  dq.End(),
			Dollar: dq.Dollar,
			Value:  newVal,
		}
	}
	return wps
}

func (s *simplifier) removeParensArithm(x ArithmExpr) ArithmExpr {
	for {
		par, _ := x.(*ParenArithm)
		if par == nil {
			return x
		}
		s.modified = true
		x = par.X
	}
}

func (s *simplifier) inlineSimpleParams(x ArithmExpr) ArithmExpr {
	w, _ := x.(*Word)
	if w == nil || len(w.Parts) != 1 {
		return x
	}
	pe, _ := w.Parts[0].(*ParamExp)
	if pe == nil || !ValidName(pe.Param.Value) {
		// Not a parameter expansion, or not a valid name, like $3.
		return x
	}
	if pe.Excl || pe.Length || pe.Width || pe.Slice != nil ||
		pe.Repl != nil || pe.Exp != nil || pe.Index != nil {
		// A complex parameter expansion can't be simplified.
		//
		// Note that index expressions can't generally be simplified
		// either. It's fine to turn ${a[0]} into a[0], but others like
		// a[*] are invalid in many shells including Bash.
		return x
	}
	s.modified = true
	return &Word{Parts: []WordPart{pe.Param}}
}

func (s *simplifier) inlineSubshell(stmts []*Stmt) []*Stmt {
	for len(stmts) == 1 {
		st := stmts[0]
		if st.Negated || st.Background || st.Coprocess ||
			len(st.Redirs) > 0 {
			break
		}
		sub, _ := st.Cmd.(*Subshell)
		if sub == nil {
			break
		}
		s.modified = true
		stmts = sub.Stmts
	}
	return stmts
}

func (s *simplifier) unquoteParams(x TestExpr) TestExpr {
	w, _ := x.(*Word)
	if w == nil || len(w.Parts) != 1 {
		return x
	}
	dq, _ := w.Parts[0].(*DblQuoted)
	if dq == nil || len(dq.Parts) != 1 {
		return x
	}
	if _, ok := dq.Parts[0].(*ParamExp); !ok {
		return x
	}
	s.modified = true
	w.Parts = dq.Parts
	return w
}

func (s *simplifier) removeParensTest(x TestExpr) TestExpr {
	for {
		par, _ := x.(*ParenTest)
		if par == nil {
			return x
		}
		s.modified = true
		x = par.X
	}
}

func (s *simplifier) removeNegateTest(x TestExpr) TestExpr {
	u, _ := x.(*UnaryTest)
	if u == nil || u.Op != TsNot {
		return x
	}
	switch y := u.X.(type) {
	case *UnaryTest:
		switch y.Op {
		case TsEmpStr:
			y.Op = TsNempStr
			s.modified = true
			return y
		case TsNempStr:
			y.Op = TsEmpStr
			s.modified = true
			return y
		case TsNot:
			s.modified = true
			return y.X
		}
	case *BinaryTest:
		switch y.Op {
		case TsMatch:
			y.Op = TsNoMatch
			s.modified = true
			return y
		case TsNoMatch:
			y.Op = TsMatch
			s.modified = true
			return y
		}
	}
	return x
}
