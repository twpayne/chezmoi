// Copyright (c) 2016, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package syntax

import (
	"fmt"
	"io"
	"reflect"
)

func walkStmts(stmts []*Stmt, last []Comment, f func(Node) bool) {
	for _, s := range stmts {
		Walk(s, f)
	}
	for _, c := range last {
		Walk(&c, f)
	}
}

func walkWords(words []*Word, f func(Node) bool) {
	for _, w := range words {
		Walk(w, f)
	}
}

// Walk traverses a syntax tree in depth-first order: It starts by calling
// f(node); node must not be nil. If f returns true, Walk invokes f
// recursively for each of the non-nil children of node, followed by
// f(nil).
func Walk(node Node, f func(Node) bool) {
	if !f(node) {
		return
	}

	switch x := node.(type) {
	case *File:
		walkStmts(x.Stmts, x.Last, f)
	case *Comment:
	case *Stmt:
		for _, c := range x.Comments {
			if !x.End().After(c.Pos()) {
				defer Walk(&c, f)
				break
			}
			Walk(&c, f)
		}
		if x.Cmd != nil {
			Walk(x.Cmd, f)
		}
		for _, r := range x.Redirs {
			Walk(r, f)
		}
	case *Assign:
		if x.Name != nil {
			Walk(x.Name, f)
		}
		if x.Value != nil {
			Walk(x.Value, f)
		}
		if x.Index != nil {
			Walk(x.Index, f)
		}
		if x.Array != nil {
			Walk(x.Array, f)
		}
	case *Redirect:
		if x.N != nil {
			Walk(x.N, f)
		}
		Walk(x.Word, f)
		if x.Hdoc != nil {
			Walk(x.Hdoc, f)
		}
	case *CallExpr:
		for _, a := range x.Assigns {
			Walk(a, f)
		}
		walkWords(x.Args, f)
	case *Subshell:
		walkStmts(x.Stmts, x.Last, f)
	case *Block:
		walkStmts(x.Stmts, x.Last, f)
	case *IfClause:
		walkStmts(x.Cond, x.CondLast, f)
		walkStmts(x.Then, x.ThenLast, f)
		if x.Else != nil {
			Walk(x.Else, f)
		}
	case *WhileClause:
		walkStmts(x.Cond, x.CondLast, f)
		walkStmts(x.Do, x.DoLast, f)
	case *ForClause:
		Walk(x.Loop, f)
		walkStmts(x.Do, x.DoLast, f)
	case *WordIter:
		Walk(x.Name, f)
		walkWords(x.Items, f)
	case *CStyleLoop:
		if x.Init != nil {
			Walk(x.Init, f)
		}
		if x.Cond != nil {
			Walk(x.Cond, f)
		}
		if x.Post != nil {
			Walk(x.Post, f)
		}
	case *BinaryCmd:
		Walk(x.X, f)
		Walk(x.Y, f)
	case *FuncDecl:
		Walk(x.Name, f)
		Walk(x.Body, f)
	case *Word:
		for _, wp := range x.Parts {
			Walk(wp, f)
		}
	case *Lit:
	case *SglQuoted:
	case *DblQuoted:
		for _, wp := range x.Parts {
			Walk(wp, f)
		}
	case *CmdSubst:
		walkStmts(x.Stmts, x.Last, f)
	case *ParamExp:
		Walk(x.Param, f)
		if x.Index != nil {
			Walk(x.Index, f)
		}
		if x.Repl != nil {
			if x.Repl.Orig != nil {
				Walk(x.Repl.Orig, f)
			}
			if x.Repl.With != nil {
				Walk(x.Repl.With, f)
			}
		}
		if x.Exp != nil && x.Exp.Word != nil {
			Walk(x.Exp.Word, f)
		}
	case *ArithmExp:
		Walk(x.X, f)
	case *ArithmCmd:
		Walk(x.X, f)
	case *BinaryArithm:
		Walk(x.X, f)
		Walk(x.Y, f)
	case *BinaryTest:
		Walk(x.X, f)
		Walk(x.Y, f)
	case *UnaryArithm:
		Walk(x.X, f)
	case *UnaryTest:
		Walk(x.X, f)
	case *ParenArithm:
		Walk(x.X, f)
	case *ParenTest:
		Walk(x.X, f)
	case *CaseClause:
		Walk(x.Word, f)
		for _, ci := range x.Items {
			Walk(ci, f)
		}
		for _, c := range x.Last {
			Walk(&c, f)
		}
	case *CaseItem:
		for _, c := range x.Comments {
			if c.Pos().After(x.Pos()) {
				defer Walk(&c, f)
				break
			}
			Walk(&c, f)
		}
		walkWords(x.Patterns, f)
		walkStmts(x.Stmts, x.Last, f)
	case *TestClause:
		Walk(x.X, f)
	case *DeclClause:
		for _, a := range x.Args {
			Walk(a, f)
		}
	case *ArrayExpr:
		for _, el := range x.Elems {
			Walk(el, f)
		}
		for _, c := range x.Last {
			Walk(&c, f)
		}
	case *ArrayElem:
		for _, c := range x.Comments {
			if c.Pos().After(x.Pos()) {
				defer Walk(&c, f)
				break
			}
			Walk(&c, f)
		}
		if x.Index != nil {
			Walk(x.Index, f)
		}
		if x.Value != nil {
			Walk(x.Value, f)
		}
	case *ExtGlob:
		Walk(x.Pattern, f)
	case *ProcSubst:
		walkStmts(x.Stmts, x.Last, f)
	case *TimeClause:
		if x.Stmt != nil {
			Walk(x.Stmt, f)
		}
	case *CoprocClause:
		if x.Name != nil {
			Walk(x.Name, f)
		}
		Walk(x.Stmt, f)
	case *LetClause:
		for _, expr := range x.Exprs {
			Walk(expr, f)
		}
	case *TestDecl:
		Walk(x.Description, f)
		Walk(x.Body, f)
	default:
		panic(fmt.Sprintf("syntax.Walk: unexpected node type %T", x))
	}

	f(nil)
}

