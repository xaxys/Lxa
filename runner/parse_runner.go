package runner

import (
	"fmt"
	"lxa/binchunk"
	"lxa/vm"
)

func ParseBinary(chunk []byte) {
	proto := binchunk.Undump(chunk)
	fmt.Println("Source:", proto.Source)
	fmt.Println()

	params := uint8(proto.NumParams)
	isVararg := bool(proto.IsVararg == 0)
	slots := uint8(proto.MaxStackSize)
	upvals := len(proto.Upvalues) - 1
	locals := len(proto.LocVars)
	constants := len(proto.Constants) - 1
	functions := len(proto.Protos)
	fmt.Println(params, "params,", "Vararg[", isVararg, "],", slots, "slots,", upvals, "upvalue,", locals, "locals,", constants, "constants", functions, "functions")
	fmt.Println()

	for i, code := range proto.Code {
		inst := vm.Instruction(code)
		switch inst.OpMode() {
		case vm.IABC:
			a, b, c := inst.ABC()
			fmt.Println("vm @", i, inst.OpName(), "A =", a, "B =", b, "C =", c)
		case vm.IABx:
			a, bx := inst.ABx()
			if inst.Opcode() == vm.OP_LOADK {
				fmt.Println("vm @", i, inst.OpName(), "A =", a, "BX =", bx, "		;", proto.Constants[bx])
			} else {
				fmt.Println("vm @", i, inst.OpName(), "A =", a, "BX =", bx)
			}
		case vm.IAsBx:
			a, sbx := inst.AsBx()
			if inst.Opcode() == vm.OP_JMP {
				fmt.Println("vm @", i, inst.OpName(), "A =", a, "SBX =", sbx, "		;", "to", i+sbx+1)
			} else {
				fmt.Println("vm @", i, inst.OpName(), "A =", a, "SBX =", sbx)
			}
		case vm.IAx:
			ax := inst.Ax()
			fmt.Println("vm @", i, inst.OpName(), "AX =", ax)
		}
	}
	fmt.Println()

	fmt.Println(constants, "Constants:")
	for i := 1; i <= constants; i++ {
		fmt.Println("[", i, "]", proto.Constants[i])
	}
	fmt.Println()

	fmt.Println(locals, "Locals:")
	for i := 0; i < locals; i++ {
		fmt.Println("[", i, "]", proto.LocVars[i])
	}
	fmt.Println()

	fmt.Println(upvals, "Upvalues:")
	for i := 1; i <= upvals; i++ {
		fmt.Println("[", i, "]", proto.Upvalues[i])
	}
	fmt.Println()
}
