package script

import "fmt"

func (m *Machine) decompilePushImm(block *BasicBlock) {
	istr := block.instrs.nextInstruction()
	op := istr.Operands

	switch op := op.(type) {
	case ImmediateIntOperands:
		block.pushNode(Immediate{op.(Operands)})
	case *Immediate8x2Operands:
		block.pushNode(Immediate{&Immediate8Operands{op.Val0}})
		block.pushNode(Immediate{&Immediate8Operands{op.Val1}})
	case *Immediate8x3Operands:
		block.pushNode(Immediate{&Immediate8Operands{op.Val0}})
		block.pushNode(Immediate{&Immediate8Operands{op.Val1}})
		block.pushNode(Immediate{&Immediate8Operands{op.Val2}})
	}

}

func (m *Machine) decompilePushStr(block *BasicBlock) {
	_ = block.instrs.nextInstruction()

	strIndex := block.popNode().(Immediate).Value.(ImmediateIntOperands)
	value := m.script.StringTableEntry(strIndex.Int())
	block.pushNode(Immediate{
		Value: &StringOperands{value},
	})
}

func (m *Machine) decompileImplode(block *BasicBlock) {
	_ = block.instrs.nextInstruction()
	dest := block.popNode()

	length := block.popNode().(Immediate).Value.(ImmediateIntOperands).Int()
	elems := make(ArrayLiteral, length)
	for i := range elems {
		elems[length-i-1] = block.popNode()
	}

	expectedtype := elems[0].(DataTypeable).DataType()
	dest.(TypeInferrable).InferType(expectedtype)

	block.emitStatement(AssignStmt{dest, elems})
}

func (m *Machine) decompileExplode(block *BasicBlock) {
	_ = block.instrs.nextInstruction()
	src := block.popNode()
	length := block.popNode().(Immediate).Value.(ImmediateIntOperands).Int()

	if inferrable, ok := src.(TypeInferrable); ok {
		if length == 3 {
			inferrable.InferType(GetType("Vector3*"))
		} else {
			inferrable.InferType(PtrType{
				BaseType: ArrayType{
					BaseType: UnknownType,
					NumElems: length,
				}})
		}
	}

	fmt.Printf("src datatype is %v\n", src.(DataTypeable).DataType())

	// eww
	srcType := src.(DataTypeable).DataType().(ComplexType)

	exploded := srcType.Explode(src, length)
	for _, n := range exploded {
		block.pushNode(n)
	}

}
