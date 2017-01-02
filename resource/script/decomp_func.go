package script

func (m *Machine) decompileCall(block *BasicBlock) {
	istr := block.instrs.nextInstruction()

	var node Node
	var targetFn *Function
	inferArgIdentifiers := false

	switch op := istr.Operands.(type) {
	case *CallOperands:
		targetFn = m.file.FunctionByAddress(op.Val)
	case *CallNOperands:
		targetFn, inferArgIdentifiers = m.file.FunctionForNative(m.script.HashTable, op)
		// FunctionForNative's second parameter is true if the native spec had to be generated
		// (and the arg identifiers are useless)
		inferArgIdentifiers = !inferArgIdentifiers
	}

	args := make([]Node, targetFn.In.Size())
	for i := range args {
		argIdx := targetFn.In.Size() - i - 1

		argValue := block.popNode()
		argVariable := targetFn.In.Vars[argIdx]

		if inferrable, ok := argValue.(TypeInferrable); ok {
			expectedType := argVariable.DataType()
			inferrable.InferType(expectedType)
		}

		if v, valIsVariable := argValue.(*Variable); inferArgIdentifiers && valIsVariable {
			v.Identifier = argVariable.Identifier
		}

		args[argIdx] = argValue
	}

	node = FunctionCall{
		Fn:   targetFn,
		Args: args,
	}

	outSize := targetFn.Out.DataType().StackSize()
	if outSize > 0 {
		tempDecl := &VariableDeclaration{
			Variable: &Variable{
				Identifier: m.genTempIdentifier(),
				Type:       targetFn.Out.DataType(),
			},
			Value: node,
		}

		resultRef := tempDecl.Variable.Reference()
		if outSize > 1 {
			outType := targetFn.Out.DataType().(ComplexType)
			exploded := outType.Explode(resultRef, outType.StackSize())
			for _, n := range exploded {
				block.pushNode(n)
			}
		} else {
			block.pushNode(resultRef)
		}

		node = tempDecl
	}

	block.emitStatement(node)
}

func (m *Machine) decompileReturn(block *BasicBlock) {
	istr := block.instrs.nextInstruction()
	op := istr.Operands.(*RetOperands)

	retVar := block.ParentFunc.Out

	if op.NumReturnVals == 0 {
		retVar.InferType(VoidType)
		block.emitStatement(ReturnStmt{nil})
	} else {
		retVal := block.peekNode()
		retVar.InferType(retVal.(DataTypeable).DataType())

		if v, ok := retVal.(*Variable); ok {
			retVal = v.Reference()
		}

		block.emitStatement(ReturnStmt{retVal})
	}
}