// DebugPrint prints the provided syntax tree, spanning multiple lines and with
// indentation. Can be useful to investigate the content of a syntax tree.
func DebugPrint(w io.Writer, node Node) error {
	p := debugPrinter{out: w}
	p.print(reflect.ValueOf(node))
	return p.err
}

type debugPrinter struct {
	out   io.Writer
	level int
	err   error
}

func (p *debugPrinter) printf(format string, args ...any) {
	_, err := fmt.Fprintf(p.out, format, args...)
	if err != nil && p.err == nil {
		p.err = err
	}
}

func (p *debugPrinter) newline() {
	p.printf("\n")
	for i := 0; i < p.level; i++ {
		p.printf(".  ")
	}
}

func (p *debugPrinter) print(x reflect.Value) {
	switch x.Kind() {
	case reflect.Interface:
		if x.IsNil() {
			p.printf("nil")
			return
		}
		p.print(x.Elem())
	case reflect.Ptr:
		if x.IsNil() {
			p.printf("nil")
			return
		}
		p.printf("*")
		p.print(x.Elem())
	case reflect.Slice:
		p.printf("%s (len = %d) {", x.Type(), x.Len())
		if x.Len() > 0 {
			p.level++
			p.newline()
			for i := 0; i < x.Len(); i++ {
				p.printf("%d: ", i)
				p.print(x.Index(i))
				if i == x.Len()-1 {
					p.level--
				}
				p.newline()
			}
		}
		p.printf("}")

	case reflect.Struct:
		if v, ok := x.Interface().(Pos); ok {
			p.printf("%v:%v", v.Line(), v.Col())
			return
		}
		t := x.Type()
		p.printf("%s {", t)
		p.level++
		p.newline()
		for i := 0; i < t.NumField(); i++ {
			p.printf("%s: ", t.Field(i).Name)
			p.print(x.Field(i))
			if i == x.NumField()-1 {
				p.level--
			}
			p.newline()
		}
		p.printf("}")
	default:
		p.printf("%#v", x.Interface())
	}
}
