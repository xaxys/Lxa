package generator

import (
	. "lxa/binchunk"
)

func (fi *funcInfo) toProto() *Prototype {
	var protos []*Prototype
	for _, fi := range fi.subFuncs {
		protos = append(protos, fi.toProto())
	}

	proto := &Prototype{
		LineDefined:     uint32(fi.line),
		LastLineDefined: uint32(fi.lastLine),
		NumParams:       byte(fi.numParams),
		MaxStackSize:    byte(fi.maxRegs),
		Code:            fi.insts,
		Constants:       fi.getConstants(),
		Upvalues:        fi.getUpvalues(),
		Protos:          protos,
		LineInfo:        fi.lineNums,
		LocVars:         fi.getLocVars(),
		UpvalueNames:    fi.getUpvalueNames(),
	}

	if fi.line == 0 {
		proto.LastLineDefined = 0
	}
	if proto.MaxStackSize < 2 {
		proto.MaxStackSize = 2 // todo
	}
	if fi.isVararg {
		proto.IsVararg = 1 // todo
	}

	return proto
}

func (fi *funcInfo) getConstants() []interface{} {
	consts := make([]interface{}, len(fi.constants))
	for k, idx := range fi.constants {
		consts[idx] = k
	}
	return consts
}

func (fi *funcInfo) getLocVars() []LocVar {
	locVars := make([]LocVar, len(fi.locVars))
	for i, locVar := range fi.locVars {
		locVars[i] = LocVar{
			VarName: locVar.name,
			StartPC: uint32(locVar.startPC),
			EndPC:   uint32(locVar.endPC),
		}
	}
	return locVars
}

func (fi *funcInfo) getUpvalues() []Upvalue {
	upvals := make([]Upvalue, len(fi.upvals))
	for _, uv := range fi.upvals {
		if uv.locVarSlot >= 0 { // instack
			upvals[uv.index] = Upvalue{Instack: 1, Idx: byte(uv.locVarSlot)}
		} else {
			upvals[uv.index] = Upvalue{Instack: 0, Idx: byte(uv.upvalIndex)}
		}
	}
	return upvals
}

func (fi *funcInfo) getUpvalueNames() []string {
	names := make([]string, len(fi.upvals))
	for name, uv := range fi.upvals {
		names[uv.index] = name
	}
	return names
}
