package script

import (
	"fmt"
)

type Instruction struct {
	Address   uint32
	Operation uint8 /* one of Op* */
	Opcode    uint8
	Operands  Operands
}

func (i *Instruction) String() string {
	var mn, mnSuffix string

	if v, ok := OpMnemonic[i.Operation]; ok {
		mn = v
	} else {
		panic(fmt.Sprintf("no mnemonic for %v", i.Opcode))
	}

	if v, ok := OpSuffix[i.Opcode]; ok {
		mnSuffix = v
	}

	return fmt.Sprintf("%v%v %v", mn, mnSuffix, i.Operands.String())
}
