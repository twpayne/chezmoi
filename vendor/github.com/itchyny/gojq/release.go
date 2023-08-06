//go:build !gojq_debug
// +build !gojq_debug

package gojq

type codeinfo struct{}

func (c *compiler) appendCodeInfo(any) {}

func (c *compiler) deleteCodeInfo(string) {}

func (env *env) debugCodes() {}

func (env *env) debugState(int, bool) {}

func (env *env) debugForks(int, string) {}
