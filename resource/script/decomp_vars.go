package script

import "fmt"

func (m *Machine) decompileSetP(block *BasicBlock, peek bool) {
	_ = block.instrs.nextInstruction()

	var value, dest Node
	if peek {
		value = block.popNode()
		dest = block.peekNode()
	} else {
		dest = block.popNode()
		value = block.popNode()
	}

	expectedtype := value.(DataTypeable).DataType()
	dest.(TypeInferrable).InferType(expectedtype)

	block.emitStatement(AssignStmt{dest, value})
}

func (m *Machine) decompileGetP(block *BasicBlock) {
	_ = block.instrs.nextInstruction()
	src := DeRefExpr{
		Node: block.popNode(),
	}

	block.pushNode(src)
}

func (m *Machine) decompileAssignment(block *BasicBlock) {
	istr := block.instrs.nextInstruction()
	op := istr.Operands.(ImmediateIntOperands)

	var dest Node
	switch istr.Operation {
	case OpSetGlobal:
		dest = m.file.GlobalByIndex(op.Int())
	case OpSetStatic:
		dest = m.file.Decls.VariableByIndex(op.Int())
	case OpSetLocal:
		dest = block.VariableByIndex(op.Int())
	case OpSetField:
		struc := block.popNode()
		dest = StructField{
			Struct: struc,
			Field: &Variable{
				Index:      op.Int(),
				Identifier: fmt.Sprintf("field_%v", op.Int()),
				Type:       UnknownType,
			},
		}
	case OpSetArray:
		dest = ArrayIndex{
			Array: block.popNode(),
			Index: block.popNode(),
		}
	default:
		fmt.Printf("dont know how to find var\n")
	}

	value := block.popNode()

	expectedtype := value.(DataTypeable).DataType()
	dest.(TypeInferrable).InferType(expectedtype)

	block.emitStatement(AssignStmt{dest, value})
}

func (m *Machine) decompileVarAccess(block *BasicBlock) {
	istr := block.instrs.nextInstruction()
	op := istr.Operands.(ImmediateIntOperands)

	isPtr := false
	var src Node
	switch istr.Operation {
	case OpGetGlobalP:
		isPtr = true
		fallthrough
	case OpGetGlobal:
		src = m.file.GlobalByIndex(op.Int())

	case OpGetStaticP:
		isPtr = true
		fallthrough
	case OpGetStatic:
		src = m.file.Decls.VariableByIndex(op.Int())

	case OpGetLocalP:
		isPtr = true
		fallthrough
	case OpGetLocal:
		src = block.VariableByIndex(op.Int())

	case OpGetFieldP:
		isPtr = true
		fallthrough
	case OpGetField:
		struc := block.popNode()
		src = StructField{
			Struct: struc,
			Field: &Variable{
				Index:      op.Int(),
				Identifier: fmt.Sprintf("field_%v", op.Int()),
				Type:       UnknownType,
			},
		}

	case OpGetArrayP:
		isPtr = true
		fallthrough
	case OpGetArray:
		src = ArrayIndex{
			Array: block.popNode(),
			Index: block.popNode(),
		}

	default:
		fmt.Printf("dont know how to find var\n")
	}

	if isPtr {
		src = PtrNode{
			Node: src,
		}
	}

	block.pushNode(src)
}
