package gojq

import (
	"context"
	"math"
	"reflect"
	"sort"
)

func (env *env) execute(bc *Code, v any, vars ...any) Iter {
	env.codes = bc.codes
	env.codeinfos = bc.codeinfos
	env.push(v)
	for i := len(vars) - 1; i >= 0; i-- {
		env.push(vars[i])
	}
	env.debugCodes()
	return env
}

func (env *env) Next() (any, bool) {
	var err error
	pc, callpc, index := env.pc, len(env.codes)-1, -1
	backtrack, hasCtx := env.backtrack, env.ctx != context.Background()
	defer func() { env.pc, env.backtrack = pc, true }()
loop:
	for ; pc < len(env.codes); pc++ {
		env.debugState(pc, backtrack)
		code := env.codes[pc]
		if hasCtx {
			select {
			case <-env.ctx.Done():
				pc, env.forks = len(env.codes), nil
				return env.ctx.Err(), true
			default:
			}
		}
		switch code.op {
		case opnop:
			// nop
		case oppush:
			env.push(code.v)
		case oppop:
			env.pop()
		case opdup:
			v := env.pop()
			env.push(v)
			env.push(v)
		case opconst:
			env.pop()
			env.push(code.v)
		case opload:
			env.push(env.values[env.index(code.v.([2]int))])
		case opstore:
			env.values[env.index(code.v.([2]int))] = env.pop()
		case opobject:
			if backtrack {
				break loop
			}
			n := code.v.(int)
			m := make(map[string]any, n)
			for i := 0; i < n; i++ {
				v, k := env.pop(), env.pop()
				s, ok := k.(string)
				if !ok {
					err = &objectKeyNotStringError{k}
					break loop
				}
				m[s] = v
			}
			env.push(m)
		case opappend:
			i := env.index(code.v.([2]int))
			env.values[i] = append(env.values[i].([]any), env.pop())
		case opfork:
			if backtrack {
				if err != nil {
					break loop
				}
				pc, backtrack = code.v.(int), false
				goto loop
			}
			env.pushfork(pc)
		case opforktrybegin:
			if backtrack {
				if err == nil {
					break loop
				}
				switch er := err.(type) {
				case *tryEndError:
					err = er.err
					break loop
				case *breakError:
					break loop
				case ValueError:
					if er, ok := er.(*exitCodeError); ok && er.halt {
						break loop
					}
					if v := er.Value(); v != nil {
						env.pop()
						env.push(v)
					} else {
						err = nil
						break loop
					}
				default:
					env.pop()
					env.push(err.Error())
				}
				pc, backtrack, err = code.v.(int), false, nil
				goto loop
			}
			env.pushfork(pc)
		case opforktryend:
			if backtrack {
				if err != nil {
					err = &tryEndError{err}
				}
				break loop
			}
			env.pushfork(pc)
		case opforkalt:
			if backtrack {
				if err == nil {
					break loop
				}
				pc, backtrack, err = code.v.(int), false, nil
				goto loop
			}
			env.pushfork(pc)
		case opforklabel:
			if backtrack {
				label := env.pop()
				if e, ok := err.(*breakError); ok && e.v == label {
					err = nil
				}
				break loop
			}
			env.push(env.label)
			env.pushfork(pc)
			env.pop()
			env.values[env.index(code.v.([2]int))] = env.label
			env.label++
		case opbacktrack:
			break loop
		case opjump:
			pc = code.v.(int)
			goto loop
		case opjumpifnot:
			if v := env.pop(); v == nil || v == false {
				pc = code.v.(int)
				goto loop
			}
		case opindex, opindexarray:
			if backtrack {
				break loop
			}
			p, v := code.v, env.pop()
			if code.op == opindexarray && v != nil {
				if _, ok := v.([]any); !ok {
					err = &expectedArrayError{v}
					break loop
				}
			}
			w := funcIndex2(nil, v, p)
			if e, ok := w.(error); ok {
				err = e
				break loop
			}
			env.push(w)
			if !env.paths.empty() && env.expdepth == 0 {
				if !env.pathIntact(v) {
					err = &invalidPathError{v}
					break loop
				}
				env.paths.push(pathValue{path: p, value: w})
			}
		case opcall:
			if backtrack {
				break loop
			}
			switch v := code.v.(type) {
			case int:
				pc, callpc, index = v, pc, env.scopes.index
				goto loop
			case [3]any:
				argcnt := v[1].(int)
				x, args := env.pop(), env.args[:argcnt]
				for i := 0; i < argcnt; i++ {
					args[i] = env.pop()
				}
				w := v[0].(func(any, []any) any)(x, args)
				if e, ok := w.(error); ok {
					if er, ok := e.(*exitCodeError); !ok || er.value != nil || er.halt {
						err = e
					}
					break loop
				}
				env.push(w)
				if !env.paths.empty() && env.expdepth == 0 {
					switch v[2].(string) {
					case "_index":
						if x = args[0]; !env.pathIntact(x) {
							err = &invalidPathError{x}
							break loop
						}
						env.paths.push(pathValue{path: args[1], value: w})
					case "_slice":
						if x = args[0]; !env.pathIntact(x) {
							err = &invalidPathError{x}
							break loop
						}
						env.paths.push(pathValue{
							path:  map[string]any{"start": args[2], "end": args[1]},
							value: w,
						})
					case "getpath":
						if !env.pathIntact(x) {
							err = &invalidPathError{x}
							break loop
						}
						for _, p := range args[0].([]any) {
							env.paths.push(pathValue{path: p, value: w})
						}
					}
				}
			default:
				panic(v)
			}
		case opcallrec:
			pc, callpc, index = code.v.(int), -1, env.scopes.index
			goto loop
		case oppushpc:
			env.push([2]int{code.v.(int), env.scopes.index})
		case opcallpc:
			xs := env.pop().([2]int)
			pc, callpc, index = xs[0], pc, xs[1]
			goto loop
		case opscope:
			xs := code.v.([3]int)
			var saveindex, outerindex int
			if index == env.scopes.index {
				if callpc >= 0 {
					saveindex = index
				} else {
					callpc, saveindex = env.popscope()
				}
			} else {
				saveindex, _ = env.scopes.save()
				env.scopes.index = index
			}
			if outerindex = index; outerindex >= 0 {
				if s := env.scopes.data[outerindex].value; s.id == xs[0] {
					outerindex = s.outerindex
				}
			}
			env.scopes.push(scope{xs[0], env.offset, callpc, saveindex, outerindex})
			env.offset += xs[1]
			if env.offset > len(env.values) {
				vs := make([]any, env.offset*2)
				copy(vs, env.values)
				env.values = vs
			}
		case opret:
			if backtrack {
				break loop
			}
			pc, env.scopes.index = env.popscope()
			if env.scopes.empty() {
				return env.pop(), true
			}
		case opiter:
			if err != nil {
				break loop
			}
			backtrack = false
			var xs []pathValue
			switch v := env.pop().(type) {
			case []pathValue:
				xs = v
			case []any:
				if !env.paths.empty() && env.expdepth == 0 && !env.pathIntact(v) {
					err = &invalidPathIterError{v}
					break loop
				}
				if len(v) == 0 {
					break loop
				}
				xs = make([]pathValue, len(v))
				for i, v := range v {
					xs[i] = pathValue{path: i, value: v}
				}
			case map[string]any:
				if !env.paths.empty() && env.expdepth == 0 && !env.pathIntact(v) {
					err = &invalidPathIterError{v}
					break loop
				}
				if len(v) == 0 {
					break loop
				}
				xs = make([]pathValue, len(v))
				var i int
				for k, v := range v {
					xs[i] = pathValue{path: k, value: v}
					i++
				}
				sort.Slice(xs, func(i, j int) bool {
					return xs[i].path.(string) < xs[j].path.(string)
				})
			case Iter:
				if w, ok := v.Next(); ok {
					env.push(v)
					env.pushfork(pc)
					env.pop()
					if e, ok := w.(error); ok {
						err = e
						break loop
					}
					env.push(w)
					continue
				}
				break loop
			default:
				err = &iteratorError{v}
				env.push(emptyIter{})
				break loop
			}
			if len(xs) > 1 {
				env.push(xs[1:])
				env.pushfork(pc)
				env.pop()
			}
			env.push(xs[0].value)
			if !env.paths.empty() && env.expdepth == 0 {
				env.paths.push(xs[0])
			}
		case opexpbegin:
			env.expdepth++
		case opexpend:
			env.expdepth--
		case oppathbegin:
			env.paths.push(env.expdepth)
			env.paths.push(pathValue{value: env.stack.top()})
			env.expdepth = 0
		case oppathend:
			if backtrack {
				break loop
			}
			env.pop()
			if v := env.pop(); !env.pathIntact(v) {
				err = &invalidPathError{v}
				break loop
			}
			env.push(env.poppaths())
			env.expdepth = env.paths.pop().(int)
		default:
			panic(code.op)
		}
	}
	if len(env.forks) > 0 {
		pc, backtrack = env.popfork(), true
		goto loop
	}
	if err != nil {
		return err, true
	}
	return nil, false
}

