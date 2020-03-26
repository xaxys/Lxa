package binchunk

import (
	"encoding/binary"
	"math"
)

type writer struct {
	data []byte
}

func (self *writer) writeByte(bytes byte) {
	self.data = append(self.data, bytes)
}

func (self *writer) writeString(s string) {
	self.writeBytes([]byte(s))
}

func (self *writer) writeBytes(bytes []byte) {
	self.data = append(self.data, bytes...)
}

func (self *writer) writeUint32(i uint32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, i)
	self.data = append(self.data, b...)
}

func (self *writer) writeUint64(i uint64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	self.data = append(self.data, b...)
}

func (self *writer) writeLuaInteger(i int64) {
	self.writeUint64(uint64(i))
}

func (self *writer) writeLuaNumber(f float64) {
	self.writeUint64(math.Float64bits(f))
}

func (self *writer) writeLuaString(s string) {
	size := len(s)
	if size == 0 {
		self.writeByte(0x00)
		return
	}
	if size < 254 {
		self.writeByte(byte(size + 1))
	} else {
		self.writeByte(0xFF)
		self.writeUint64(uint64(size + 1))
	}
	self.writeBytes([]byte(s))
}

func (self *writer) writeHeader() {
	self.writeString(LUA_SIGNATURE)
	self.writeByte(LUAC_VERSION)
	self.writeByte(LUAC_FORMAT)
	self.writeString(LUAC_DATA)
	self.writeByte(CINT_SIZE)
	self.writeByte(CSIZET_SIZE)
	self.writeByte(INSTRUCTION_SIZE)
	self.writeByte(LUA_INTEGER_SIZE)
	self.writeByte(LUA_NUMBER_SIZE)
	self.writeLuaInteger(LUAC_INT)
	self.writeLuaNumber(LUAC_NUM)
}

func (self *writer) writeProto(proto *Prototype) {
	self.writeLuaString(proto.Source)
	self.writeUint32(proto.LineDefined)
	self.writeUint32(proto.LastLineDefined)
	self.writeByte(proto.NumParams)
	self.writeByte(proto.IsVararg)
	self.writeByte(proto.MaxStackSize)
	self.writeCode(proto.Code)
	self.writeConstants(proto.Constants)
	self.writeUpvalues(proto.Upvalues)
	self.writeProtos(proto.Protos)
	self.writeLineInfo(proto.LineInfo)
	self.writeLocVars(proto.LocVars)
	self.writeUpvalueNames(proto.UpvalueNames)
}

func (self *writer) writeCode(code []uint32) {
	self.writeUint32(uint32(len(code)))
	for i := range code {
		self.writeUint32(code[i])
	}
}

func (self *writer) writeConstants(constants []interface{}) {
	self.writeUint32(uint32(len(constants)))
	for _, v := range constants {
		self.writeConstant(v)
	}
}

func (self *writer) writeConstant(constant interface{}) {
	switch cst := constant.(type) {
	case nil:
		self.writeByte(TAG_NIL)
	case bool:
		self.writeByte(TAG_BOOLEAN)
		if cst {
			self.writeByte(1)
		} else {
			self.writeByte(0)
		}
	case int64:
		self.writeByte(TAG_INTEGER)
		self.writeLuaInteger(cst)
	case float64:
		self.writeByte(TAG_NUMBER)
		self.writeLuaNumber(cst)
	case string:
		if len(cst) > 0 && len(cst) < 254 {
			self.writeByte(TAG_SHORT_STR)
		} else if len(cst) >= 254 {
			self.writeByte(TAG_LONG_STR)
		}
		self.writeLuaString(cst)
	default:
		panic("unsupported constant value type!")
	}
}

func (self *writer) writeUpvalues(upvalues []Upvalue) {
	self.writeUint32(uint32(len(upvalues)))
	for _, v := range upvalues {
		self.writeByte(v.Instack)
		self.writeByte(v.Idx)
	}
}

func (self *writer) writeProtos(protos []*Prototype) {
	self.writeUint32(uint32(len(protos)))
	for _, v := range protos {
		self.writeProto(v)
	}
}

func (self *writer) writeLineInfo(lineInfo []uint32) {
	self.writeUint32(uint32(len(lineInfo)))
	for _, v := range lineInfo {
		self.writeUint32(v)
	}
}

func (self *writer) writeLocVars(locVars []LocVar) {
	self.writeUint32(uint32(len(locVars)))
	for _, v := range locVars {
		self.writeLuaString(v.VarName)
		self.writeUint32(v.StartPC)
		self.writeUint32(v.EndPC)
	}
}

func (self *writer) writeUpvalueNames(names []string) {
	self.writeUint32(uint32(len(names)))
	for _, v := range names {
		self.writeLuaString(v)
	}
}
