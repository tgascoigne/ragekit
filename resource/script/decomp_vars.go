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

	var result Node
	v := block.popNode()

	switch v := v.(type) {
	case PtrNode:
		result = v.DeRef()
	default:
		result = DeRefExpr{v}
	}

	block.pushNode(result)
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
	var index int
	switch op := istr.Operands.(type) {
	case ImmediateIntOperands:
		index = op.Int()
	case *NoOperands:
		index = block.popNode().(Immediate).Value.(ImmediateIntOperands).Int()
	}

	isPtr := false
	var src Node
	switch istr.Operation {
	case OpGetGlobalP:
		isPtr = true
		fallthrough
	case OpGetGlobal:
		src = m.file.GlobalByIndex(index)

	case OpGetStaticP:
		isPtr = true
		fallthrough
	case OpGetStatic:
		src = m.file.Decls.VariableByIndex(index)

	case OpGetLocalP:
		isPtr = true
		fallthrough
	case OpGetLocal:
		src = block.VariableByIndex(index)

	case OpGetFieldP:
		isPtr = true
		fallthrough
	case OpGetField:
		struc := block.popNode()
		if s, ok := struc.(PtrNode); ok {
			struc = s.DeRef()
		}
		src = StructField{
			Struct: struc,
			Field: &Variable{
				Index:      index,
				Identifier: fmt.Sprintf("field_%v", index),
				Type:       UnknownType,
			},
		}

	case OpGetArrayP:
		isPtr = true
		fallthrough
	case OpGetArray:
		arr := block.popNode()
		if s, ok := arr.(PtrNode); ok {
			arr = s.DeRef()
		}
		src = ArrayIndex{
			Array: arr,
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
