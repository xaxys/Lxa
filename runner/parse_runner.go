package runner

import (
	"fmt"
	"lxa/binchunk"
	"lxa/vm"
)

func ParseBinary(chunk []byte) {
	proto := binchunk.Undump(chunk)
	for i, code := range proto.Code {
		inst := vm.Instruction(code)
		switch inst.OpMode() {
		case vm.IABC:
			a, b, c := inst.ABC()
			fmt.Println("vm @", i, inst.OpName(), "A =", a, "B =", b, "C =", c)
		case vm.IABx:
			a, bx := inst.ABx()
			fmt.Println("vm @", i, inst.OpName(), "A =", a, "BX =", bx)
		case vm.IAsBx:
			a, sbx := inst.AsBx()
			fmt.Println("vm @", i, inst.OpName(), "A =", a, "SBX =", sbx)
		case vm.IAx:
			ax := inst.Ax()
			fmt.Println("vm @", i, inst.OpName(), "AX =", ax)
		}
	}
}
