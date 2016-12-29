package script

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

	// Is it a vec3 that we dont know about?
	if inferrable, ok := src.(TypeInferrable); ok && length == 3 {
		inferrable.InferType(GetType("Vector3"))
	}

	// eww
	srcType := src.(DataTypeable).DataType().(ComplexType)

	exploded := srcType.Explode(src, length)
	for _, n := range exploded {
		block.pushNode(n)
	}

}
