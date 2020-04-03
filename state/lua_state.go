package state

import (
	. "lxa/api"
)

type luaState struct {
	debug    bool
	registry *luaTable
	stack    *luaStack
	/* coroutine */
	coStatus int
	coCaller *luaState
	coChan   chan int
}

func New() LuaState {
	return NewState(false)
}

func NewState(debug bool) LuaState {
	ls := &luaState{debug: debug}

	registry := newLuaTable(8, 0)
	registry.put(LUA_RIDX_MAINTHREAD, ls)
	registry.put(LUA_RIDX_GLOBALS, newLuaTable(0, 20))

	ls.registry = registry
	ls.pushLuaStack(newLuaStack(LUA_MINSTACK, ls))
	return ls
}

func (self *luaState) isMainThread() bool {
	return self.registry.get(LUA_RIDX_MAINTHREAD) == self
}

func (self *luaState) pushLuaStack(stack *luaStack) {
	stack.prev = self.stack
	self.stack = stack
}

func (self *luaState) popLuaStack() {
	stack := self.stack
	self.stack = stack.prev
	stack.prev = nil
}