func (env *env) push(v any) {
	env.stack.push(v)
}

func (env *env) pop() any {
	return env.stack.pop()
}

func (env *env) popscope() (int, int) {
	free := env.scopes.index > env.scopes.limit
	s := env.scopes.pop()
	if free {
		env.offset = s.offset
	}
	return s.pc, s.saveindex
}

func (env *env) pushfork(pc int) {
	f := fork{pc: pc, expdepth: env.expdepth}
	f.stackindex, f.stacklimit = env.stack.save()
	f.scopeindex, f.scopelimit = env.scopes.save()
	f.pathindex, f.pathlimit = env.paths.save()
	env.forks = append(env.forks, f)
	env.debugForks(pc, ">>>")
}

func (env *env) popfork() int {
	f := env.forks[len(env.forks)-1]
	env.debugForks(f.pc, "<<<")
	env.forks, env.expdepth = env.forks[:len(env.forks)-1], f.expdepth
	env.stack.restore(f.stackindex, f.stacklimit)
	env.scopes.restore(f.scopeindex, f.scopelimit)
	env.paths.restore(f.pathindex, f.pathlimit)
	return f.pc
}

func (env *env) index(v [2]int) int {
	for id, i := v[0], env.scopes.index; i >= 0; {
		s := env.scopes.data[i].value
		if s.id == id {
			return s.offset + v[1]
		}
		i = s.outerindex
	}
	panic("env.index")
}

type pathValue struct {
	path, value any
}

func (env *env) pathIntact(v any) bool {
	w := env.paths.top().(pathValue).value
	switch v := v.(type) {
	case []any, map[string]any:
		switch w.(type) {
		case []any, map[string]any:
			v, w := reflect.ValueOf(v), reflect.ValueOf(w)
			return v.Pointer() == w.Pointer() && v.Len() == w.Len()
		}
	case float64:
		if w, ok := w.(float64); ok {
			return v == w || math.IsNaN(v) && math.IsNaN(w)
		}
	}
	return v == w
}

func (env *env) poppaths() []any {
	xs := []any{}
	for {
		p := env.paths.pop().(pathValue)
		if p.path == nil {
			break
		}
		xs = append(xs, p.path)
	}
	for i, j := 0, len(xs)-1; i < j; i, j = i+1, j-1 {
		xs[i], xs[j] = xs[j], xs[i]
	}
	return xs
}
